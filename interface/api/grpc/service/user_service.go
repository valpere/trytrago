package service

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/interface/api/grpc/proto"
)

// UserService implements the gRPC UserService server
type UserService struct {
	proto.UnimplementedUserServiceServer
	service service.UserService
	logger  logging.Logger
}

// NewUserService creates a new UserService instance
func NewUserService(service service.UserService, logger logging.Logger) *UserService {
	return &UserService{
		service: service,
		logger:  logger.With(logging.String("component", "grpc_user_service")),
	}
}

// CreateUser implements proto.UserServiceServer
func (s *UserService) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.UserResponse, error) {
	s.logger.Debug("gRPC CreateUser called", logging.String("username", req.Username))

	// Map proto request to application DTO
	createReq := &request.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	// Call application service
	resp, err := s.service.CreateUser(ctx, createReq)
	if err != nil {
		s.logger.Error("failed to create user", 
			logging.Error(err),
			logging.String("username", req.Username),
		)
		
		if database.IsDuplicateError(err) {
			return nil, status.Errorf(codes.AlreadyExists, "user already exists")
		}
		
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	// Map application response to proto response
	return s.mapUserToProto(resp), nil
}

// GetUser implements proto.UserServiceServer
func (s *UserService) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.UserResponse, error) {
	s.logger.Debug("gRPC GetUser called", logging.String("id", req.Id))

	// Parse UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Call application service
	resp, err := s.service.GetUser(ctx, id)
	if err != nil {
		s.logger.Error("failed to get user", 
			logging.Error(err),
			logging.String("id", req.Id),
		)
		
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// Map application response to proto response
	return s.mapUserToProto(resp), nil
}

// UpdateUser implements proto.UserServiceServer
func (s *UserService) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UserResponse, error) {
	s.logger.Debug("gRPC UpdateUser called", logging.String("id", req.Id))

	// Parse UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Map proto request to application DTO
	updateReq := &request.UpdateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Avatar:   req.Avatar,
	}

	// Call application service
	resp, err := s.service.UpdateUser(ctx, id, updateReq)
	if err != nil {
		s.logger.Error("failed to update user", 
			logging.Error(err),
			logging.String("id", req.Id),
		)
		
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		
		if database.IsDuplicateError(err) {
			return nil, status.Errorf(codes.AlreadyExists, "username or email already exists")
		}
		
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	// Map application response to proto response
	return s.mapUserToProto(resp), nil
}

// DeleteUser implements proto.UserServiceServer
func (s *UserService) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*emptypb.Empty, error) {
	s.logger.Debug("gRPC DeleteUser called", logging.String("id", req.Id))

	// Parse UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Call application service
	err = s.service.DeleteUser(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete user", 
			logging.Error(err),
			logging.String("id", req.Id),
		)
		
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// Login implements proto.UserServiceServer
func (s *UserService) Login(ctx context.Context, req *proto.LoginRequest) (*proto.AuthResponse, error) {
	s.logger.Debug("gRPC Login called", logging.String("username", req.Username))

	// Map proto request to application DTO
	authReq := &request.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}

	// Call application service
	resp, err := s.service.Authenticate(ctx, authReq)
	if err != nil {
		s.logger.Warn("authentication failed", 
			logging.Error(err),
			logging.String("username", req.Username),
		)
		
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	// Map application response to proto response
	return s.mapAuthToProto(resp), nil
}

// RefreshToken implements proto.UserServiceServer
func (s *UserService) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.AuthResponse, error) {
	s.logger.Debug("gRPC RefreshToken called")

	// Call application service
	resp, err := s.service.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Warn("token refresh failed", logging.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token")
	}

	// Map application response to proto response
	return s.mapAuthToProto(resp), nil
}

