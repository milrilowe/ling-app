package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
)

// MockSessionRepository is a mock implementation of SessionRepository for testing.
type MockSessionRepository struct {
	mock.Mock
}

// Ensure MockSessionRepository implements SessionRepository.
var _ repository.SessionRepository = (*MockSessionRepository)(nil)

func (m *MockSessionRepository) Create(exec repository.Executor, session *models.Session) error {
	args := m.Called(exec, session)
	return args.Error(0)
}

func (m *MockSessionRepository) FindByIDWithUser(exec repository.Executor, token string) (*models.Session, error) {
	args := m.Called(exec, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) DeleteByID(exec repository.Executor, token string) error {
	args := m.Called(exec, token)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteByUserID(exec repository.Executor, userID uuid.UUID) error {
	args := m.Called(exec, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpiredBefore(exec repository.Executor, t time.Time) (int64, error) {
	args := m.Called(exec, t)
	return args.Get(0).(int64), args.Error(1)
}
