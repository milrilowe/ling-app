package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
)

// CreateSession creates a new session for a user.
// Returns the session token to be stored in a cookie.
func (s *AuthService) CreateSession(userID uuid.UUID, userAgent, ipAddress string) (string, error) {
	token, err := s.GenerateSessionToken()
	if err != nil {
		return "", err
	}

	session := &models.Session{
		ID:        token,
		UserID:    userID,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: time.Now().Add(s.sessionMaxAge),
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(s.exec, session); err != nil {
		return "", err
	}

	return token, nil
}

// ValidateSession checks if a session token is valid and returns the associated user.
func (s *AuthService) ValidateSession(token string) (*models.User, error) {
	session, err := s.sessionRepo.FindByIDWithUser(s.exec, token)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	if session.IsExpired() {
		// Clean up expired session
		_ = s.sessionRepo.DeleteByID(s.exec, token)
		return nil, ErrSessionNotFound
	}

	return &session.User, nil
}

// DeleteSession removes a session (logout).
func (s *AuthService) DeleteSession(token string) error {
	return s.sessionRepo.DeleteByID(s.exec, token)
}

// DeleteAllUserSessions removes all sessions for a user (logout everywhere).
// TODO: Wire to handler for "logout everywhere" feature
func (s *AuthService) DeleteAllUserSessions(userID uuid.UUID) error {
	return s.sessionRepo.DeleteByUserID(s.exec, userID)
}

// CleanupExpiredSessions removes all expired sessions from the database.
// This should be called periodically (e.g., via a background goroutine).
// TODO: Schedule as background job for session cleanup
func (s *AuthService) CleanupExpiredSessions() (int64, error) {
	return s.sessionRepo.DeleteExpiredBefore(s.exec, time.Now())
}
