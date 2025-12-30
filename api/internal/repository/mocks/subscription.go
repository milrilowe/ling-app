package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
)

// MockSubscriptionRepository is a mock implementation of SubscriptionRepository for testing.
type MockSubscriptionRepository struct {
	mock.Mock
}

// Ensure MockSubscriptionRepository implements SubscriptionRepository.
var _ repository.SubscriptionRepository = (*MockSubscriptionRepository)(nil)

func (m *MockSubscriptionRepository) FindByUserID(exec repository.Executor, userID uuid.UUID) (*models.Subscription, error) {
	args := m.Called(exec, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByStripeCustomerID(exec repository.Executor, customerID string) (*models.Subscription, error) {
	args := m.Called(exec, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByStripeSubscriptionID(exec repository.Executor, subscriptionID string) (*models.Subscription, error) {
	args := m.Called(exec, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Create(exec repository.Executor, sub *models.Subscription) error {
	args := m.Called(exec, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) Save(exec repository.Executor, sub *models.Subscription) error {
	args := m.Called(exec, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) UpdateStatus(exec repository.Executor, subscriptionID string, status string) error {
	args := m.Called(exec, subscriptionID, status)
	return args.Error(0)
}
