package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"ling-app/api/internal/config"
	"ling-app/api/internal/db"
	"ling-app/api/internal/models"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	portalsession "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/webhook"
	"gorm.io/gorm"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrInvalidWebhook       = errors.New("invalid webhook signature")
)

type StripeService struct {
	config         *config.Config
	db             *db.DB
	creditsService *CreditsService
}

func NewStripeService(cfg *config.Config, database *db.DB, creditsService *CreditsService) *StripeService {
	stripe.Key = cfg.StripeSecretKey
	return &StripeService{
		config:         cfg,
		db:             database,
		creditsService: creditsService,
	}
}

// GetSubscription retrieves a user's subscription record
func (s *StripeService) GetSubscription(userID uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	err := s.db.Where("user_id = ?", userID).First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	return &sub, nil
}

// GetOrCreateSubscription returns existing or creates a new free subscription
func (s *StripeService) GetOrCreateSubscription(userID uuid.UUID, email, name string) (*models.Subscription, error) {
	sub, err := s.GetSubscription(userID)
	if err == nil {
		return sub, nil
	}
	if !errors.Is(err, ErrSubscriptionNotFound) {
		return nil, err
	}

	// Create Stripe customer
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}
	cust, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("create stripe customer: %w", err)
	}

	// Create subscription record with free tier
	sub = &models.Subscription{
		UserID:           userID,
		StripeCustomerID: cust.ID,
		Tier:             models.TierFree,
		Status:           "active",
	}
	if err := s.db.Create(sub).Error; err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	return sub, nil
}

// CreateCheckoutSession creates a Stripe checkout URL for upgrading
func (s *StripeService) CreateCheckoutSession(userID uuid.UUID, email, name string, tier models.SubscriptionTier) (string, error) {
	sub, err := s.GetOrCreateSubscription(userID, email, name)
	if err != nil {
		return "", err
	}

	var priceID string
	switch tier {
	case models.TierBasic:
		priceID = s.config.StripePriceBasic
	case models.TierPro:
		priceID = s.config.StripePricePro
	default:
		return "", fmt.Errorf("invalid tier: %s", tier)
	}

	if priceID == "" {
		return "", fmt.Errorf("price not configured for tier: %s", tier)
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(sub.StripeCustomerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(s.config.StripeSuccessURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(s.config.StripeCancelURL),
	}
	params.AddMetadata("user_id", userID.String())
	params.AddMetadata("tier", string(tier))

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("create checkout session: %w", err)
	}

	return sess.URL, nil
}

// CreatePortalSession creates a billing portal URL
func (s *StripeService) CreatePortalSession(userID uuid.UUID) (string, error) {
	sub, err := s.GetSubscription(userID)
	if err != nil {
		return "", err
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(sub.StripeCustomerID),
		ReturnURL: stripe.String(s.config.FrontendURL + "/settings"),
	}

	sess, err := portalsession.New(params)
	if err != nil {
		return "", fmt.Errorf("create portal session: %w", err)
	}

	return sess.URL, nil
}

// HandleWebhook processes a Stripe webhook event
// payload is the raw request body, signature is the Stripe-Signature header
func (s *StripeService) HandleWebhook(payload []byte, signature string) error {
	if s.config.StripeWebhookSecret == "" {
		return fmt.Errorf("%w: webhook secret not configured", ErrInvalidWebhook)
	}

	event, err := webhook.ConstructEventWithOptions(payload, signature, s.config.StripeWebhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		return fmt.Errorf("%w: %v", ErrInvalidWebhook, err)
	}

	log.Printf("Stripe webhook: %s", event.Type)

	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutCompleted(event.Data.Raw)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(event.Data.Raw)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(event.Data.Raw)
	case "invoice.paid":
		return s.handleInvoicePaid(event.Data.Raw)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(event.Data.Raw)
	default:
		log.Printf("Unhandled webhook event: %s", event.Type)
	}

	return nil
}

