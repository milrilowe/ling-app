package auth

import (
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
	"ling-app/api/internal/repository/mocks"
)

// mockCreditsInitializer is a mock for CreditsInitializer
type mockCreditsInitializer struct {
	mock.Mock
}

func (m *mockCreditsInitializer) InitializeCreditsWithTx(exec repository.Executor, userID uuid.UUID, tier models.SubscriptionTier) error {
	args := m.Called(exec, userID, tier)
	return args.Error(0)
}

// mockTxRunner executes the transaction function immediately with nil tx
type mockTxRunner struct {
	shouldFail bool
	failErr    error
}

func (m *mockTxRunner) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	if m.shouldFail {
		return m.failErr
	}
	// Execute the transaction function with nil (repos are mocked)
	return fc(nil)
}

func TestAuthenticateUser(t *testing.T) {
	t.Run("returns user for valid credentials", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockUserRepo := &mocks.MockUserRepository{}
		mockSessionRepo := &mocks.MockSessionRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		// Hash a password for testing
		hash, _ := service.HashPassword("correctpassword")

		expectedUser := &models.User{
			ID:           uuid.New(),
			Email:        "test@example.com",
			Name:         "Test User",
			PasswordHash: &hash,
		}

		mockUserRepo.On("FindByEmail", mockExec, "test@example.com").
			Return(expectedUser, nil)

		user, err := service.AuthenticateUser("test@example.com", "correctpassword")

		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockUserRepo := &mocks.MockUserRepository{}
		mockSessionRepo := &mocks.MockSessionRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		mockUserRepo.On("FindByEmail", mockExec, "nonexistent@example.com").
			Return(nil, repository.ErrNotFound)

		user, err := service.AuthenticateUser("nonexistent@example.com", "password")

		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns error for wrong password", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockUserRepo := &mocks.MockUserRepository{}
		mockSessionRepo := &mocks.MockSessionRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		hash, _ := service.HashPassword("correctpassword")
		existingUser := &models.User{
			ID:           uuid.New(),
			Email:        "test@example.com",
			PasswordHash: &hash,
		}

		mockUserRepo.On("FindByEmail", mockExec, "test@example.com").
			Return(existingUser, nil)

		user, err := service.AuthenticateUser("test@example.com", "wrongpassword")

		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns error for OAuth-only user without password", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockUserRepo := &mocks.MockUserRepository{}
		mockSessionRepo := &mocks.MockSessionRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		oauthUser := &models.User{
			ID:           uuid.New(),
			Email:        "oauth@example.com",
			PasswordHash: nil, // OAuth users don't have password
		}

		mockUserRepo.On("FindByEmail", mockExec, "oauth@example.com").
			Return(oauthUser, nil)

		user, err := service.AuthenticateUser("oauth@example.com", "anypassword")

		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestCreateUser(t *testing.T) {
	t.Run("creates user successfully", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockUserRepo := &mocks.MockUserRepository{}
		mockSessionRepo := &mocks.MockSessionRepository{}
		mockCredits := &mockCreditsInitializer{}
		mockTxRunner := &mockTxRunner{}

		service := NewAuthServiceForTest(mockExec, mockTxRunner, mockUserRepo, mockSessionRepo, 86400)

		// User doesn't exist yet
		mockUserRepo.On("FindByEmail", mock.Anything, "new@example.com").
			Return(nil, repository.ErrNotFound)

		// User will be created
		mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
			Return(nil).
			Run(func(args mock.Arguments) {
				user := args.Get(1).(*models.User)
				assert.Equal(t, "new@example.com", user.Email)
				assert.Equal(t, "New User", user.Name)
				assert.NotNil(t, user.PasswordHash)
				// Assign an ID like the DB would
				user.ID = uuid.New()
			})

		// Credits will be initialized
		mockCredits.On("InitializeCreditsWithTx", mock.Anything, mock.AnythingOfType("uuid.UUID"), models.TierFree).
			Return(nil)

		user, err := service.CreateUser("new@example.com", "password123", "New User", mockCredits)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "new@example.com", user.Email)
		assert.Equal(t, "New User", user.Name)
		mockUserRepo.AssertExpectations(t)
		mockCredits.AssertExpectations(t)
	})

	t.Run("returns error when email already exists", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockUserRepo := &mocks.MockUserRepository{}
		mockSessionRepo := &mocks.MockSessionRepository{}
		mockCredits := &mockCreditsInitializer{}
		mockTxRunner := &mockTxRunner{}

		service := NewAuthServiceForTest(mockExec, mockTxRunner, mockUserRepo, mockSessionRepo, 86400)

		existingUser := &models.User{
			ID:    uuid.New(),
			Email: "existing@example.com",
		}

		mockUserRepo.On("FindByEmail", mock.Anything, "existing@example.com").
			Return(existingUser, nil)

		user, err := service.CreateUser("existing@example.com", "password123", "New User", mockCredits)

		assert.ErrorIs(t, err, ErrEmailTaken)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("rolls back on credits initialization failure", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockUserRepo := &mocks.MockUserRepository{}
		mockSessionRepo := &mocks.MockSessionRepository{}
		mockCredits := &mockCreditsInitializer{}
		mockTxRunner := &mockTxRunner{}

		service := NewAuthServiceForTest(mockExec, mockTxRunner, mockUserRepo, mockSessionRepo, 86400)

		mockUserRepo.On("FindByEmail", mock.Anything, "new@example.com").
			Return(nil, repository.ErrNotFound)

		mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
			Return(nil).
			Run(func(args mock.Arguments) {
				user := args.Get(1).(*models.User)
				user.ID = uuid.New()
			})

		// Credits initialization fails
		mockCredits.On("InitializeCreditsWithTx", mock.Anything, mock.AnythingOfType("uuid.UUID"), models.TierFree).
			Return(assert.AnError)

		user, err := service.CreateUser("new@example.com", "password123", "New User", mockCredits)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
		mockCredits.AssertExpectations(t)
	})
}
