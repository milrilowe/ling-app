package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Credit cost per voice message (only input type in this pronunciation app)
const CreditCostPerMessage = 1

// Credits tracks a user's credit balance
type Credits struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"userId"`

	// Credit balance
	Balance          int `gorm:"not null;default:20" json:"balance"`
	MonthlyAllowance int `gorm:"not null;default:20" json:"monthlyAllowance"`

	// Tracking
	UsedThisPeriod  int       `gorm:"not null;default:0" json:"usedThisPeriod"`
	LastRefreshedAt time.Time `json:"lastRefreshedAt"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BeforeCreate generates a UUID for new credit records
func (c *Credits) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// HasEnough returns true if the user has enough credits for an operation
func (c *Credits) HasEnough(amount int) bool {
	return c.Balance >= amount
}

// CreditTransactionType represents the type of credit transaction
type CreditTransactionType string

const (
	TransactionDebit   CreditTransactionType = "debit"
	TransactionCredit  CreditTransactionType = "credit"
	TransactionRefresh CreditTransactionType = "refresh"
)

// CreditTransaction records credit balance changes for auditing
type CreditTransaction struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;index;not null" json:"userId"`

	Type         CreditTransactionType `gorm:"type:varchar(50);not null" json:"type"`
	Amount       int                   `gorm:"not null" json:"amount"` // Positive for credit, negative for debit
	BalanceAfter int                   `gorm:"not null" json:"balanceAfter"`

	// Reference to what triggered this (message ID, subscription change, etc.)
	Reference   *string `gorm:"type:varchar(255)" json:"reference,omitempty"`
	Description string  `gorm:"type:varchar(500)" json:"description"`

	CreatedAt time.Time `json:"createdAt"`
}

// BeforeCreate generates a UUID for new transactions
func (t *CreditTransaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
