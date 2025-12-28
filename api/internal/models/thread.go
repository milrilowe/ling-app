package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Thread struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"-"` // Owner of the thread
	Name       *string    `gorm:"type:varchar(255)" json:"name"`
	ArchivedAt *time.Time `gorm:"index" json:"archivedAt,omitempty"`
	Messages   []Message  `gorm:"foreignKey:ThreadID;constraint:OnDelete:CASCADE" json:"messages"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func (t *Thread) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (t *Thread) AfterFind(tx *gorm.DB) error {
	if t.Messages == nil {
		t.Messages = []Message{}
	}
	return nil
}