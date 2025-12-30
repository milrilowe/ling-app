package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ling-app/api/internal/models"
)

// messageRepository implements MessageRepository using GORM.
type messageRepository struct{}

// NewMessageRepository creates a new GORM-backed message repository.
func NewMessageRepository() MessageRepository {
	return &messageRepository{}
}

func (r *messageRepository) Create(exec Executor, message *models.Message) error {
	return exec.Create(message).Error
}

func (r *messageRepository) FindByID(exec Executor, id uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := exec.First(&message, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) FindByThreadID(exec Executor, threadID uuid.UUID) ([]models.Message, error) {
	var messages []models.Message
	err := exec.Where("thread_id = ?", threadID).Order("timestamp ASC").Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) UpdatePronunciationStatus(exec Executor, id uuid.UUID, status string) error {
	return exec.Model(&models.Message{}).Where("id = ?", id).Update("pronunciation_status", status).Error
}

func (r *messageRepository) UpdatePronunciationAnalysis(exec Executor, id uuid.UUID, status string, analysis models.JSONMap, updatedAt time.Time) error {
	return exec.Model(&models.Message{}).
		Where("id = ?", id).
		Update("pronunciation_status", status).
		Update("pronunciation_analysis", analysis).
		Update("pronunciation_error", nil).
		Update("pronunciation_updated_at", updatedAt).Error
}

func (r *messageRepository) UpdatePronunciationError(exec Executor, id uuid.UUID, status string, errMsg string, updatedAt time.Time) error {
	return exec.Model(&models.Message{}).
		Where("id = ?", id).
		Update("pronunciation_status", status).
		Update("pronunciation_error", errMsg).
		Update("pronunciation_updated_at", updatedAt).Error
}
