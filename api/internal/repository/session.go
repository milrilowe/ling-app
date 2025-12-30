package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ling-app/api/internal/models"
)

// sessionRepository implements SessionRepository using GORM.
type sessionRepository struct{}

// NewSessionRepository creates a new GORM-backed session repository.
func NewSessionRepository() SessionRepository {
	return &sessionRepository{}
}

func (r *sessionRepository) Create(exec Executor, session *models.Session) error {
	return exec.Create(session).Error
}

func (r *sessionRepository) FindByIDWithUser(exec Executor, token string) (*models.Session, error) {
	var session models.Session
	err := exec.Preload("User").Where("id = ?", token).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) DeleteByID(exec Executor, token string) error {
	return exec.Where("id = ?", token).Delete(&models.Session{}).Error
}

func (r *sessionRepository) DeleteByUserID(exec Executor, userID uuid.UUID) error {
	return exec.Where("user_id = ?", userID).Delete(&models.Session{}).Error
}

func (r *sessionRepository) DeleteExpiredBefore(exec Executor, t time.Time) (int64, error) {
	result := exec.Where("expires_at < ?", t).Delete(&models.Session{})
	return result.RowsAffected, result.Error
}
