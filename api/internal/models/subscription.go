package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionTier represents the user's subscription level
type SubscriptionTier string

const (
	TierFree  SubscriptionTier = "free"
	TierBasic SubscriptionTier = "basic"
	TierPro   SubscriptionTier = "pro"
)

// TierCredits defines how many credits each tier gets per month
var TierCredits = map[SubscriptionTier]int{
	TierFree:  20,
	TierBasic: 400,  // Increased from 200
	TierPro:   1200, // Increased from 600
}

// Subscription tracks a user's Stripe subscription status
type Subscription struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"userId"`

	// Stripe IDs
	StripeCustomerID     string  `gorm:"type:varchar(255);uniqueIndex" json:"-"`
	StripeSubscriptionID *string `gorm:"type:varchar(255);uniqueIndex" json:"-"`
	StripePriceID        *string `gorm:"type:varchar(255)" json:"-"`

	// Subscription details
	Tier   SubscriptionTier `gorm:"type:varchar(50);default:'free'" json:"tier"`
	Status string           `gorm:"type:varchar(50);default:'active'" json:"status"` // active, canceled, past_due

	// Billing period
	CurrentPeriodStart *time.Time `json:"currentPeriodStart,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"currentPeriodEnd,omitempty"`
	CancelAtPeriodEnd  bool       `gorm:"default:false" json:"cancelAtPeriodEnd"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BeforeCreate generates a UUID for new subscriptions
func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// IsPaid returns true if the subscription is a paid tier
func (s *Subscription) IsPaid() bool {
	return s.Tier == TierBasic || s.Tier == TierPro
}

// IsActive returns true if the subscription is in good standing
func (s *Subscription) IsActive() bool {
	return s.Status == "active"
}
