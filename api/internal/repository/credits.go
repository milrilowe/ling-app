package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ling-app/api/internal/models"
)

// creditsRepository implements CreditsRepository using GORM.
type creditsRepository struct{}

// NewCreditsRepository creates a new GORM-backed credits repository.
func NewCreditsRepository() CreditsRepository {
	return &creditsRepository{}
}

func (r *creditsRepository) FindByUserID(exec Executor, userID uuid.UUID) (*models.Credits, error) {
	var credits models.Credits
	err := exec.Where("user_id = ?", userID).First(&credits).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &credits, nil
}

func (r *creditsRepository) Create(exec Executor, credits *models.Credits) error {
	return exec.Create(credits).Error
}

func (r *creditsRepository) Save(exec Executor, credits *models.Credits) error {
	return exec.Save(credits).Error
}

func (r *creditsRepository) UpdateAllowance(exec Executor, userID uuid.UUID, allowance int) error {
	return exec.Model(&models.Credits{}).Where("user_id = ?", userID).Update("monthly_allowance", allowance).Error
}

// creditTransactionRepository implements CreditTransactionRepository using GORM.
type creditTransactionRepository struct{}

// NewCreditTransactionRepository creates a new GORM-backed credit transaction repository.
func NewCreditTransactionRepository() CreditTransactionRepository {
	return &creditTransactionRepository{}
}

func (r *creditTransactionRepository) Create(exec Executor, tx *models.CreditTransaction) error {
	return exec.Create(tx).Error
}

func (r *creditTransactionRepository) FindByUserID(exec Executor, userID uuid.UUID, limit int) ([]models.CreditTransaction, error) {
	var transactions []models.CreditTransaction
	err := exec.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}
