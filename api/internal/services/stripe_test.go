package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"ling-app/api/internal/config"
	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
	"ling-app/api/internal/repository/mocks"
)

// mockCreditsService is a mock for credits service in stripe tests.
type mockCreditsService struct {
	mock.Mock
}

func (m *mockCreditsService) UpdateAllowance(userID uuid.UUID, tier models.SubscriptionTier) error {
	args := m.Called(userID, tier)
	return args.Error(0)
}

func (m *mockCreditsService) AddCredits(userID uuid.UUID, amount int, description string) error {
	args := m.Called(userID, amount, description)
	return args.Error(0)
}

func (m *mockCreditsService) RefreshMonthlyCredits(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

// NewStripeServiceForTest creates a StripeService with injected dependencies for testing.
// Note: This cannot test methods that require Stripe API calls.
func NewStripeServiceForTest(
	cfg *config.Config,
	exec repository.Executor,
	txRunner TxRunner,
	subRepo repository.SubscriptionRepository,
	creditsService *CreditsService,
) *StripeService {
	return &StripeService{
		config:         cfg,
		db:             nil,
		exec:           exec,
		txRunner:       txRunner,
		subRepo:        subRepo,
		creditsService: creditsService,
	}
}

func TestStripeService_GetSubscription(t *testing.T) {
	userID := uuid.New()

	t.Run("returns subscription when found", func(t *testing.T) {
		subRepo := new(mocks.MockSubscriptionRepository)
		cfg := &config.Config{}

		expectedSub := &models.Subscription{
			UserID: userID,
			Tier:   models.TierPro,
			Status: "active",
		}

		subRepo.On("FindByUserID", mock.Anything, userID).Return(expectedSub, nil)

		service := NewStripeServiceForTest(cfg, nil, nil, subRepo, nil)
		sub, err := service.GetSubscription(userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedSub, sub)
		subRepo.AssertExpectations(t)
	})

	t.Run("returns ErrSubscriptionNotFound when not found", func(t *testing.T) {
		subRepo := new(mocks.MockSubscriptionRepository)
		cfg := &config.Config{}

		subRepo.On("FindByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

		service := NewStripeServiceForTest(cfg, nil, nil, subRepo, nil)
		sub, err := service.GetSubscription(userID)

		assert.ErrorIs(t, err, ErrSubscriptionNotFound)
		assert.Nil(t, sub)
	})

	t.Run("wraps other errors", func(t *testing.T) {
		subRepo := new(mocks.MockSubscriptionRepository)
		cfg := &config.Config{}

		dbError := errors.New("database error")
		subRepo.On("FindByUserID", mock.Anything, userID).Return(nil, dbError)

		service := NewStripeServiceForTest(cfg, nil, nil, subRepo, nil)
		sub, err := service.GetSubscription(userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get subscription")
		assert.Nil(t, sub)
	})
}

func TestStripeService_handleSubscriptionUpdated(t *testing.T) {
	userID := uuid.New()
	stripeSubID := "sub_test123"

	t.Run("updates subscription status and tier", func(t *testing.T) {
		subRepo := new(mocks.MockSubscriptionRepository)
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		cfg := &config.Config{
			StripePriceBasic: "price_basic",
			StripePricePro:   "price_pro",
		}

		existingSub := &models.Subscription{
			UserID:               userID,
			StripeSubscriptionID: &stripeSubID,
			Tier:                 models.TierBasic,
			Status:               "active",
		}

		// Webhook data simulating a subscription update to Pro tier
		webhookData := map[string]interface{}{
			"id":                   stripeSubID,
			"status":               "active",
			"cancel_at_period_end": false,
			"items": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"price": map[string]interface{}{
							"id": "price_pro",
						},
					},
				},
			},
		}
		data, _ := json.Marshal(webhookData)

		subRepo.On("FindByStripeSubscriptionID", mock.Anything, stripeSubID).Return(existingSub, nil)
		subRepo.On("Save", mock.Anything, mock.MatchedBy(func(s *models.Subscription) bool {
			return s.Tier == models.TierPro && s.Status == "active"
		})).Return(nil)

		// Credits service mock - UpdateAllowance is called
		creditsService := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)

		// Mock the UpdateAllowance dependency
		creditsRepo.On("UpdateAllowance", mock.Anything, userID, models.TierCredits[models.TierPro]).Return(nil)

		service := NewStripeServiceForTest(cfg, nil, nil, subRepo, creditsService)
		err := service.handleSubscriptionUpdated(data)

		assert.NoError(t, err)
		subRepo.AssertExpectations(t)
	})

	t.Run("returns nil when subscription not found", func(t *testing.T) {
		subRepo := new(mocks.MockSubscriptionRepository)
		cfg := &config.Config{}

		webhookData := map[string]interface{}{
			"id":     stripeSubID,
			"status": "active",
		}
		data, _ := json.Marshal(webhookData)

		subRepo.On("FindByStripeSubscriptionID", mock.Anything, stripeSubID).Return(nil, repository.ErrNotFound)

		service := NewStripeServiceForTest(cfg, nil, nil, subRepo, nil)
		err := service.handleSubscriptionUpdated(data)

		assert.NoError(t, err) // Not an error for webhooks
	})
}

