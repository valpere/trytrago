package response

import (
	"time"

	"github.com/google/uuid"
)

// TranslationResponse represents a translation in API responses
type TranslationResponse struct {
	ID             uuid.UUID          `json:"id"`
	MeaningID      uuid.UUID          `json:"meaning_id"`
	LanguageID     string             `json:"language_id"` // ISO 639-1 code
	Text           string             `json:"text"`
	Comments       []CommentResponse  `json:"comments,omitempty"`
	LikesCount     int                `json:"likes_count"`
	CurrentUserLiked bool             `json:"current_user_liked,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	CreatedBy      *UserSummary       `json:"created_by,omitempty"` // Translation creator
}

// TranslationListResponse represents a paginated list of translations
type TranslationListResponse struct {
	Translations []*TranslationResponse `json:"translations"`
	Total        int                    `json:"total"`
	Limit        int                    `json:"limit"`
	Offset       int                    `json:"offset"`
}

// TranslationSummary represents a compact version of translation for embedding in other responses
type TranslationSummary struct {
	ID         uuid.UUID  `json:"id"`
	LanguageID string     `json:"language_id"`
	Text       string     `json:"text"`
	CreatedAt  time.Time  `json:"created_at"`
}

// LanguageInfo provides information about a language
type LanguageInfo struct {
	Code      string `json:"code"`      // ISO 639-1 code
	Name      string `json:"name"`      // English name
	NativeName string `json:"native_name"` // Name in the language itself
	RTL       bool   `json:"rtl"`       // Right-to-left writing
}
