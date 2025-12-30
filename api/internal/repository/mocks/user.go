package mocks

import (
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
)

// MockUserRepository is a mock implementation of UserRepository for testing.
type MockUserRepository struct {
	mock.Mock
}

// Ensure MockUserRepository implements UserRepository.
var _ repository.UserRepository = (*MockUserRepository)(nil)

func (m *MockUserRepository) FindByEmail(exec repository.Executor, email string) (*models.User, error) {
	args := m.Called(exec, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByGoogleID(exec repository.Executor, googleID string) (*models.User, error) {
	args := m.Called(exec, googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByGitHubID(exec repository.Executor, githubID string) (*models.User, error) {
	args := m.Called(exec, githubID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Create(exec repository.Executor, user *models.User) error {
	args := m.Called(exec, user)
	return args.Error(0)
}

func (m *MockUserRepository) Save(exec repository.Executor, user *models.User) error {
	args := m.Called(exec, user)
	return args.Error(0)
}