func TestStripeService_handleSubscriptionDeleted(t *testing.T) {
	userID := uuid.New()
	stripeSubID := "sub_test123"

	t.Run("downgrades to free tier", func(t *testing.T) {
		subRepo := new(mocks.MockSubscriptionRepository)
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)
		cfg := &config.Config{}

		existingSub := &models.Subscription{
			UserID:               userID,
			StripeSubscriptionID: &stripeSubID,
			Tier:                 models.TierPro,
			Status:               "active",
		}

		webhookData := map[string]interface{}{
			"id": stripeSubID,
		}
		data, _ := json.Marshal(webhookData)

		subRepo.On("FindByStripeSubscriptionID", mock.Anything, stripeSubID).Return(existingSub, nil)
		subRepo.On("Save", mock.Anything, mock.MatchedBy(func(s *models.Subscription) bool {
			return s.Tier == models.TierFree && s.Status == "canceled" && s.StripeSubscriptionID == nil
		})).Return(nil)

		creditsService := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		creditsRepo.On("UpdateAllowance", mock.Anything, userID, models.TierCredits[models.TierFree]).Return(nil)

		service := NewStripeServiceForTest(cfg, nil, nil, subRepo, creditsService)
		err := service.handleSubscriptionDeleted(data)

		assert.NoError(t, err)
		subRepo.AssertExpectations(t)
	})

	t.Run("returns nil when subscription not found", func(t *testing.T) {
		subRepo := new(mocks.MockSubscriptionRepository)
		cfg := &config.Config{}

		webhookData := map[string]interface{}{
			"id": stripeSubID,
		}
		data, _ := json.Marshal(webhookData)

		subRepo.On("FindByStripeSubscriptionID", mock.Anything, stripeSubID).Return(nil, repository.ErrNotFound)

		service := NewStripeServiceForTest(cfg, nil, nil, subRepo, nil)
		err := service.handleSubscriptionDeleted(data)

		assert.NoError(t, err)
	})
}

