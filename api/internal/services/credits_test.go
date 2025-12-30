package services

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"ling-app/api/internal/models"
	"ling-app/api/internal/repository"
	"ling-app/api/internal/repository/mocks"
)

// mockTxRunner is a mock implementation of TxRunner for testing.
type mockTxRunner struct {
	mock.Mock
}

func (m *mockTxRunner) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	args := m.Called(fc)
	// Execute the transaction function with nil (mocks handle the logic)
	if args.Error(0) == nil {
		return fc(nil)
	}
	return args.Error(0)
}

func TestCreditsService_GetCredits(t *testing.T) {
	userID := uuid.New()

	t.Run("returns credits when found", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		expectedCredits := &models.Credits{
			UserID:           userID,
			Balance:          100,
			MonthlyAllowance: 100,
		}

		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(expectedCredits, nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		credits, err := service.GetCredits(userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedCredits, credits)
		creditsRepo.AssertExpectations(t)
	})

	t.Run("returns ErrCreditsNotFound when not found", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		credits, err := service.GetCredits(userID)

		assert.ErrorIs(t, err, ErrCreditsNotFound)
		assert.Nil(t, credits)
		creditsRepo.AssertExpectations(t)
	})

	t.Run("wraps other errors", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		dbError := errors.New("database connection failed")
		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(nil, dbError)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		credits, err := service.GetCredits(userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get credits")
		assert.Nil(t, credits)
		creditsRepo.AssertExpectations(t)
	})
}

func TestCreditsService_GetBalance(t *testing.T) {
	userID := uuid.New()

	t.Run("returns balance when credits exist", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(&models.Credits{
			UserID:  userID,
			Balance: 50,
		}, nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		balance, err := service.GetBalance(userID)

		assert.NoError(t, err)
		assert.Equal(t, 50, balance)
	})

	t.Run("returns error when credits not found", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		balance, err := service.GetBalance(userID)

		assert.ErrorIs(t, err, ErrCreditsNotFound)
		assert.Equal(t, 0, balance)
	})
}

func TestCreditsService_HasCredits(t *testing.T) {
	userID := uuid.New()

	t.Run("returns true when user has enough credits", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(&models.Credits{
			UserID:  userID,
			Balance: 100,
		}, nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		hasCredits, err := service.HasCredits(userID, 50)

		assert.NoError(t, err)
		assert.True(t, hasCredits)
	})

	t.Run("returns false when user has insufficient credits", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(&models.Credits{
			UserID:  userID,
			Balance: 10,
		}, nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		hasCredits, err := service.HasCredits(userID, 50)

		assert.NoError(t, err)
		assert.False(t, hasCredits)
	})

	t.Run("returns false when credits not found", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(nil, repository.ErrNotFound)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		hasCredits, err := service.HasCredits(userID, 50)

		assert.NoError(t, err)
		assert.False(t, hasCredits)
	})
}

func TestCreditsService_DeductCredits(t *testing.T) {
	userID := uuid.New()

	t.Run("successfully deducts credits", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		credits := &models.Credits{
			UserID:         userID,
			Balance:        100,
			UsedThisPeriod: 0,
		}

		txRunner.On("Transaction", mock.Anything).Return(nil)
		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(credits, nil)
		creditsRepo.On("Save", mock.Anything, mock.MatchedBy(func(c *models.Credits) bool {
			return c.Balance == 90 && c.UsedThisPeriod == 10
		})).Return(nil)
		txRepo.On("Create", mock.Anything, mock.MatchedBy(func(tx *models.CreditTransaction) bool {
			return tx.Amount == -10 && tx.Type == models.TransactionDebit
		})).Return(nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		err := service.DeductCredits(userID, 10, "msg-123", "Test deduction")

		assert.NoError(t, err)
		txRunner.AssertExpectations(t)
		creditsRepo.AssertExpectations(t)
		txRepo.AssertExpectations(t)
	})

	t.Run("returns ErrInsufficientCredits when balance too low", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		credits := &models.Credits{
			UserID:  userID,
			Balance: 5,
		}

		txRunner.On("Transaction", mock.Anything).Return(nil)
		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(credits, nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		err := service.DeductCredits(userID, 10, "msg-123", "Test deduction")

		assert.ErrorIs(t, err, ErrInsufficientCredits)
	})
}

