package database

import (
	"github.com/google/uuid"
	"time"
)

// EntryType represents the type of dictionary entry
type EntryType string

const (
	WordType         EntryType = "WORD"
	CompoundWordType EntryType = "COMPOUND_WORD"
	PhraseType       EntryType = "PHRASE"
)

// Entry represents a dictionary entry
type Entry struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key"`
	Word          string    `gorm:"index:idx_word;not null"`
	Type          EntryType `gorm:"type:varchar(20);not null"`
	Pronunciation string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Meanings      []Meaning `gorm:"foreignKey:EntryID"`
}

// Meaning represents a specific meaning of a dictionary entry
type Meaning struct {
	ID           uuid.UUID     `gorm:"type:uuid;primary_key"`
	EntryID      uuid.UUID     `gorm:"type:uuid;index"`
	PartOfSpeech string        `gorm:"type:varchar(50)"`
	Description  string        `gorm:"type:text"`
	Examples     []Example     `gorm:"foreignKey:MeaningID"`
	Translations []Translation `gorm:"foreignKey:MeaningID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Example represents usage examples for a meaning
type Example struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	MeaningID uuid.UUID `gorm:"type:uuid;index"`
	Text      string    `gorm:"type:text"`
	Context   string    `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Translation represents a translation of a meaning
type Translation struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key"`
	MeaningID  uuid.UUID `gorm:"type:uuid;index"`
	LanguageID string    `gorm:"type:varchar(5);index"` // ISO 639-1 code
	Text       string    `gorm:"type:text"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ChangeHistory tracks changes to dictionary entries
type ChangeHistory struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	EntryID   uuid.UUID `gorm:"type:uuid;index"`
	Action    string    `gorm:"type:varchar(20)"`
	Data      []byte    `gorm:"type:jsonb"` // PostgreSQL JSONB for storing change details
	UserID    uuid.UUID `gorm:"type:uuid"`
	CreatedAt time.Time
}
