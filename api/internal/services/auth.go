package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"ling-app/api/internal/db"
	"ling-app/api/internal/models"
)

// Common auth errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrSessionNotFound    = errors.New("session not found or expired")
)

// AuthService handles authentication logic including password hashing,
// session management, and user lookup.
type AuthService struct {
	db             *db.DB
	sessionMaxAge  time.Duration
	bcryptCost     int
}

// NewAuthService creates a new auth service.
// sessionMaxAgeSec is how long sessions last (e.g., 86400 for 24 hours)
func NewAuthService(database *db.DB, sessionMaxAgeSec int) *AuthService {
	return &AuthService{
		db:            database,
		sessionMaxAge: time.Duration(sessionMaxAgeSec) * time.Second,
		bcryptCost:    12, // Good balance of security vs speed
	}
}

// HashPassword creates a bcrypt hash of the password.
// bcrypt automatically handles salting - each hash includes a unique salt.
func (s *AuthService) HashPassword(password string) (string, error) {
	// bcrypt.GenerateFromPassword:
	// - Adds a random salt automatically
	// - The cost factor (12) means 2^12 iterations
	// - Returns a string containing: algorithm + cost + salt + hash
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword verifies a password against its hash.
// Returns true if the password matches.
func (s *AuthService) CheckPassword(hash, password string) bool {
	// bcrypt.CompareHashAndPassword extracts the salt from the hash
	// and uses it to hash the provided password, then compares
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateSessionToken creates a cryptographically secure random token.
// Returns a 32-byte random value encoded as base64url (43 characters).
func (s *AuthService) GenerateSessionToken() (string, error) {
	// crypto/rand uses the OS's cryptographic random source
	// (e.g., /dev/urandom on Linux)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// base64url encoding is URL-safe (no +, /, or = padding issues)
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CreateUser registers a new user with email and password.
// Returns the created user or an error if email is already taken.
func (s *AuthService) CreateUser(email, password, name string) (*models.User, error) {
	// Check if email already exists
	var existing models.User
	err := s.db.Where("email = ?", email).First(&existing).Error
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err // Database error
	}

	// Hash the password
	hash, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:        email,
		PasswordHash: &hash,
		Name:         name,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser validates email/password and returns the user if valid.
func (s *AuthService) AuthenticateUser(email, password string) (*models.User, error) {
	var user models.User
	err := s.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	// OAuth-only users won't have a password hash
	if user.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}

	if !s.CheckPassword(*user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

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

	if err := s.db.Create(session).Error; err != nil {
		return "", err
	}

	return token, nil
}

// ValidateSession checks if a session token is valid and returns the associated user.
func (s *AuthService) ValidateSession(token string) (*models.User, error) {
	var session models.Session
	err := s.db.Preload("User").Where("id = ?", token).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	if session.IsExpired() {
		// Clean up expired session
		s.db.Delete(&session)
		return nil, ErrSessionNotFound
	}

	return &session.User, nil
}

// DeleteSession removes a session (logout).
func (s *AuthService) DeleteSession(token string) error {
	return s.db.Where("id = ?", token).Delete(&models.Session{}).Error
}

// DeleteAllUserSessions removes all sessions for a user (logout everywhere).
func (s *AuthService) DeleteAllUserSessions(userID uuid.UUID) error {
	return s.db.Where("user_id = ?", userID).Delete(&models.Session{}).Error
}

// CleanupExpiredSessions removes all expired sessions from the database.
// This should be called periodically (e.g., via a background goroutine).
func (s *AuthService) CleanupExpiredSessions() (int64, error) {
	result := s.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{})
	return result.RowsAffected, result.Error
}

// FindOrCreateOAuthUser finds an existing user by OAuth provider ID,
// or creates a new user if one doesn't exist.
// This is used when a user logs in via Google or GitHub.
func (s *AuthService) FindOrCreateOAuthUser(provider, providerID, email, name, avatarURL string) (*models.User, error) {
	var user models.User
	var err error

	// Try to find by provider ID first
	switch provider {
	case "google":
		err = s.db.Where("google_id = ?", providerID).First(&user).Error
	case "github":
		err = s.db.Where("github_id = ?", providerID).First(&user).Error
	default:
		return nil, errors.New("unknown OAuth provider")
	}

	// Found existing user by OAuth ID
	if err == nil {
		return &user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err // Database error
	}

	// Try to find by email (user might have registered with email first)
	err = s.db.Where("email = ?", email).First(&user).Error
	if err == nil {
		// Link OAuth to existing account
		switch provider {
		case "google":
			user.GoogleID = &providerID
		case "github":
			user.GitHubID = &providerID
		}
		if avatarURL != "" && user.AvatarURL == nil {
			user.AvatarURL = &avatarURL
		}
		if err := s.db.Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err // Database error
	}

	// Create new user
	user = models.User{
		Email:         email,
		Name:          name,
		EmailVerified: true, // OAuth providers verify email
	}

	if avatarURL != "" {
		user.AvatarURL = &avatarURL
	}

	switch provider {
	case "google":
		user.GoogleID = &providerID
	case "github":
		user.GitHubID = &providerID
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
