package auth

import (
	"database/sql"
	"time"

	"gorm.io/gorm"

	"ling-app/api/internal/db"
	"ling-app/api/internal/repository"
)

// TxRunner is an interface for running database transactions.
// Both *db.DB and *gorm.DB implement this.
type TxRunner interface {
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
}

// AuthService handles authentication logic including password hashing,
// session management, and user lookup.
type AuthService struct {
	db            *db.DB
	exec          repository.Executor // The executor to use for queries (usually s.db.DB)
	txRunner      TxRunner            // For running transactions
	userRepo      repository.UserRepository
	sessionRepo   repository.SessionRepository
	sessionMaxAge time.Duration
	bcryptCost    int
}

// NewAuthService creates a new auth service.
// sessionMaxAgeSec is how long sessions last (e.g., 86400 for 24 hours)
func NewAuthService(
	database *db.DB,
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	sessionMaxAgeSec int,
) *AuthService {
	return &AuthService{
		db:            database,
		exec:          database.DB,
		txRunner:      database.DB,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		sessionMaxAge: time.Duration(sessionMaxAgeSec) * time.Second,
		bcryptCost:    12, // Good balance of security vs speed
	}
}

// NewAuthServiceForTest creates an AuthService with injected dependencies for testing.
// This allows mocking the executor, transaction runner, and repositories.
func NewAuthServiceForTest(
	exec repository.Executor,
	txRunner TxRunner,
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	sessionMaxAgeSec int,
) *AuthService {
	return &AuthService{
		db:            nil, // Not needed in tests
		exec:          exec,
		txRunner:      txRunner,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		sessionMaxAge: time.Duration(sessionMaxAgeSec) * time.Second,
		bcryptCost:    4, // Low cost for fast tests
	}
}
