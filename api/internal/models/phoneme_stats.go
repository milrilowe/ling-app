package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PhonemeStats tracks per-user accuracy for each phoneme (IPA symbol)
type PhonemeStats struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_phoneme_stats_user_phoneme" json:"userId"`

	// The IPA phoneme symbol (e.g., "θ", "ɪ", "r")
	Phoneme string `gorm:"type:varchar(10);not null;uniqueIndex:idx_phoneme_stats_user_phoneme" json:"phoneme"`

	// Aggregate statistics
	TotalAttempts int `gorm:"not null;default:0" json:"totalAttempts"`
	CorrectCount  int `gorm:"not null;default:0" json:"correctCount"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BeforeCreate generates a UUID for new records
func (p *PhonemeStats) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// Accuracy returns the accuracy percentage (0-100)
func (p *PhonemeStats) Accuracy() float64 {
	if p.TotalAttempts == 0 {
		return 0
	}
	return float64(p.CorrectCount) / float64(p.TotalAttempts) * 100
}

// PhonemeSubstitution tracks common substitution patterns per user
// e.g., user often says /t/ instead of /θ/
type PhonemeSubstitution struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_phoneme_subs_user_expected_actual" json:"userId"`

	// The expected phoneme (what should have been said)
	ExpectedPhoneme string `gorm:"type:varchar(10);not null;uniqueIndex:idx_phoneme_subs_user_expected_actual" json:"expectedPhoneme"`

	// The actual phoneme (what was said instead)
	ActualPhoneme string `gorm:"type:varchar(10);not null;uniqueIndex:idx_phoneme_subs_user_expected_actual" json:"actualPhoneme"`

	// How many times this substitution occurred
	OccurrenceCount int `gorm:"not null;default:1" json:"occurrenceCount"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BeforeCreate generates a UUID for new records
func (p *PhonemeSubstitution) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
