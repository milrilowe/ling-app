package mocks

import (
	"ling-app/api/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockCreditsManager is a mock implementation of CreditsManager interface
type MockCreditsManager struct {
	mock.Mock
}

func (m *MockCreditsManager) GetCredits(userID uuid.UUID) (*models.Credits, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Credits), args.Error(1)
}

func (m *MockCreditsManager) GetBalance(userID uuid.UUID) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *MockCreditsManager) HasCredits(userID uuid.UUID, amount int) (bool, error) {
	args := m.Called(userID, amount)
	return args.Bool(0), args.Error(1)
}

func (m *MockCreditsManager) DeductCredits(userID uuid.UUID, amount int, reference, description string) error {
	args := m.Called(userID, amount, reference, description)
	return args.Error(0)
}

func (m *MockCreditsManager) AddCredits(userID uuid.UUID, amount int, description string) error {
	args := m.Called(userID, amount, description)
	return args.Error(0)
}

func (m *MockCreditsManager) RefreshMonthlyCredits(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockCreditsManager) InitializeCredits(userID uuid.UUID, tier models.SubscriptionTier) error {
	args := m.Called(userID, tier)
	return args.Error(0)
}

func (m *MockCreditsManager) UpdateAllowance(userID uuid.UUID, tier models.SubscriptionTier) error {
	args := m.Called(userID, tier)
	return args.Error(0)
}

func (m *MockCreditsManager) GetTransactionHistory(userID uuid.UUID, limit int) ([]models.CreditTransaction, error) {
	args := m.Called(userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CreditTransaction), args.Error(1)
}
