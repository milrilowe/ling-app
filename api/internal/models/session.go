package models

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated session stored in the database.
// This is used for session-based auth with HTTP-only cookies.
//
// Why database sessions instead of JWTs?
// - Sessions can be revoked instantly (logout, security breach)
// - No token size limits in cookies
// - Server controls session lifetime
// - Trade-off: requires DB lookup per request (negligible with indexes)
type Session struct {
	// ID is a cryptographically random token (base64url encoded, 32 bytes)
	// This is what gets stored in the cookie
	ID string `gorm:"type:varchar(64);primary_key"`

	// UserID links to the authenticated user
	UserID uuid.UUID `gorm:"type:uuid;index;not null"`
	User   User      `gorm:"foreignKey:UserID"`

	// Metadata for security auditing
	UserAgent string `gorm:"type:varchar(500)"`
	IPAddress string `gorm:"type:varchar(45)"` // IPv6 can be up to 45 chars

	// ExpiresAt determines when the session becomes invalid
	// Index allows efficient cleanup of expired sessions
	ExpiresAt time.Time `gorm:"index;not null"`
	CreatedAt time.Time
}

// IsExpired checks if the session has passed its expiration time
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
