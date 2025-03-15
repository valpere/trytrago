package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/application/mapper"
	"github.com/valpere/trytrago/domain/database/repository"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/infrastructure/auth"
	"golang.org/x/crypto/bcrypt"
)

// userService implements the UserService interface
type userService struct {
	repo   repository.Repository
	logger logging.Logger
}

// NewUserService creates a new instance of UserService
func NewUserService(repo repository.Repository, logger logging.Logger) UserService {
	return &userService{
		repo:   repo,
		logger: logger.With(logging.String("service", "user")),
	}
}

// CreateUser implements UserService.CreateUser
func (s *userService) CreateUser(ctx context.Context, req *request.CreateUserRequest) (*response.UserResponse, error) {
	s.logger.Debug("creating user", logging.String("username", req.Username))

	// Create domain model from request
	user := mapper.CreateUserRequestToModel(req)
	user.ID = uuid.New()
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", logging.Error(err))
		return nil, fmt.Errorf("failed to process user data: %w", err)
	}
	user.Password = string(hashedPassword)

	// Persist to database
	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.logger.Error("failed to create user", logging.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Map domain model to response DTO
	resp := mapper.UserToResponse(user)
	return resp, nil
}

// GetUser implements UserService.GetUser
func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*response.UserResponse, error) {
	s.logger.Debug("getting user by ID", logging.String("id", id.String()))

	// Fetch user from repository
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get user", logging.Error(err), logging.String("id", id.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Map domain model to response DTO
	resp := mapper.UserToResponse(user)
	return resp, nil
}

// UpdateUser implements UserService.UpdateUser
func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, req *request.UpdateUserRequest) (*response.UserResponse, error) {
	s.logger.Debug("updating user", logging.String("id", id.String()))

	// Fetch user from repository
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get user for update", logging.Error(err), logging.String("id", id.String()))
		return nil, fmt.Errorf("failed to get user for update: %w", err)
	}

	// Apply updates
	mapper.UpdateUserRequestToModel(user, req)
	user.UpdatedAt = time.Now().UTC()

	// If password was updated, hash it
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("failed to hash password", logging.Error(err))
			return nil, fmt.Errorf("failed to process user data: %w", err)
		}
		user.Password = string(hashedPassword)
	}

	// Persist to database
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		s.logger.Error("failed to update user", logging.Error(err), logging.String("id", id.String()))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Map domain model to response DTO
	resp := mapper.UserToResponse(user)
	return resp, nil
}

// DeleteUser implements UserService.DeleteUser
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("deleting user", logging.String("id", id.String()))

	if err := s.repo.DeleteUser(ctx, id); err != nil {
		s.logger.Error("failed to delete user", logging.Error(err), logging.String("id", id.String()))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// Authenticate implements UserService.Authenticate
func (s *userService) Authenticate(ctx context.Context, req *request.AuthRequest) (*response.AuthResponse, error) {
	s.logger.Debug("authenticating user", logging.String("username", req.Username))

	// Fetch user by username
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Warn("authentication failed", logging.Error(err), logging.String("username", req.Username))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		s.logger.Warn("authentication failed", logging.Error(err), logging.String("username", req.Username))
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	accessToken, err := auth.GenerateToken(user.ID.String(), user.Username, string(user.Role))
	if err != nil {
		s.logger.Error("failed to generate access token", logging.Error(err))
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Generate refresh token
	refreshToken, err := auth.GenerateRefreshToken(user.ID.String())
	if err != nil {
		s.logger.Error("failed to generate refresh token", logging.Error(err))
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Update last login
	user.LastLogin = &time.Time{}
	*user.LastLogin = time.Now().UTC()

	// Create response
	userResp := mapper.UserToResponse(user)
	resp := &response.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour in seconds
		User:         *userResp,
	}

	return resp, nil
}

// RefreshToken implements UserService.RefreshToken
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*response.AuthResponse, error) {
	s.logger.Debug("refreshing token")

	// Validate refresh token
	userID, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logger.Warn("invalid refresh token", logging.Error(err))
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Parse UUID from string
	id, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Error("invalid user ID in token", logging.Error(err))
		return nil, fmt.Errorf("invalid token")
	}

	// Fetch user by ID
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get user for token refresh", logging.Error(err), logging.String("userId", id.String()))
		return nil, fmt.Errorf("invalid token")
	}

	// Generate new JWT token
	accessToken, err := auth.GenerateToken(user.ID.String(), user.Username, string(user.Role))
	if err != nil {
		s.logger.Error("failed to generate access token", logging.Error(err))
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, err := auth.GenerateRefreshToken(user.ID.String())
	if err != nil {
		s.logger.Error("failed to generate refresh token", logging.Error(err))
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	// Create response
	userResp := mapper.UserToResponse(user)
	resp := &response.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    3600, // 1 hour in seconds
		User:         *userResp,
	}

	return resp, nil
}

