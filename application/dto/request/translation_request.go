package request

import "github.com/google/uuid"

// CreateTranslationRequest contains data for creating a new translation
type CreateTranslationRequest struct {
	LanguageID string `json:"language_id" binding:"required,min=2,max=5"` // ISO 639-1 code
	Text       string `json:"text" binding:"required"`
}

// UpdateTranslationRequest contains data for updating an existing translation
type UpdateTranslationRequest struct {
	Text string `json:"text" binding:"required"`
}

// ListTranslationsRequest contains filtering and pagination parameters for translations
type ListTranslationsRequest struct {
	Limit      int    `json:"limit" form:"limit" binding:"omitempty,min=1,max=100"`
	Offset     int    `json:"offset" form:"offset" binding:"omitempty,min=0"`
	SortBy     string `json:"sort_by" form:"sort_by" binding:"omitempty,oneof=created_at updated_at language_id"`
	SortDesc   bool   `json:"sort_desc" form:"sort_desc"`
	LanguageID string `json:"language_id" form:"language_id" binding:"omitempty,min=2,max=5"`
}

// TranslationCommentRequest contains data for adding a comment to a translation
type TranslationCommentRequest struct {
	Content string    `json:"content" binding:"required,min=1,max=500"`
	UserID  uuid.UUID `json:"-"` // Set from authentication context, not from client
}

// TranslationLikeRequest contains data for toggling a like on a translation
type TranslationLikeRequest struct {
	UserID uuid.UUID `json:"-"` // Set from authentication context, not from client
}
