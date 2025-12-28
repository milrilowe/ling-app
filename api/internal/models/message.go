package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONMap is a custom type that stores as JSONB in PostgreSQL but serializes
// as a JSON object (not a string) in API responses
type JSONMap map[string]interface{}

// Scan implements sql.Scanner for reading from the database
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// Value implements driver.Valuer for writing to the database
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

type Message struct {
	ID                   uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ThreadID             uuid.UUID `gorm:"type:uuid;index;not null" json:"threadId"`
	Role                 string    `gorm:"type:varchar(20);not null" json:"role"` // "user" or "assistant"
	Content              string    `gorm:"type:text;not null" json:"content"`
	AudioURL             *string   `gorm:"type:varchar(500)" json:"audioUrl,omitempty"`
	AudioDurationSeconds *float64  `gorm:"type:decimal(10,2)" json:"audioDurationSeconds,omitempty"`
	HasAudio             bool      `gorm:"default:false" json:"hasAudio"`
	Timestamp            time.Time `json:"timestamp"`

	// Pronunciation analysis fields (for user messages)
	PronunciationStatus    string     `gorm:"type:varchar(20);default:'none'" json:"pronunciationStatus"` // "none", "pending", "complete", "failed"
	PronunciationAnalysis  JSONMap    `gorm:"type:jsonb" json:"pronunciationAnalysis,omitempty"`          // Full analysis JSON object
	PronunciationError     *string    `gorm:"type:text" json:"pronunciationError,omitempty"`              // Error message if failed
	PronunciationUpdatedAt *time.Time `json:"pronunciationUpdatedAt,omitempty"`

}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
