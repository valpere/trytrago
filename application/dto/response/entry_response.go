package response

import (
	"time"

	"github.com/google/uuid"
)

// EntryResponse represents a dictionary entry in API responses
type EntryResponse struct {
	ID            uuid.UUID        `json:"id"`
	Word          string           `json:"word"`
	Type          string           `json:"type"`
	Pronunciation string           `json:"pronunciation,omitempty"`
	Meanings      []MeaningResponse `json:"meanings,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// EntryListResponse represents a paginated list of dictionary entries
type EntryListResponse struct {
	Entries []*EntryResponse `json:"entries"`
	Total   int              `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// MeaningResponse represents a meaning in API responses
type MeaningResponse struct {
	ID             uuid.UUID             `json:"id"`
	EntryID        uuid.UUID             `json:"entry_id"`
	PartOfSpeech   string                `json:"part_of_speech"`
	Description    string                `json:"description"`
	Examples       []ExampleResponse      `json:"examples,omitempty"`
	Translations   []TranslationResponse  `json:"translations,omitempty"`
	Comments       []CommentResponse      `json:"comments,omitempty"`
	LikesCount     int                   `json:"likes_count"`
	CurrentUserLiked bool                `json:"current_user_liked,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

// MeaningListResponse represents a list of meanings
type MeaningListResponse struct {
	Meanings []*MeaningResponse `json:"meanings"`
	Total    int                `json:"total"`
}

// ExampleResponse represents a usage example in API responses
type ExampleResponse struct {
	ID        uuid.UUID `json:"id"`
	Text      string    `json:"text"`
	Context   string    `json:"context,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CommentResponse represents a comment in API responses
type CommentResponse struct {
	ID        uuid.UUID    `json:"id"`
	Content   string       `json:"content"`
	User      UserSummary  `json:"user"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// UserSummary represents minimal user information for embedding in other responses
type UserSummary struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Avatar   string    `json:"avatar,omitempty"`
}

// FeedResponse represents a user's activity feed
type FeedResponse struct {
	Items      []FeedItem `json:"items"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

// FeedItem represents an item in a user's feed
type FeedItem struct {
	ID        uuid.UUID    `json:"id"`
	Type      string       `json:"type"` // "entry", "meaning", "translation", "comment", "like"
	Content   interface{}  `json:"content"`
	User      UserSummary  `json:"user"`
	Timestamp time.Time    `json:"timestamp"`
}
