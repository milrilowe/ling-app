package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ling-app/api/internal/db"
	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrCreditsNotFound     = errors.New("credits record not found")
)

// TxRunner is an interface for running database transactions.
type TxRunner interface {
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
}

// CreditsService handles credit balance operations
type CreditsService struct {
	db          *db.DB
	exec        repository.Executor
	txRunner    TxRunner
	creditsRepo repository.CreditsRepository
	txRepo      repository.CreditTransactionRepository
}

// NewCreditsService creates a new credits service
func NewCreditsService(
	database *db.DB,
	creditsRepo repository.CreditsRepository,
	txRepo repository.CreditTransactionRepository,
) *CreditsService {
	return &CreditsService{
		db:          database,
		exec:        database.DB,
		txRunner:    database.DB,
		creditsRepo: creditsRepo,
		txRepo:      txRepo,
	}
}

// NewCreditsServiceForTest creates a CreditsService with injected dependencies for testing.
func NewCreditsServiceForTest(
	exec repository.Executor,
	txRunner TxRunner,
	creditsRepo repository.CreditsRepository,
	txRepo repository.CreditTransactionRepository,
) *CreditsService {
	return &CreditsService{
		db:          nil,
		exec:        exec,
		txRunner:    txRunner,
		creditsRepo: creditsRepo,
		txRepo:      txRepo,
	}
}

// GetCredits returns the credits record for a user
func (s *CreditsService) GetCredits(userID uuid.UUID) (*models.Credits, error) {
	credits, err := s.creditsRepo.FindByUserID(s.exec, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCreditsNotFound
		}
		return nil, fmt.Errorf("failed to get credits: %w", err)
	}
	return credits, nil
}

// GetBalance returns the current credit balance for a user
func (s *CreditsService) GetBalance(userID uuid.UUID) (int, error) {
	credits, err := s.GetCredits(userID)
	if err != nil {
		return 0, err
	}
	return credits.Balance, nil
}

// HasCredits checks if a user has enough credits for an operation
func (s *CreditsService) HasCredits(userID uuid.UUID, amount int) (bool, error) {
	balance, err := s.GetBalance(userID)
	if err != nil {
		// If no credits record exists, they don't have credits
		if errors.Is(err, ErrCreditsNotFound) {
			return false, nil
		}
		return false, err
	}
	return balance >= amount, nil
}

// DeductCredits removes credits from a user's balance
func (s *CreditsService) DeductCredits(userID uuid.UUID, amount int, reference, description string) error {
	return s.txRunner.Transaction(func(tx *gorm.DB) error {
		credits, err := s.creditsRepo.FindByUserID(tx, userID)
		if err != nil {
			return fmt.Errorf("failed to get credits: %w", err)
		}

		if credits.Balance < amount {
			return ErrInsufficientCredits
		}

		// Update balance
		credits.Balance -= amount
		credits.UsedThisPeriod += amount
		if err := s.creditsRepo.Save(tx, credits); err != nil {
			return fmt.Errorf("failed to update credits: %w", err)
		}

		// Record transaction
		transaction := &models.CreditTransaction{
			UserID:       userID,
			Type:         models.TransactionDebit,
			Amount:       -amount,
			BalanceAfter: credits.Balance,
			Reference:    &reference,
			Description:  description,
		}
		if err := s.txRepo.Create(tx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		return nil
	})
}

// AddCredits adds credits to a user's balance
func (s *CreditsService) AddCredits(userID uuid.UUID, amount int, description string) error {
	return s.txRunner.Transaction(func(tx *gorm.DB) error {
		credits, err := s.creditsRepo.FindByUserID(tx, userID)
		if err != nil {
			return fmt.Errorf("failed to get credits: %w", err)
		}

		// Update balance
		credits.Balance += amount
		if err := s.creditsRepo.Save(tx, credits); err != nil {
			return fmt.Errorf("failed to update credits: %w", err)
		}

		// Record transaction
		transaction := &models.CreditTransaction{
			UserID:       userID,
			Type:         models.TransactionCredit,
			Amount:       amount,
			BalanceAfter: credits.Balance,
			Description:  description,
		}
		if err := s.txRepo.Create(tx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		return nil
	})
}

// RefreshMonthlyCredits resets the user's credits to their monthly allowance
func (s *CreditsService) RefreshMonthlyCredits(userID uuid.UUID) error {
	return s.txRunner.Transaction(func(tx *gorm.DB) error {
		credits, err := s.creditsRepo.FindByUserID(tx, userID)
		if err != nil {
			return fmt.Errorf("failed to get credits: %w", err)
		}

		// Reset to monthly allowance
		oldBalance := credits.Balance
		credits.Balance = credits.MonthlyAllowance
		credits.UsedThisPeriod = 0
		credits.LastRefreshedAt = time.Now()
		if err := s.creditsRepo.Save(tx, credits); err != nil {
			return fmt.Errorf("failed to update credits: %w", err)
		}

		// Record transaction
		transaction := &models.CreditTransaction{
			UserID:       userID,
			Type:         models.TransactionRefresh,
			Amount:       credits.MonthlyAllowance - oldBalance,
			BalanceAfter: credits.Balance,
			Description:  "Monthly credit refresh",
		}
		if err := s.txRepo.Create(tx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		return nil
	})
}

// InitializeCredits creates a credits record for a new user
func (s *CreditsService) InitializeCredits(userID uuid.UUID, tier models.SubscriptionTier) error {
	return s.InitializeCreditsWithTx(s.exec, userID, tier)
}

// InitializeCreditsWithTx creates a credits record using an existing transaction/executor
func (s *CreditsService) InitializeCreditsWithTx(exec repository.Executor, userID uuid.UUID, tier models.SubscriptionTier) error {
	allowance := models.TierCredits[tier]
	if allowance == 0 {
		allowance = models.TierCredits[models.TierFree]
	}

	credits := &models.Credits{
		UserID:           userID,
		Balance:          allowance,
		MonthlyAllowance: allowance,
		UsedThisPeriod:   0,
		LastRefreshedAt:  time.Now(),
	}

	if err := s.creditsRepo.Create(exec, credits); err != nil {
		return fmt.Errorf("failed to initialize credits: %w", err)
	}
	return nil
}

// UpdateAllowance updates the monthly allowance based on subscription tier
func (s *CreditsService) UpdateAllowance(userID uuid.UUID, tier models.SubscriptionTier) error {
	allowance := models.TierCredits[tier]
	if allowance == 0 {
		allowance = models.TierCredits[models.TierFree]
	}

	return s.creditsRepo.UpdateAllowance(s.exec, userID, allowance)
}

// GetTransactionHistory returns recent credit transactions for a user
func (s *CreditsService) GetTransactionHistory(userID uuid.UUID, limit int) ([]models.CreditTransaction, error) {
	transactions, err := s.txRepo.FindByUserID(s.exec, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}
	return transactions, nil
}