// ListUserEntries implements UserService.ListUserEntries
func (s *userService) ListUserEntries(ctx context.Context, userID uuid.UUID, req *request.ListEntriesRequest) (*response.EntryListResponse, error) {
	s.logger.Debug("listing user entries",
		logging.String("userId", userID.String()),
		logging.Int("limit", req.Limit),
		logging.Int("offset", req.Offset),
	)

	// Prepare repository query parameters
	params := repository.ListParams{
		Offset:   req.Offset,
		Limit:    req.Limit,
		SortBy:   req.SortBy,
		SortDesc: req.SortDesc,
		Filters:  make(map[string]interface{}),
	}

	// Add user ID filter
	// This assumes the entries table has a created_by_id column
	params.Filters["created_by_id = ?"] = userID

	// Add additional filters
	if req.WordFilter != "" {
		params.Filters["word LIKE ?"] = "%" + req.WordFilter + "%"
	}

	if req.Type != "" {
		params.Filters["type = ?"] = req.Type
	}

	// Execute query
	entries, err := s.repo.ListUserEntries(ctx, userID, params)
	if err != nil {
		s.logger.Error("failed to list user entries",
			logging.Error(err),
			logging.String("userId", userID.String()),
		)
		return nil, fmt.Errorf("failed to list user entries: %w", err)
	}

	// Map domain models to response DTOs
	resp := &response.EntryListResponse{
		Entries: make([]*response.EntryResponse, len(entries)),
		Total:   len(entries),
		Limit:   req.Limit,
		Offset:  req.Offset,
	}

	for i, entry := range entries {
		resp.Entries[i] = mapper.EntryToResponse(&entry)
	}

	return resp, nil
}

// ListUserTranslations implements UserService.ListUserTranslations
func (s *userService) ListUserTranslations(ctx context.Context, userID uuid.UUID, req *request.ListTranslationsRequest) (*response.TranslationListResponse, error) {
	s.logger.Debug("listing user translations",
		logging.String("userId", userID.String()),
		logging.Int("limit", req.Limit),
		logging.Int("offset", req.Offset),
	)

	// In a real implementation, we would query the database for translations created by this user
	// For now, return an empty list as a placeholder

	resp := &response.TranslationListResponse{
		Translations: []*response.TranslationResponse{},
		Total:        0,
		Limit:        req.Limit,
		Offset:       req.Offset,
	}

	return resp, nil
}

// ListUserComments implements UserService.ListUserComments
func (s *userService) ListUserComments(ctx context.Context, userID uuid.UUID, req *request.ListCommentsRequest) (*response.CommentListResponse, error) {
	s.logger.Debug("listing user comments",
		logging.String("userId", userID.String()),
		logging.Int("limit", req.Limit),
		logging.Int("offset", req.Offset),
	)

	// In a real implementation, we would query the database for comments created by this user
	// For now, return an empty list as a placeholder

	resp := &response.CommentListResponse{
		Comments: []response.CommentResponse{},
		Total:    0,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}

	return resp, nil
}

// ListUserLikes implements UserService.ListUserLikes
func (s *userService) ListUserLikes(ctx context.Context, userID uuid.UUID, req *request.ListLikesRequest) (*response.LikeListResponse, error) {
	s.logger.Debug("listing user likes",
		logging.String("userId", userID.String()),
		logging.Int("limit", req.Limit),
		logging.Int("offset", req.Offset),
	)

	// In a real implementation, we would query the database for likes created by this user
	// For now, return an empty list as a placeholder

	resp := &response.LikeListResponse{
		Likes:  []response.LikeResponse{},
		Total:  0,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	return resp, nil
}