func (s *StripeService) handleCheckoutCompleted(data json.RawMessage) error {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(data, &sess); err != nil {
		return fmt.Errorf("unmarshal checkout session: %w", err)
	}

	userIDStr := sess.Metadata["user_id"]
	if userIDStr == "" {
		return fmt.Errorf("user_id not in metadata")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("parse user_id: %w", err)
	}

	tierStr := sess.Metadata["tier"]
	if tierStr == "" {
		return fmt.Errorf("tier not in metadata")
	}
	tier := models.SubscriptionTier(tierStr)

	return s.db.Transaction(func(tx *gorm.DB) error {
		var sub models.Subscription
		if err := tx.Where("user_id = ?", userID).First(&sub).Error; err != nil {
			return fmt.Errorf("find subscription: %w", err)
		}

		// Get subscription ID from the checkout session
		var subID string
		if sess.Subscription != nil {
			subID = sess.Subscription.ID
		}

		sub.StripeSubscriptionID = &subID
		sub.Tier = tier
		sub.Status = "active"

		if err := tx.Save(&sub).Error; err != nil {
			return fmt.Errorf("update subscription: %w", err)
		}

		// Update credit allowance
		if err := s.creditsService.UpdateAllowance(userID, tier); err != nil {
			log.Printf("Failed to update allowance: %v", err)
		}

		// Grant new tier credits immediately
		newAllowance := models.TierCredits[tier]
		if err := s.creditsService.AddCredits(userID, newAllowance, fmt.Sprintf("Upgraded to %s", tier)); err != nil {
			log.Printf("Failed to add upgrade credits: %v", err)
		}

		return nil
	})
}

func (s *StripeService) handleSubscriptionUpdated(data json.RawMessage) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(data, &stripeSub); err != nil {
		return fmt.Errorf("unmarshal subscription: %w", err)
	}

	var sub models.Subscription
	if err := s.db.Where("stripe_subscription_id = ?", stripeSub.ID).First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Subscription not found for stripe_subscription_id: %s", stripeSub.ID)
			return nil // Not an error - might be from another system
		}
		return fmt.Errorf("find subscription: %w", err)
	}

	sub.Status = string(stripeSub.Status)
	sub.CancelAtPeriodEnd = stripeSub.CancelAtPeriodEnd

	// Determine tier from price
	if len(stripeSub.Items.Data) > 0 {
		priceID := stripeSub.Items.Data[0].Price.ID
		sub.StripePriceID = &priceID

		switch priceID {
		case s.config.StripePriceBasic:
			sub.Tier = models.TierBasic
		case s.config.StripePricePro:
			sub.Tier = models.TierPro
		}

		if err := s.creditsService.UpdateAllowance(sub.UserID, sub.Tier); err != nil {
			log.Printf("Failed to update allowance: %v", err)
		}
	}

	return s.db.Save(&sub).Error
}

func (s *StripeService) handleSubscriptionDeleted(data json.RawMessage) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(data, &stripeSub); err != nil {
		return fmt.Errorf("unmarshal subscription: %w", err)
	}

	var sub models.Subscription
	if err := s.db.Where("stripe_subscription_id = ?", stripeSub.ID).First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("find subscription: %w", err)
	}

	// Downgrade to free
	sub.Tier = models.TierFree
	sub.Status = "canceled"
	sub.StripeSubscriptionID = nil
	sub.StripePriceID = nil

	if err := s.db.Save(&sub).Error; err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}

	return s.creditsService.UpdateAllowance(sub.UserID, models.TierFree)
}

func (s *StripeService) handleInvoicePaid(data json.RawMessage) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(data, &invoice); err != nil {
		return fmt.Errorf("unmarshal invoice: %w", err)
	}

	// Get subscription ID from parent (stripe-go v82+ structure)
	if invoice.Parent == nil || invoice.Parent.SubscriptionDetails == nil || invoice.Parent.SubscriptionDetails.Subscription == nil {
		return nil
	}
	subID := invoice.Parent.SubscriptionDetails.Subscription.ID

	var sub models.Subscription
	if err := s.db.Where("stripe_subscription_id = ?", subID).First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("find subscription: %w", err)
	}

	return s.creditsService.RefreshMonthlyCredits(sub.UserID)
}

func (s *StripeService) handleInvoicePaymentFailed(data json.RawMessage) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(data, &invoice); err != nil {
		return fmt.Errorf("unmarshal invoice: %w", err)
	}

	// Get subscription ID from parent (stripe-go v82+ structure)
	if invoice.Parent == nil || invoice.Parent.SubscriptionDetails == nil || invoice.Parent.SubscriptionDetails.Subscription == nil {
		return nil
	}
	subID := invoice.Parent.SubscriptionDetails.Subscription.ID

	return s.db.Model(&models.Subscription{}).
		Where("stripe_subscription_id = ?", subID).
		Update("status", "past_due").Error
}
