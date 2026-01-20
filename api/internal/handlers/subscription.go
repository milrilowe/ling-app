package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"ling-app/api/internal/middleware"
	"ling-app/api/internal/models"
	"ling-app/api/internal/services"
)

type SubscriptionHandler struct {
	stripeService  services.StripeProcessor
	creditsService services.CreditsManager
}

func NewSubscriptionHandler(stripeService services.StripeProcessor, creditsService services.CreditsManager) *SubscriptionHandler {
	return &SubscriptionHandler{
		stripeService:  stripeService,
		creditsService: creditsService,
	}
}

// GetSubscriptionStatus returns the user's subscription and credits
// GET /api/subscription
func (h *SubscriptionHandler) GetSubscriptionStatus(c *gin.Context) {
	user := middleware.MustGetUser(c)

	sub, err := h.stripeService.GetOrCreateSubscription(user.ID, user.Email, user.Name)
	if err != nil {
		log.Printf("GetSubscriptionStatus error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription"})
		return
	}

	credits, err := h.creditsService.GetCredits(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get credits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": sub,
		"credits":      credits,
	})
}

type CheckoutRequest struct {
	Tier string `json:"tier" binding:"required,oneof=basic pro"`
}

// CreateCheckoutSession creates a Stripe checkout session
// POST /api/subscription/checkout
func (h *SubscriptionHandler) CreateCheckoutSession(c *gin.Context) {
	user := middleware.MustGetUser(c)

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tier must be 'basic' or 'pro'"})
		return
	}

	url, err := h.stripeService.CreateCheckoutSession(user.ID, user.Email, user.Name, models.SubscriptionTier(req.Tier))
	if err != nil {
		log.Printf("CreateCheckoutSession error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// CreatePortalSession creates a Stripe billing portal session
// POST /api/subscription/portal
func (h *SubscriptionHandler) CreatePortalSession(c *gin.Context) {
	user := middleware.MustGetUser(c)

	url, err := h.stripeService.CreatePortalSession(user.ID)
	if err != nil {
		if errors.Is(err, services.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No subscription found"})
			return
		}
		log.Printf("CreatePortalSession error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create portal session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// GetCreditsBalance returns the user's credit balance
// GET /api/credits
func (h *SubscriptionHandler) GetCreditsBalance(c *gin.Context) {
	user := middleware.MustGetUser(c)

	credits, err := h.creditsService.GetCredits(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get credits"})
		return
	}

	c.JSON(http.StatusOK, credits)
}

// GetCreditHistory returns the user's credit transaction history
// GET /api/credits/history
func (h *SubscriptionHandler) GetCreditHistory(c *gin.Context) {
	user := middleware.MustGetUser(c)

	transactions, err := h.creditsService.GetTransactionHistory(user.ID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// HandleStripeWebhook processes Stripe webhook events
// POST /api/webhooks/stripe
func (h *SubscriptionHandler) HandleStripeWebhook(c *gin.Context) {
	// Read raw body - must be done before any parsing
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Webhook: failed to read body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		log.Printf("Webhook: missing Stripe-Signature header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Stripe-Signature header"})
		return
	}

	if err := h.stripeService.HandleWebhook(payload, signature); err != nil {
		if errors.Is(err, services.ErrInvalidWebhook) {
			log.Printf("Webhook: invalid signature: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook signature"})
			return
		}
		log.Printf("Webhook: processing error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
