package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	ID                   uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ThreadID             uuid.UUID `gorm:"type:uuid;index;not null" json:"threadId"`
	Role                 string    `gorm:"type:varchar(20);not null" json:"role"` // "user" or "assistant"
	Content              string    `gorm:"type:text;not null" json:"content"`
	AudioURL             *string   `gorm:"type:varchar(500)" json:"audioUrl,omitempty"`
	AudioDurationSeconds *float64  `gorm:"type:decimal(10,2)" json:"audioDurationSeconds,omitempty"`
	HasAudio             bool      `gorm:"default:false" json:"hasAudio"`
	Timestamp            time.Time `json:"timestamp"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
