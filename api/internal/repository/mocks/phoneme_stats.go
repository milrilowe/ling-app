package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
)

// MockPhonemeStatsRepository is a mock implementation of PhonemeStatsRepository for testing.
type MockPhonemeStatsRepository struct {
	mock.Mock
}

// Ensure MockPhonemeStatsRepository implements PhonemeStatsRepository.
var _ repository.PhonemeStatsRepository = (*MockPhonemeStatsRepository)(nil)

func (m *MockPhonemeStatsRepository) Upsert(exec repository.Executor, stats *models.PhonemeStats) error {
	args := m.Called(exec, stats)
	return args.Error(0)
}

func (m *MockPhonemeStatsRepository) FindByUserID(exec repository.Executor, userID uuid.UUID) ([]models.PhonemeStats, error) {
	args := m.Called(exec, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PhonemeStats), args.Error(1)
}

func (m *MockPhonemeStatsRepository) GetAccuracyRanking(exec repository.Executor, userID uuid.UUID) ([]repository.PhonemeAccuracy, error) {
	args := m.Called(exec, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.PhonemeAccuracy), args.Error(1)
}

// MockPhonemeSubstitutionRepository is a mock implementation of PhonemeSubstitutionRepository for testing.
type MockPhonemeSubstitutionRepository struct {
	mock.Mock
}

// Ensure MockPhonemeSubstitutionRepository implements PhonemeSubstitutionRepository.
var _ repository.PhonemeSubstitutionRepository = (*MockPhonemeSubstitutionRepository)(nil)

func (m *MockPhonemeSubstitutionRepository) Upsert(exec repository.Executor, sub *models.PhonemeSubstitution) error {
	args := m.Called(exec, sub)
	return args.Error(0)
}

func (m *MockPhonemeSubstitutionRepository) FindTopByUserID(exec repository.Executor, userID uuid.UUID, limit int) ([]models.PhonemeSubstitution, error) {
	args := m.Called(exec, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PhonemeSubstitution), args.Error(1)
}