// ListUserEntries implements proto.UserServiceServer
func (s *UserService) ListUserEntries(ctx context.Context, req *proto.ListUserEntriesRequest) (*proto.ListEntriesResponse, error) {
	s.logger.Debug("gRPC ListUserEntries called", 
		logging.String("userId", req.UserId),
		logging.Int("limit", int(req.Limit)),
	)

	// Parse UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Map proto request to application DTO
	listReq := &request.ListEntriesRequest{
		Limit:      int(req.Limit),
		Offset:     int(req.Offset),
		SortBy:     req.SortBy,
		SortDesc:   req.SortDesc,
		WordFilter: req.WordFilter,
		Type:       req.Type,
	}

	// Call application service
	resp, err := s.service.ListUserEntries(ctx, userID, listReq)
	if err != nil {
		s.logger.Error("failed to list user entries", 
			logging.Error(err),
			logging.String("userId", req.UserId),
		)
		
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		
		return nil, status.Errorf(codes.Internal, "failed to list entries: %v", err)
	}

	// Map application response to proto response
	protoResp := &proto.ListEntriesResponse{
		Total:  int32(resp.Total),
		Limit:  int32(resp.Limit),
		Offset: int32(resp.Offset),
	}

	// Map entries
	protoResp.Entries = make([]*proto.EntryResponse, len(resp.Entries))
	for i, entry := range resp.Entries {
		protoResp.Entries[i] = s.mapEntryToProto(entry)
	}

	return protoResp, nil
}

// ListUserTranslations implements proto.UserServiceServer
func (s *UserService) ListUserTranslations(ctx context.Context, req *proto.ListUserTranslationsRequest) (*proto.ListTranslationsResponse, error) {
	s.logger.Debug("gRPC ListUserTranslations called", 
		logging.String("userId", req.UserId),
		logging.Int("limit", int(req.Limit)),
	)

	// Parse UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Map proto request to application DTO
	listReq := &request.ListTranslationsRequest{
		Limit:      int(req.Limit),
		Offset:     int(req.Offset),
		SortBy:     req.SortBy,
		SortDesc:   req.SortDesc,
		LanguageID: req.LanguageId,
		TextSearch: req.TextSearch,
	}

	// Call application service
	resp, err := s.service.ListUserTranslations(ctx, userID, listReq)
	if err != nil {
		s.logger.Error("failed to list user translations", 
			logging.Error(err),
			logging.String("userId", req.UserId),
		)
		
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		
		return nil, status.Errorf(codes.Internal, "failed to list translations: %v", err)
	}

	// Map application response to proto response
	protoResp := &proto.ListTranslationsResponse{
		Total: int32(resp.Total),
	}

	// Map translations
	protoResp.Translations = make([]*proto.TranslationResponse, len(resp.Translations))
	for i, translation := range resp.Translations {
		protoResp.Translations[i] = s.mapTranslationToProto(translation)
	}

	return protoResp, nil
}

// ListUserComments implements proto.UserServiceServer
func (s *UserService) ListUserComments(ctx context.Context, req *proto.ListUserCommentsRequest) (*proto.ListCommentsResponse, error) {
	s.logger.Debug("gRPC ListUserComments called", 
		logging.String("userId", req.UserId),
		logging.Int("limit", int(req.Limit)),
	)

	// Parse UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Map proto request to application DTO
	listReq := &request.ListCommentsRequest{
		Limit:      int(req.Limit),
		Offset:     int(req.Offset),
		SortBy:     req.SortBy,
		SortDesc:   req.SortDesc,
		TargetType: req.TargetType,
	}

	// Convert timestamp if provided
	if req.FromDate != nil {
		listReq.FromDate = req.FromDate.AsTime()
	}
	if req.ToDate != nil {
		listReq.ToDate = req.ToDate.AsTime()
	}

	// Call application service
	resp, err := s.service.ListUserComments(ctx, userID, listReq)
	if err != nil {
		s.logger.Error("failed to list user comments", 
			logging.Error(err),
			logging.String("userId", req.UserId),
		)
		
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		
		return nil, status.Errorf(codes.Internal, "failed to list comments: %v", err)
	}

	// Map application response to proto response
	protoResp := &proto.ListCommentsResponse{
		Total:  int32(resp.Total),
		Limit:  int32(resp.Limit),
		Offset: int32(resp.Offset),
	}

	// Map comments
	protoResp.Comments = make([]*proto.CommentResponse, len(resp.Comments))
	for i, comment := range resp.Comments {
		protoResp.Comments[i] = s.mapCommentToProto(&comment)
	}

	return protoResp, nil
}

