package request

import "github.com/google/uuid"

// CreateEntryRequest contains data for creating a new dictionary entry
type CreateEntryRequest struct {
	Word          string `json:"word" binding:"required"`
	Type          string `json:"type" binding:"required,oneof=WORD COMPOUND_WORD PHRASE"`
	Pronunciation string `json:"pronunciation"`
}

// UpdateEntryRequest contains data for updating an existing dictionary entry
type UpdateEntryRequest struct {
	Word          string `json:"word"`
	Type          string `json:"type" binding:"omitempty,oneof=WORD COMPOUND_WORD PHRASE"`
	Pronunciation string `json:"pronunciation"`
}

// ListEntriesRequest contains filtering and pagination parameters
type ListEntriesRequest struct {
	Limit      int    `json:"limit" form:"limit" binding:"omitempty,min=1,max=100"`
	Offset     int    `json:"offset" form:"offset" binding:"omitempty,min=0"`
	SortBy     string `json:"sort_by" form:"sort_by" binding:"omitempty,oneof=word created_at updated_at"`
	SortDesc   bool   `json:"sort_desc" form:"sort_desc"`
	WordFilter string `json:"word_filter" form:"word_filter"`
	Type       string `json:"type" form:"type" binding:"omitempty,oneof=WORD COMPOUND_WORD PHRASE"`
}

// CreateMeaningRequest contains data for adding a new meaning to an entry
type CreateMeaningRequest struct {
	PartOfSpeechID uuid.UUID `json:"part_of_speech_id" binding:"required"`
	Description    string    `json:"description" binding:"required"`
	Examples       []string  `json:"examples"`
}

// UpdateMeaningRequest contains data for updating a meaning
type UpdateMeaningRequest struct {
	PartOfSpeechID uuid.UUID `json:"part_of_speech_id"`
	Description    string    `json:"description"`
	Examples       []string  `json:"examples"`
}

// CreateCommentRequest contains data for creating a comment
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=500"`
}

// FeedRequest contains parameters for retrieving a user's feed
type FeedRequest struct {
	Limit  int    `json:"limit" form:"limit" binding:"omitempty,min=1,max=50"`
	Cursor string `json:"cursor" form:"cursor"` // For cursor-based pagination
}
