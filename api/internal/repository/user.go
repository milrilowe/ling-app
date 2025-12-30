package repository

import (
	"errors"

	"gorm.io/gorm"

	"ling-app/api/internal/models"
)

// userRepository implements UserRepository using GORM.
type userRepository struct{}

// NewUserRepository creates a new GORM-backed user repository.
func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) FindByEmail(exec Executor, email string) (*models.User, error) {
	var user models.User
	err := exec.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByGoogleID(exec Executor, googleID string) (*models.User, error) {
	var user models.User
	err := exec.Where("google_id = ?", googleID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByGitHubID(exec Executor, githubID string) (*models.User, error) {
	var user models.User
	err := exec.Where("github_id = ?", githubID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(exec Executor, user *models.User) error {
	return exec.Create(user).Error
}

func (r *userRepository) Save(exec Executor, user *models.User) error {
	return exec.Save(user).Error
}
