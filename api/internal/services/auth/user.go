package auth

import (
	"errors"

	"gorm.io/gorm"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
)

// CreateUser registers a new user and initializes their credits atomically.
// Both operations happen in a single transaction - if either fails, both roll back.
func (s *AuthService) CreateUser(email, password, name string, creditsService CreditsInitializer) (*models.User, error) {
	var user *models.User

	err := s.txRunner.Transaction(func(tx *gorm.DB) error {
		// Check if email already exists
		_, err := s.userRepo.FindByEmail(tx, email)
		if err == nil {
			return ErrEmailTaken
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return err
		}

		// Hash the password
		hash, err := s.HashPassword(password)
		if err != nil {
			return err
		}

		// Create user
		user = &models.User{
			Email:        email,
			PasswordHash: &hash,
			Name:         name,
		}
		if err := s.userRepo.Create(tx, user); err != nil {
			return err
		}

		// Initialize credits
		return creditsService.InitializeCreditsWithTx(tx, user.ID, models.TierFree)
	})

	if err != nil {
		return nil, err
	}
	return user, nil
}

// AuthenticateUser validates email/password and returns the user if valid.
func (s *AuthService) AuthenticateUser(email, password string) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(s.exec, email)
	if errors.Is(err, repository.ErrNotFound) {
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

	return user, nil
}
