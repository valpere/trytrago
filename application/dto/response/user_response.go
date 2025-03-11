package response

import (
	"time"

	"github.com/google/uuid"
)

// CommentListResponse represents a paginated list of comments
type CommentListResponse struct {
	Comments []CommentResponse `json:"comments"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// LikeListResponse represents a paginated list of likes
type LikeListResponse struct {
	Likes  []LikeResponse `json:"likes"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// LikeResponse represents a single like in API responses
type LikeResponse struct {
	ID         uuid.UUID    `json:"id"`
	UserID     uuid.UUID    `json:"user_id"`
	TargetType string       `json:"target_type"` // "meaning" or "translation"
	TargetID   uuid.UUID    `json:"target_id"`
	User       UserSummary  `json:"user,omitempty"`
	Target     interface{}  `json:"target,omitempty"` // Can be MeaningResponse or TranslationResponse
	CreatedAt  time.Time    `json:"created_at"`
}
