package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
)

// MockCreditsRepository is a mock implementation of CreditsRepository for testing.
type MockCreditsRepository struct {
	mock.Mock
}

// Ensure MockCreditsRepository implements CreditsRepository.
var _ repository.CreditsRepository = (*MockCreditsRepository)(nil)

func (m *MockCreditsRepository) FindByUserID(exec repository.Executor, userID uuid.UUID) (*models.Credits, error) {
	args := m.Called(exec, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Credits), args.Error(1)
}

func (m *MockCreditsRepository) Create(exec repository.Executor, credits *models.Credits) error {
	args := m.Called(exec, credits)
	return args.Error(0)
}

func (m *MockCreditsRepository) Save(exec repository.Executor, credits *models.Credits) error {
	args := m.Called(exec, credits)
	return args.Error(0)
}

func (m *MockCreditsRepository) UpdateAllowance(exec repository.Executor, userID uuid.UUID, allowance int) error {
	args := m.Called(exec, userID, allowance)
	return args.Error(0)
}

// MockCreditTransactionRepository is a mock implementation of CreditTransactionRepository for testing.
type MockCreditTransactionRepository struct {
	mock.Mock
}

// Ensure MockCreditTransactionRepository implements CreditTransactionRepository.
var _ repository.CreditTransactionRepository = (*MockCreditTransactionRepository)(nil)

func (m *MockCreditTransactionRepository) Create(exec repository.Executor, tx *models.CreditTransaction) error {
	args := m.Called(exec, tx)
	return args.Error(0)
}

func (m *MockCreditTransactionRepository) FindByUserID(exec repository.Executor, userID uuid.UUID, limit int) ([]models.CreditTransaction, error) {
	args := m.Called(exec, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CreditTransaction), args.Error(1)
}