func TestCreditsService_AddCredits(t *testing.T) {
	userID := uuid.New()

	t.Run("successfully adds credits", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		credits := &models.Credits{
			UserID:  userID,
			Balance: 50,
		}

		txRunner.On("Transaction", mock.Anything).Return(nil)
		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(credits, nil)
		creditsRepo.On("Save", mock.Anything, mock.MatchedBy(func(c *models.Credits) bool {
			return c.Balance == 100
		})).Return(nil)
		txRepo.On("Create", mock.Anything, mock.MatchedBy(func(tx *models.CreditTransaction) bool {
			return tx.Amount == 50 && tx.Type == models.TransactionCredit
		})).Return(nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		err := service.AddCredits(userID, 50, "Bonus credits")

		assert.NoError(t, err)
		txRunner.AssertExpectations(t)
		creditsRepo.AssertExpectations(t)
		txRepo.AssertExpectations(t)
	})
}

func TestCreditsService_RefreshMonthlyCredits(t *testing.T) {
	userID := uuid.New()

	t.Run("successfully refreshes credits", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		credits := &models.Credits{
			UserID:           userID,
			Balance:          20,
			MonthlyAllowance: 100,
			UsedThisPeriod:   80,
		}

		txRunner.On("Transaction", mock.Anything).Return(nil)
		creditsRepo.On("FindByUserID", mock.Anything, userID).Return(credits, nil)
		creditsRepo.On("Save", mock.Anything, mock.MatchedBy(func(c *models.Credits) bool {
			return c.Balance == 100 && c.UsedThisPeriod == 0
		})).Return(nil)
		txRepo.On("Create", mock.Anything, mock.MatchedBy(func(tx *models.CreditTransaction) bool {
			return tx.Amount == 80 && tx.Type == models.TransactionRefresh
		})).Return(nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		err := service.RefreshMonthlyCredits(userID)

		assert.NoError(t, err)
		txRunner.AssertExpectations(t)
		creditsRepo.AssertExpectations(t)
		txRepo.AssertExpectations(t)
	})
}

func TestCreditsService_InitializeCredits(t *testing.T) {
	userID := uuid.New()

	t.Run("initializes credits with free tier allowance", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		creditsRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *models.Credits) bool {
			return c.UserID == userID &&
				c.Balance == models.TierCredits[models.TierFree] &&
				c.MonthlyAllowance == models.TierCredits[models.TierFree]
		})).Return(nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		err := service.InitializeCredits(userID, models.TierFree)

		assert.NoError(t, err)
		creditsRepo.AssertExpectations(t)
	})

	t.Run("initializes credits with pro tier allowance", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		creditsRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *models.Credits) bool {
			return c.Balance == models.TierCredits[models.TierPro]
		})).Return(nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		err := service.InitializeCredits(userID, models.TierPro)

		assert.NoError(t, err)
		creditsRepo.AssertExpectations(t)
	})
}

func TestCreditsService_UpdateAllowance(t *testing.T) {
	userID := uuid.New()

	t.Run("updates allowance based on tier", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		creditsRepo.On("UpdateAllowance", mock.Anything, userID, models.TierCredits[models.TierBasic]).Return(nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		err := service.UpdateAllowance(userID, models.TierBasic)

		assert.NoError(t, err)
		creditsRepo.AssertExpectations(t)
	})
}

func TestCreditsService_GetTransactionHistory(t *testing.T) {
	userID := uuid.New()

	t.Run("returns transaction history", func(t *testing.T) {
		creditsRepo := new(mocks.MockCreditsRepository)
		txRepo := new(mocks.MockCreditTransactionRepository)
		txRunner := new(mockTxRunner)

		expectedTxs := []models.CreditTransaction{
			{UserID: userID, Amount: -10, Type: models.TransactionDebit},
			{UserID: userID, Amount: 100, Type: models.TransactionCredit},
		}

		txRepo.On("FindByUserID", mock.Anything, userID, 50).Return(expectedTxs, nil)

		service := NewCreditsServiceForTest(nil, txRunner, creditsRepo, txRepo)
		txs, err := service.GetTransactionHistory(userID, 50)

		assert.NoError(t, err)
		assert.Equal(t, expectedTxs, txs)
		txRepo.AssertExpectations(t)
	})
}
