package request

import "time"

// ListCommentsRequest contains parameters for listing user comments
type ListCommentsRequest struct {
    Limit      int       `json:"limit" form:"limit" binding:"omitempty,min=1,max=100"`
    Offset     int       `json:"offset" form:"offset" binding:"omitempty,min=0"`
    SortBy     string    `json:"sort_by" form:"sort_by" binding:"omitempty,oneof=created_at target_type"`
    SortDesc   bool      `json:"sort_desc" form:"sort_desc"`
    TargetType string    `json:"target_type" form:"target_type" binding:"omitempty,oneof=meaning translation"`
    FromDate   time.Time `json:"from_date" form:"from_date"`
    ToDate     time.Time `json:"to_date" form:"to_date"`
}

// ListLikesRequest contains parameters for listing user likes
type ListLikesRequest struct {
    Limit      int       `json:"limit" form:"limit" binding:"omitempty,min=1,max=100"`
    Offset     int       `json:"offset" form:"offset" binding:"omitempty,min=0"`
    SortBy     string    `json:"sort_by" form:"sort_by" binding:"omitempty,oneof=created_at target_type"`
    SortDesc   bool      `json:"sort_desc" form:"sort_desc"`
    TargetType string    `json:"target_type" form:"target_type" binding:"omitempty,oneof=meaning translation"`
    FromDate   time.Time `json:"from_date" form:"from_date"`
    ToDate     time.Time `json:"to_date" form:"to_date"`
}

// UserTranslationsRequest extends the basic listing functionality with user-specific fields
type UserTranslationsRequest struct {
    Limit      int    `json:"limit" form:"limit" binding:"omitempty,min=1,max=100"`
    Offset     int    `json:"offset" form:"offset" binding:"omitempty,min=0"`
    SortBy     string `json:"sort_by" form:"sort_by" binding:"omitempty,oneof=created_at updated_at language_id"`
    SortDesc   bool   `json:"sort_desc" form:"sort_desc"`
    LanguageID string `json:"language_id" form:"language_id" binding:"omitempty,min=2,max=5"`
    TextSearch string `json:"text_search" form:"text_search" binding:"omitempty,min=1"`
}

// CreateUserRequest contains data for creating a new user
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

// UpdateUserRequest contains data for updating an existing user
type UpdateUserRequest struct {
    Username string `json:"username" binding:"omitempty,min=3,max=50"`
    Email    string `json:"email" binding:"omitempty,email"`
    Password string `json:"password" binding:"omitempty,min=8"`
    Avatar   string `json:"avatar"`
}

// AuthRequest contains data for user authentication
type AuthRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest contains data for refreshing an authentication token
type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}
