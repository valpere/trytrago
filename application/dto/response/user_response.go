package response

import (
	"time"

	"github.com/google/uuid"
)

// UserResponse represents a user in API responses
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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
	ID         uuid.UUID   `json:"id"`
	UserID     uuid.UUID   `json:"user_id"`
	TargetType string      `json:"target_type"` // "meaning" or "translation"
	TargetID   uuid.UUID   `json:"target_id"`
	User       UserSummary `json:"user,omitempty"`
	Target     interface{} `json:"target,omitempty"` // Can be MeaningResponse or TranslationResponse
	CreatedAt  time.Time   `json:"created_at"`
}

// AuthResponse represents the response for authentication endpoints
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"` // Expiration time in seconds
	User         UserResponse `json:"user"`
}
