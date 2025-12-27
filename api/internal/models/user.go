package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an authenticated user in the system.
// Users can authenticate via email/password OR OAuth providers (Google/GitHub).
// OAuth-only users will have a nil PasswordHash.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash *string   `gorm:"type:varchar(255)" json:"-"` // json:"-" means never serialize to JSON
	Name         string    `gorm:"type:varchar(255)" json:"name"`
	AvatarURL    *string   `gorm:"type:varchar(500)" json:"avatarUrl,omitempty"`

	// OAuth provider IDs - uniqueIndex allows nil but enforces uniqueness for non-nil values
	GoogleID *string `gorm:"type:varchar(255);uniqueIndex" json:"-"`
	GitHubID *string `gorm:"type:varchar(255);uniqueIndex" json:"-"`

	// Account status
	EmailVerified bool `gorm:"default:false" json:"emailVerified"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Relationships - a user has many threads
	Threads []Thread `gorm:"foreignKey:UserID" json:"-"`

	// Subscription and credits
	Subscription *Subscription `gorm:"foreignKey:UserID" json:"-"`
	Credits      *Credits      `gorm:"foreignKey:UserID" json:"-"`
}

// BeforeCreate is a GORM hook that runs before inserting a new record.
// It auto-generates a UUID if one isn't set.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