func TestStripeService_handleInvoicePaid(t *testing.T) {
	userID := uuid.New()
	stripeSubID := "sub_test123"

	t.Run("refreshes monthly credits", func(t *testing.T) {
		subRepo := new(mocks.MockSubscriptionRepository)
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)
		cfg := &config.Config{}

		existingSub := &models.Subscription{
			UserID:               userID,
			StripeSubscriptionID: &stripeSubID,
			Tier:                 models.TierPro,
			Status:               "active",
		}

		webhookData := map[string]interface{}{
			"id": "inv_123",
			"parent": map[string]interface{}{
				"type": "subscription_details",
				"subscription_details": map[string]interface{}{
					"subscription": map[string]interface{}{
						"id": stripeSubID,
					},
				},
			},
		}
		data, _ := json.Marshal(webhookData)

		subRepo.On("FindByStripeSubscriptionID", mock.Anything, stripeSubID).Return(existingSub, nil)

		// Mock credits service for RefreshMonthlyCredits
		credits := &models.Credits{
			UserID:           userID,
			Balance:          20,
			MonthlyAllowance: 100,
			UsedThisPeriod:   80,
		}
		txRunner.On("Transaction", mock.Anything).Return(nil)
		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(credits, nil)
		creditsRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
		txRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		creditsService := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)

		service := NewStripeServiceForTest(cfg, nil, nil, subRepo, creditsService)
		err := service.handleInvoicePaid(data)

		assert.NoError(t, err)
		subRepo.AssertExpectations(t)
	})

	t.Run("returns nil when subscription not found", func(t *testing.T) {
		subRepo := new(mocks.MockSubscriptionRepository)
		cfg := &config.Config{}

		webhookData := map[string]interface{}{
			"id": "inv_123",
			"parent": map[string]interface{}{
				"type": "subscription_details",
				"subscription_details": map[string]interface{}{
					"subscription": map[string]interface{}{
						"id": stripeSubID,
					},
				},
			},
		}
		data, _ := json.Marshal(webhookData)

		subRepo.On("FindByStripeSubscriptionID", mock.Anything, stripeSubID).Return(nil, repository.ErrNotFound)

		service := NewStripeServiceForTest(cfg, nil, nil, subRepo, nil)
		err := service.handleInvoicePaid(data)

		assert.NoError(t, err)
	})

	t.Run("returns nil when parent is missing", func(t *testing.T) {
		cfg := &config.Config{}

		webhookData := map[string]interface{}{
			"id": "inv_123",
		}
		data, _ := json.Marshal(webhookData)

		service := NewStripeServiceForTest(cfg, nil, nil, nil, nil)
		err := service.handleInvoicePaid(data)

		assert.NoError(t, err)
	})
}

func TestStripeService_handleInvoicePaymentFailed(t *testing.T) {
	stripeSubID := "sub_test123"

	t.Run("updates status to past_due", func(t *testing.T) {
		subRepo := new(mocks.MockSubscriptionRepository)
		cfg := &config.Config{}

		webhookData := map[string]interface{}{
			"id": "inv_123",
			"parent": map[string]interface{}{
				"type": "subscription_details",
				"subscription_details": map[string]interface{}{
					"subscription": map[string]interface{}{
						"id": stripeSubID,
					},
				},
			},
		}
		data, _ := json.Marshal(webhookData)

		subRepo.On("UpdateStatus", mock.Anything, stripeSubID, "past_due").Return(nil)

		service := NewStripeServiceForTest(cfg, nil, nil, subRepo, nil)
		err := service.handleInvoicePaymentFailed(data)

		assert.NoError(t, err)
		subRepo.AssertExpectations(t)
	})

	t.Run("returns nil when parent is missing", func(t *testing.T) {
		cfg := &config.Config{}

		webhookData := map[string]interface{}{
			"id": "inv_123",
		}
		data, _ := json.Marshal(webhookData)

		service := NewStripeServiceForTest(cfg, nil, nil, nil, nil)
		err := service.handleInvoicePaymentFailed(data)

		assert.NoError(t, err)
	})
}

func TestStripeService_HandleWebhook(t *testing.T) {
	t.Run("returns error when webhook secret not configured", func(t *testing.T) {
		cfg := &config.Config{
			StripeWebhookSecret: "",
		}

		service := NewStripeServiceForTest(cfg, nil, nil, nil, nil)
		err := service.HandleWebhook([]byte("{}"), "sig")

		assert.ErrorIs(t, err, ErrInvalidWebhook)
		assert.Contains(t, err.Error(), "webhook secret not configured")
	})
}

// txRunner mock for stripe tests uses the same mockTxRunner from credits_test.go
type stripeTestTxRunner struct {
	mock.Mock
}

func (m *stripeTestTxRunner) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	args := m.Called(fc)
	if args.Error(0) == nil {
		return fc(nil)
	}
	return args.Error(0)
}
