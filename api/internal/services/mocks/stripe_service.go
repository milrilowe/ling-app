package mocks

import (
	"ling-app/api/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockStripeProcessor is a mock implementation of StripeProcessor interface
type MockStripeProcessor struct {
	mock.Mock
}

func (m *MockStripeProcessor) GetSubscription(userID uuid.UUID) (*models.Subscription, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockStripeProcessor) GetOrCreateSubscription(userID uuid.UUID, email, name string) (*models.Subscription, error) {
	args := m.Called(userID, email, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockStripeProcessor) CreateCheckoutSession(userID uuid.UUID, email, name string, tier models.SubscriptionTier) (string, error) {
	args := m.Called(userID, email, name, tier)
	return args.String(0), args.Error(1)
}

func (m *MockStripeProcessor) CreatePortalSession(userID uuid.UUID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockStripeProcessor) HandleWebhook(payload []byte, signature string) error {
	args := m.Called(payload, signature)
	return args.Error(0)
}
