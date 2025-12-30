package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
)

// MockMessageRepository is a mock implementation of MessageRepository for testing.
type MockMessageRepository struct {
	mock.Mock
}

// Ensure MockMessageRepository implements MessageRepository.
var _ repository.MessageRepository = (*MockMessageRepository)(nil)

func (m *MockMessageRepository) Create(exec repository.Executor, message *models.Message) error {
	args := m.Called(exec, message)
	return args.Error(0)
}

func (m *MockMessageRepository) FindByID(exec repository.Executor, id uuid.UUID) (*models.Message, error) {
	args := m.Called(exec, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageRepository) FindByThreadID(exec repository.Executor, threadID uuid.UUID) ([]models.Message, error) {
	args := m.Called(exec, threadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Message), args.Error(1)
}

func (m *MockMessageRepository) UpdatePronunciationStatus(exec repository.Executor, id uuid.UUID, status string) error {
	args := m.Called(exec, id, status)
	return args.Error(0)
}

func (m *MockMessageRepository) UpdatePronunciationAnalysis(exec repository.Executor, id uuid.UUID, status string, analysis models.JSONMap, updatedAt time.Time) error {
	args := m.Called(exec, id, status, analysis, updatedAt)
	return args.Error(0)
}

func (m *MockMessageRepository) UpdatePronunciationError(exec repository.Executor, id uuid.UUID, status string, errMsg string, updatedAt time.Time) error {
	args := m.Called(exec, id, status, errMsg, updatedAt)
	return args.Error(0)
}