// ListUserLikes implements proto.UserServiceServer
func (s *UserService) ListUserLikes(ctx context.Context, req *proto.ListUserLikesRequest) (*proto.ListLikesResponse, error) {
	s.logger.Debug("gRPC ListUserLikes called", 
		logging.String("userId", req.UserId),
		logging.Int("limit", int(req.Limit)),
	)

	// Parse UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Map proto request to application DTO
	listReq := &request.ListLikesRequest{
		Limit:      int(req.Limit),
		Offset:     int(req.Offset),
		SortBy:     req.SortBy,
		SortDesc:   req.SortDesc,
		TargetType: req.TargetType,
	}

	// Convert timestamp if provided
	if req.FromDate != nil {
		listReq.FromDate = req.FromDate.AsTime()
	}
	if req.ToDate != nil {
		listReq.ToDate = req.ToDate.AsTime()
	}

	// Call application service
	resp, err := s.service.ListUserLikes(ctx, userID, listReq)
	if err != nil {
		s.logger.Error("failed to list user likes", 
			logging.Error(err),
			logging.String("userId", req.UserId),
		)
		
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		
		return nil, status.Errorf(codes.Internal, "failed to list likes: %v", err)
	}

	// Map application response to proto response
	protoResp := &proto.ListLikesResponse{
		Total:  int32(resp.Total),
		Limit:  int32(resp.Limit),
		Offset: int32(resp.Offset),
	}

	// Map likes
	protoResp.Likes = make([]*proto.LikeResponse, len(resp.Likes))
	for i, like := range resp.Likes {
		protoResp.Likes[i] = s.mapLikeToProto(&like)
	}

	return protoResp, nil
}

// Helper function to map a user response to a proto user response
func (s *UserService) mapUserToProto(user *response.UserResponse) *proto.UserResponse {
	if user == nil {
		return nil
	}

	return &proto.UserResponse{
		Id:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

// Helper function to map an auth response to a proto auth response
func (s *UserService) mapAuthToProto(auth *response.AuthResponse) *proto.AuthResponse {
	if auth == nil {
		return nil
	}

	return &proto.AuthResponse{
		AccessToken:  auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		ExpiresIn:    int64(auth.ExpiresIn),
		User:         s.mapUserToProto(auth.User),
	}
}

// Helper function to map an entry response to a proto entry response
func (s *UserService) mapEntryToProto(entry *response.EntryResponse) *proto.EntryResponse {
	if entry == nil {
		return nil
	}

	return &proto.EntryResponse{
		Id:            entry.ID.String(),
		Word:          entry.Word,
		Type:          entry.Type,
		Pronunciation: entry.Pronunciation,
		CreatedAt:     timestamppb.New(entry.CreatedAt),
		UpdatedAt:     timestamppb.New(entry.UpdatedAt),
	}
}

// Helper function to map a translation response to a proto translation response
func (s *UserService) mapTranslationToProto(translation *response.TranslationResponse) *proto.TranslationResponse {
	if translation == nil {
		return nil
	}

	return &proto.TranslationResponse{
		Id:             translation.ID.String(),
		MeaningId:      translation.MeaningID.String(),
		LanguageId:     translation.LanguageID,
		Text:           translation.Text,
		LikesCount:     int32(translation.LikesCount),
		CurrentUserLiked: translation.CurrentUserLiked,
		CreatedAt:      timestamppb.New(translation.CreatedAt),
		UpdatedAt:      timestamppb.New(translation.UpdatedAt),
	}
}

// Helper function to map a comment response to a proto comment response
func (s *UserService) mapCommentToProto(comment *response.CommentResponse) *proto.CommentResponse {
	if comment == nil {
		return nil
	}

	return &proto.CommentResponse{
		Id:        comment.ID.String(),
		Content:   comment.Content,
		User:      s.mapUserSummaryToProto(&comment.User),
		CreatedAt: timestamppb.New(comment.CreatedAt),
		UpdatedAt: timestamppb.New(comment.UpdatedAt),
	}
}

// Helper function to map a like response to a proto like response
func (s *UserService) mapLikeToProto(like *response.LikeResponse) *proto.LikeResponse {
	if like == nil {
		return nil
	}

	return &proto.LikeResponse{
		Id:         like.ID.String(),
		UserId:     like.UserID.String(),
		TargetType: like.TargetType,
		TargetId:   like.TargetID.String(),
		User:       s.mapUserSummaryToProto(&like.User),
		CreatedAt:  timestamppb.New(like.CreatedAt),
	}
}

// Helper function to map a user summary to a proto user summary
func (s *UserService) mapUserSummaryToProto(user *response.UserSummary) *proto.UserSummary {
	if user == nil {
		return nil
	}

	return &proto.UserSummary{
		Id:       user.ID.String(),
		Username: user.Username,
		Avatar:   user.Avatar,
	}
}
