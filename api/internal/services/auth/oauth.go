package auth

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
)

// CreditsInitializer is the interface for initializing credits in a transaction.
// This avoids import cycles between auth and credits packages.
type CreditsInitializer interface {
	InitializeCreditsWithTx(exec repository.Executor, userID uuid.UUID, tier models.SubscriptionTier) error
}

// FindOrCreateOAuthUser finds an existing user by OAuth provider ID,
// or creates a new user if one doesn't exist.
// This is used when a user logs in via Google or GitHub.
// Returns: user, isNewUser (true if user was just created), error
// For new users, credits are initialized atomically in the same transaction.
func (s *AuthService) FindOrCreateOAuthUser(provider, providerID, email, name, avatarURL string, creditsService CreditsInitializer) (*models.User, bool, error) {
	// Try to find by provider ID first
	var user *models.User
	var err error

	switch provider {
	case "google":
		user, err = s.userRepo.FindByGoogleID(s.exec, providerID)
	case "github":
		user, err = s.userRepo.FindByGitHubID(s.exec, providerID)
	default:
		return nil, false, errors.New("unknown OAuth provider")
	}

	// Found existing user by OAuth ID
	if err == nil {
		return user, false, nil
	}

	if !errors.Is(err, repository.ErrNotFound) {
		return nil, false, err // Database error
	}

	// Try to find by email (user might have registered with email first)
	user, err = s.userRepo.FindByEmail(s.exec, email)
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
		if err := s.userRepo.Save(s.exec, user); err != nil {
			return nil, false, err
		}
		return user, false, nil
	}

	if !errors.Is(err, repository.ErrNotFound) {
		return nil, false, err // Database error
	}

	// Create new user with credits in a transaction
	newUser := &models.User{
		Email:         email,
		Name:          name,
		EmailVerified: true, // OAuth providers verify email
	}

	if avatarURL != "" {
		newUser.AvatarURL = &avatarURL
	}

	switch provider {
	case "google":
		newUser.GoogleID = &providerID
	case "github":
		newUser.GitHubID = &providerID
	}

	err = s.txRunner.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(tx, newUser); err != nil {
			return err
		}

		// Initialize credits for new user
		return creditsService.InitializeCreditsWithTx(tx, newUser.ID, models.TierFree)
	})

	if err != nil {
		return nil, false, err
	}

	return newUser, true, nil
}
