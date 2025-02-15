package domain

import (
	"time"

	"github.com/google/uuid"
)

type EntryType string

const (
	WordType         EntryType = "WORD"
	CompoundWordType EntryType = "COMPOUND_WORD"
	PhraseType       EntryType = "PHRASE"
)

type Entry struct {
	ID            uuid.UUID
	Word          string
	Type          EntryType // Word, CompoundWord, Phrase
	Pronunciation string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Meaning struct {
	ID           uuid.UUID
	EntryID      uuid.UUID
	PartOfSpeech string // noun, verb, adjective, etc.
	Description  string
	Examples     []Example
	Translations []Translation
}

type Example struct {
	ID        uuid.UUID
	MeaningID uuid.UUID
	Text      string
	Context   string // Optional context or usage notes
}

type Translation struct {
	ID         uuid.UUID
	MeaningID  uuid.UUID
	LanguageID string // ISO 639-1 language code
	Text       string
}

type ChangeHistory struct {
	ID        uuid.UUID
	EntryID   uuid.UUID
	Action    string // Created, Updated, Deleted
	Data      []byte // JSON representation of the changes
	CreatedAt time.Time
	UserID    uuid.UUID
}
