package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
)

// MockThreadRepository is a mock implementation of ThreadRepository for testing.
type MockThreadRepository struct {
	mock.Mock
}

// Ensure MockThreadRepository implements ThreadRepository.
var _ repository.ThreadRepository = (*MockThreadRepository)(nil)

func (m *MockThreadRepository) Create(exec repository.Executor, thread *models.Thread) error {
	args := m.Called(exec, thread)
	return args.Error(0)
}

func (m *MockThreadRepository) FindByID(exec repository.Executor, id uuid.UUID) (*models.Thread, error) {
	args := m.Called(exec, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Thread), args.Error(1)
}

func (m *MockThreadRepository) FindByIDWithMessages(exec repository.Executor, id uuid.UUID) (*models.Thread, error) {
	args := m.Called(exec, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Thread), args.Error(1)
}

func (m *MockThreadRepository) FindByUserID(exec repository.Executor, userID uuid.UUID) ([]models.Thread, error) {
	args := m.Called(exec, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Thread), args.Error(1)
}

func (m *MockThreadRepository) FindArchivedByUserID(exec repository.Executor, userID uuid.UUID) ([]models.Thread, error) {
	args := m.Called(exec, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Thread), args.Error(1)
}

func (m *MockThreadRepository) FindByIDAndUserID(exec repository.Executor, id, userID uuid.UUID) (*models.Thread, error) {
	args := m.Called(exec, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Thread), args.Error(1)
}

func (m *MockThreadRepository) FindByIDAndUserIDWithMessages(exec repository.Executor, id, userID uuid.UUID) (*models.Thread, error) {
	args := m.Called(exec, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Thread), args.Error(1)
}

func (m *MockThreadRepository) Save(exec repository.Executor, thread *models.Thread) error {
	args := m.Called(exec, thread)
	return args.Error(0)
}

func (m *MockThreadRepository) Delete(exec repository.Executor, thread *models.Thread) error {
	args := m.Called(exec, thread)
	return args.Error(0)
}

func (m *MockThreadRepository) UpdateName(exec repository.Executor, id uuid.UUID, name string) error {
	args := m.Called(exec, id, name)
	return args.Error(0)
}
