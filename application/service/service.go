package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
)

// EntryService defines operations for dictionary entries
type EntryService interface {
	// Entry operations
	CreateEntry(ctx context.Context, req *request.CreateEntryRequest) (*response.EntryResponse, error)
	GetEntryByID(ctx context.Context, id uuid.UUID) (*response.EntryResponse, error)
	UpdateEntry(ctx context.Context, id uuid.UUID, req *request.UpdateEntryRequest) (*response.EntryResponse, error)
	DeleteEntry(ctx context.Context, id uuid.UUID) error
	ListEntries(ctx context.Context, req *request.ListEntriesRequest) (*response.EntryListResponse, error)

	// Meaning operations
	AddMeaning(ctx context.Context, entryID uuid.UUID, req *request.CreateMeaningRequest) (*response.MeaningResponse, error)
	UpdateMeaning(ctx context.Context, id uuid.UUID, req *request.UpdateMeaningRequest) (*response.MeaningResponse, error)
	DeleteMeaning(ctx context.Context, id uuid.UUID) error
	ListMeanings(ctx context.Context, entryID uuid.UUID) (*response.MeaningListResponse, error)

	// Social operations for meanings
	AddMeaningComment(ctx context.Context, meaningID uuid.UUID, req *request.CreateCommentRequest) (*response.CommentResponse, error)
	ToggleMeaningLike(ctx context.Context, meaningID uuid.UUID, userID uuid.UUID) error
}

// TranslationService defines operations for translations
type TranslationService interface {
	// Translation operations
	CreateTranslation(ctx context.Context, meaningID uuid.UUID, req *request.CreateTranslationRequest) (*response.TranslationResponse, error)
	UpdateTranslation(ctx context.Context, id uuid.UUID, req *request.UpdateTranslationRequest) (*response.TranslationResponse, error)
	DeleteTranslation(ctx context.Context, id uuid.UUID) error
	ListTranslations(ctx context.Context, meaningID uuid.UUID, langID string) (*response.TranslationListResponse, error)

	// Social operations for translations
	AddTranslationComment(ctx context.Context, translationID uuid.UUID, req *request.CreateCommentRequest) (*response.CommentResponse, error)
	ToggleTranslationLike(ctx context.Context, translationID uuid.UUID, userID uuid.UUID) error
}

// UserService defines operations for user management
type UserService interface {
	// User operations
	CreateUser(ctx context.Context, req *request.CreateUserRequest) (*response.UserResponse, error)
	GetUser(ctx context.Context, id uuid.UUID) (*response.UserResponse, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *request.UpdateUserRequest) (*response.UserResponse, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// Authentication
	Authenticate(ctx context.Context, req *request.AuthRequest) (*response.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*response.AuthResponse, error)
  
	// User content operations
	ListUserEntries(ctx context.Context, userID uuid.UUID, req *request.ListEntriesRequest) (*response.EntryListResponse, error)
	ListUserTranslations(ctx context.Context, userID uuid.UUID, req *request.ListTranslationsRequest) (*response.TranslationListResponse, error)
	ListUserComments(ctx context.Context, userID uuid.UUID, req *request.ListCommentsRequest) (*response.CommentListResponse, error)
	ListUserLikes(ctx context.Context, userID uuid.UUID, req *request.ListLikesRequest) (*response.LikeListResponse, error)
}
