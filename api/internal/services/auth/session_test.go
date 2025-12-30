package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
	"ling-app/api/internal/repository/mocks"
)

func TestCreateSession(t *testing.T) {
	t.Run("creates session successfully", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockSessionRepo := &mocks.MockSessionRepository{}
		mockUserRepo := &mocks.MockUserRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		userID := uuid.New()

		// Expect session to be created
		mockSessionRepo.On("Create", mockExec, mock.AnythingOfType("*models.Session")).
			Return(nil).
			Run(func(args mock.Arguments) {
				session := args.Get(1).(*models.Session)
				assert.Equal(t, userID, session.UserID)
				assert.Equal(t, "Mozilla/5.0", session.UserAgent)
				assert.Equal(t, "127.0.0.1", session.IPAddress)
				assert.NotEmpty(t, session.ID)
				assert.False(t, session.ExpiresAt.IsZero())
			})

		token, err := service.CreateSession(userID, "Mozilla/5.0", "127.0.0.1")

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockSessionRepo := &mocks.MockSessionRepository{}
		mockUserRepo := &mocks.MockUserRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		mockSessionRepo.On("Create", mockExec, mock.AnythingOfType("*models.Session")).
			Return(assert.AnError)

		token, err := service.CreateSession(uuid.New(), "Mozilla/5.0", "127.0.0.1")

		assert.Error(t, err)
		assert.Empty(t, token)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestValidateSession(t *testing.T) {
	t.Run("returns user for valid session", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockSessionRepo := &mocks.MockSessionRepository{}
		mockUserRepo := &mocks.MockUserRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		expectedUser := &models.User{
			ID:    uuid.New(),
			Email: "test@example.com",
			Name:  "Test User",
		}

		session := &models.Session{
			ID:        "valid-token",
			UserID:    expectedUser.ID,
			User:      *expectedUser,
			ExpiresAt: time.Now().Add(time.Hour), // Not expired
		}

		mockSessionRepo.On("FindByIDWithUser", mockExec, "valid-token").
			Return(session, nil)

		user, err := service.ValidateSession("valid-token")

		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("returns error for non-existent session", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockSessionRepo := &mocks.MockSessionRepository{}
		mockUserRepo := &mocks.MockUserRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		mockSessionRepo.On("FindByIDWithUser", mockExec, "invalid-token").
			Return(nil, repository.ErrNotFound)

		user, err := service.ValidateSession("invalid-token")

		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.Nil(t, user)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("returns error and cleans up expired session", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockSessionRepo := &mocks.MockSessionRepository{}
		mockUserRepo := &mocks.MockUserRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		expiredSession := &models.Session{
			ID:        "expired-token",
			UserID:    uuid.New(),
			ExpiresAt: time.Now().Add(-time.Hour), // Expired
		}

		mockSessionRepo.On("FindByIDWithUser", mockExec, "expired-token").
			Return(expiredSession, nil)
		mockSessionRepo.On("DeleteByID", mockExec, "expired-token").
			Return(nil)

		user, err := service.ValidateSession("expired-token")

		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.Nil(t, user)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestDeleteSession(t *testing.T) {
	t.Run("deletes session successfully", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockSessionRepo := &mocks.MockSessionRepository{}
		mockUserRepo := &mocks.MockUserRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		mockSessionRepo.On("DeleteByID", mockExec, "session-token").
			Return(nil)

		err := service.DeleteSession("session-token")

		assert.NoError(t, err)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestDeleteAllUserSessions(t *testing.T) {
	t.Run("deletes all user sessions successfully", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockSessionRepo := &mocks.MockSessionRepository{}
		mockUserRepo := &mocks.MockUserRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		userID := uuid.New()
		mockSessionRepo.On("DeleteByUserID", mockExec, userID).
			Return(nil)

		err := service.DeleteAllUserSessions(userID)

		assert.NoError(t, err)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestCleanupExpiredSessions(t *testing.T) {
	t.Run("cleans up expired sessions and returns count", func(t *testing.T) {
		mockExec := &mocks.MockExecutor{}
		mockSessionRepo := &mocks.MockSessionRepository{}
		mockUserRepo := &mocks.MockUserRepository{}

		service := NewAuthServiceForTest(mockExec, nil, mockUserRepo, mockSessionRepo, 86400)

		mockSessionRepo.On("DeleteExpiredBefore", mockExec, mock.AnythingOfType("time.Time")).
			Return(int64(5), nil)

		count, err := service.CleanupExpiredSessions()

		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
		mockSessionRepo.AssertExpectations(t)
	})
}
