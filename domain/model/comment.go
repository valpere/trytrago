package model

import (
	"time"

	"github.com/google/uuid"
)

// Comment represents a user's comment on a dictionary item
type Comment struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID     uuid.UUID `gorm:"type:uuid;index;not null"`
	TargetType string    `gorm:"type:varchar(20);not null"` // "meaning" or "translation"
	TargetID   uuid.UUID `gorm:"type:uuid;index;not null"`
	Content    string    `gorm:"type:text;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time `gorm:"index"`

	// Relations (not stored in database)
	User interface{} `gorm:"-"`
}
