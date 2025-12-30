package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ling-app/api/internal/models"
)

// threadRepository implements ThreadRepository using GORM.
type threadRepository struct{}

// NewThreadRepository creates a new GORM-backed thread repository.
func NewThreadRepository() ThreadRepository {
	return &threadRepository{}
}

func (r *threadRepository) Create(exec Executor, thread *models.Thread) error {
	return exec.Create(thread).Error
}

func (r *threadRepository) FindByID(exec Executor, id uuid.UUID) (*models.Thread, error) {
	var thread models.Thread
	err := exec.First(&thread, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *threadRepository) FindByIDWithMessages(exec Executor, id uuid.UUID) (*models.Thread, error) {
	var thread models.Thread
	err := exec.Preload("Messages").First(&thread, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *threadRepository) FindByUserID(exec Executor, userID uuid.UUID) ([]models.Thread, error) {
	var threads []models.Thread
	err := exec.Where("user_id = ? AND archived_at IS NULL", userID).Order("created_at DESC").Find(&threads).Error
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (r *threadRepository) FindArchivedByUserID(exec Executor, userID uuid.UUID) ([]models.Thread, error) {
	var threads []models.Thread
	err := exec.Where("user_id = ? AND archived_at IS NOT NULL", userID).Order("archived_at DESC").Find(&threads).Error
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (r *threadRepository) FindByIDAndUserID(exec Executor, id, userID uuid.UUID) (*models.Thread, error) {
	var thread models.Thread
	err := exec.Where("id = ? AND user_id = ?", id, userID).First(&thread).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *threadRepository) FindByIDAndUserIDWithMessages(exec Executor, id, userID uuid.UUID) (*models.Thread, error) {
	var thread models.Thread
	err := exec.Preload("Messages").Where("id = ? AND user_id = ?", id, userID).First(&thread).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *threadRepository) Save(exec Executor, thread *models.Thread) error {
	return exec.Save(thread).Error
}

func (r *threadRepository) Delete(exec Executor, thread *models.Thread) error {
	return exec.Delete(thread).Error
}

func (r *threadRepository) UpdateName(exec Executor, id uuid.UUID, name string) error {
	return exec.Model(&models.Thread{}).Where("id = ?", id).Update("name", name).Error
}
