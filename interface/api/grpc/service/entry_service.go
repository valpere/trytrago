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

// DictionaryService implements the gRPC DictionaryService server
type DictionaryService struct {
	proto.UnimplementedDictionaryServiceServer
	entryService       service.EntryService
	translationService service.TranslationService
	logger             logging.Logger
}

// NewDictionaryService creates a new DictionaryService instance
func NewDictionaryService(
	entryService service.EntryService,
	translationService service.TranslationService,
	logger logging.Logger,
) *DictionaryService {
	return &DictionaryService{
		entryService:       entryService,
		translationService: translationService,
		logger:             logger.With(logging.String("component", "grpc_dictionary_service")),
	}
}

// CreateEntry implements proto.DictionaryServiceServer
func (s *DictionaryService) CreateEntry(ctx context.Context, req *proto.CreateEntryRequest) (*proto.EntryResponse, error) {
	s.logger.Debug("gRPC CreateEntry called", logging.String("word", req.Word))

	// Map proto request to application DTO
	createReq := &request.CreateEntryRequest{
		Word:          req.Word,
		Type:          req.Type,
		Pronunciation: req.Pronunciation,
	}

	// Call application service
	resp, err := s.entryService.CreateEntry(ctx, createReq)
	if err != nil {
		s.logger.Error("failed to create entry",
			logging.Error(err),
			logging.String("word", req.Word),
		)

		if database.IsDuplicateError(err) {
			return nil, status.Errorf(codes.AlreadyExists, "entry already exists")
		}

		return nil, status.Errorf(codes.Internal, "failed to create entry: %v", err)
	}

	// Map application response to proto response
	return s.mapEntryToProto(resp), nil
}

// GetEntry implements proto.DictionaryServiceServer
func (s *DictionaryService) GetEntry(ctx context.Context, req *proto.GetEntryRequest) (*proto.EntryResponse, error) {
	s.logger.Debug("gRPC GetEntry called", logging.String("id", req.Id))

	// Parse UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entry ID: %v", err)
	}

	// Call application service
	resp, err := s.entryService.GetEntryByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get entry",
			logging.Error(err),
			logging.String("id", req.Id),
		)

		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "entry not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to get entry: %v", err)
	}

	// Map application response to proto response
	return s.mapEntryToProto(resp), nil
}

// UpdateEntry implements proto.DictionaryServiceServer
func (s *DictionaryService) UpdateEntry(ctx context.Context, req *proto.UpdateEntryRequest) (*proto.EntryResponse, error) {
	s.logger.Debug("gRPC UpdateEntry called", logging.String("id", req.Id))

	// Parse UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entry ID: %v", err)
	}

	// Map proto request to application DTO
	updateReq := &request.UpdateEntryRequest{
		Word:          req.Word,
		Type:          req.Type,
		Pronunciation: req.Pronunciation,
	}

	// Call application service
	resp, err := s.entryService.UpdateEntry(ctx, id, updateReq)
	if err != nil {
		s.logger.Error("failed to update entry",
			logging.Error(err),
			logging.String("id", req.Id),
		)

		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "entry not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to update entry: %v", err)
	}

	// Map application response to proto response
	return s.mapEntryToProto(resp), nil
}

// DeleteEntry implements proto.DictionaryServiceServer
func (s *DictionaryService) DeleteEntry(ctx context.Context, req *proto.DeleteEntryRequest) (*emptypb.Empty, error) {
	s.logger.Debug("gRPC DeleteEntry called", logging.String("id", req.Id))

	// Parse UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entry ID: %v", err)
	}

	// Call application service
	err = s.entryService.DeleteEntry(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete entry",
			logging.Error(err),
			logging.String("id", req.Id),
		)

		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "entry not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to delete entry: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// ListEntries implements proto.DictionaryServiceServer
func (s *DictionaryService) ListEntries(ctx context.Context, req *proto.ListEntriesRequest) (*proto.ListEntriesResponse, error) {
	s.logger.Debug("gRPC ListEntries called",
		logging.Int("limit", int(req.Limit)),
		logging.Int("offset", int(req.Offset)),
	)

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
	resp, err := s.entryService.ListEntries(ctx, listReq)
	if err != nil {
		s.logger.Error("failed to list entries", logging.Error(err))
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

// Helper function to map an entry response to a proto entry response
func (s *DictionaryService) mapEntryToProto(entry *response.EntryResponse) *proto.EntryResponse {
	if entry == nil {
		return nil
	}

	protoEntry := &proto.EntryResponse{
		Id:            entry.ID.String(),
		Word:          entry.Word,
		Type:          entry.Type,
		Pronunciation: entry.Pronunciation,
		CreatedAt:     timestamppb.New(entry.CreatedAt),
		UpdatedAt:     timestamppb.New(entry.UpdatedAt),
	}

	// Map meanings if available
	if len(entry.Meanings) > 0 {
		protoEntry.Meanings = make([]*proto.MeaningResponse, len(entry.Meanings))
		for i, meaning := range entry.Meanings {
			protoEntry.Meanings[i] = s.mapMeaningToProto(&meaning)
		}
	}

	return protoEntry
}

// Helper function to map a meaning response to a proto meaning response
func (s *DictionaryService) mapMeaningToProto(meaning *response.MeaningResponse) *proto.MeaningResponse {
	if meaning == nil {
		return nil
	}

	protoMeaning := &proto.MeaningResponse{
		Id:               meaning.ID.String(),
		EntryId:          meaning.EntryID.String(),
		PartOfSpeech:     meaning.PartOfSpeech,
		Description:      meaning.Description,
		LikesCount:       int32(meaning.LikesCount),
		CurrentUserLiked: meaning.CurrentUserLiked,
		CreatedAt:        timestamppb.New(meaning.CreatedAt),
		UpdatedAt:        timestamppb.New(meaning.UpdatedAt),
	}

	// Map examples if available
	if len(meaning.Examples) > 0 {
		protoMeaning.Examples = make([]*proto.ExampleResponse, len(meaning.Examples))
		for i, example := range meaning.Examples {
			protoMeaning.Examples[i] = &proto.ExampleResponse{
				Id:        example.ID.String(),
				Text:      example.Text,
				Context:   example.Context,
				CreatedAt: timestamppb.New(example.CreatedAt),
				UpdatedAt: timestamppb.New(example.UpdatedAt),
			}
		}
	}

	// Map translations if available
	if len(meaning.Translations) > 0 {
		protoMeaning.Translations = make([]*proto.TranslationResponse, len(meaning.Translations))
		for i, translation := range meaning.Translations {
			protoMeaning.Translations[i] = s.mapTranslationToProto(&translation)
		}
	}

	// Map comments if available
	if len(meaning.Comments) > 0 {
		protoMeaning.Comments = make([]*proto.CommentResponse, len(meaning.Comments))
		for i, comment := range meaning.Comments {
			protoMeaning.Comments[i] = s.mapCommentToProto(&comment)
		}
	}

	return protoMeaning
}

// Helper function to map a translation response to a proto translation response
func (s *DictionaryService) mapTranslationToProto(translation *response.TranslationResponse) *proto.TranslationResponse {
	if translation == nil {
		return nil
	}

	protoTranslation := &proto.TranslationResponse{
		Id:               translation.ID.String(),
		MeaningId:        translation.MeaningID.String(),
		LanguageId:       translation.LanguageID,
		Text:             translation.Text,
		LikesCount:       int32(translation.LikesCount),
		CurrentUserLiked: translation.CurrentUserLiked,
		CreatedAt:        timestamppb.New(translation.CreatedAt),
		UpdatedAt:        timestamppb.New(translation.UpdatedAt),
	}

	// Map comments if available
	if len(translation.Comments) > 0 {
		protoTranslation.Comments = make([]*proto.CommentResponse, len(translation.Comments))
		for i, comment := range translation.Comments {
			protoTranslation.Comments[i] = s.mapCommentToProto(&comment)
		}
	}

	return protoTranslation
}

// Helper function to map a comment response to a proto comment response
func (s *DictionaryService) mapCommentToProto(comment *response.CommentResponse) *proto.CommentResponse {
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

// Helper function to map a user summary to a proto user summary
func (s *DictionaryService) mapUserSummaryToProto(user *response.UserSummary) *proto.UserSummary {
	if user == nil {
		return nil
	}

	return &proto.UserSummary{
		Id:       user.ID.String(),
		Username: user.Username,
		Avatar:   user.Avatar,
	}
}

// The following methods need to be implemented for the DictionaryService
// They all need to call the corresponding application service methods

// AddMeaning implements proto.DictionaryServiceServer
func (s *DictionaryService) AddMeaning(ctx context.Context, req *proto.AddMeaningRequest) (*proto.MeaningResponse, error) {
	s.logger.Debug("gRPC AddMeaning called",
		logging.String("entryId", req.EntryId),
		logging.String("partOfSpeechId", req.PartOfSpeechId),
	)

	// Parse UUIDs
	entryID, err := uuid.Parse(req.EntryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entry ID: %v", err)
	}

	partOfSpeechID, err := uuid.Parse(req.PartOfSpeechId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid part of speech ID: %v", err)
	}

	// Map proto request to application DTO
	createReq := &request.CreateMeaningRequest{
		PartOfSpeechID: partOfSpeechID,
		Description:    req.Description,
		Examples:       req.Examples,
	}

	// Call application service
	resp, err := s.entryService.AddMeaning(ctx, entryID, createReq)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "entry not found")
		}

		s.logger.Error("failed to add meaning",
			logging.Error(err),
			logging.String("entryId", req.EntryId),
		)
		return nil, status.Errorf(codes.Internal, "failed to add meaning: %v", err)
	}

	// Map application response to proto response
	return s.mapMeaningToProto(resp), nil
}

// UpdateMeaning implements proto.DictionaryServiceServer
func (s *DictionaryService) UpdateMeaning(ctx context.Context, req *proto.UpdateMeaningRequest) (*proto.MeaningResponse, error) {
	s.logger.Debug("gRPC UpdateMeaning called", logging.String("id", req.Id))

	// Parse UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid meaning ID: %v", err)
	}

	var partOfSpeechID uuid.UUID
	if req.PartOfSpeechId != "" {
		partOfSpeechID, err = uuid.Parse(req.PartOfSpeechId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid part of speech ID: %v", err)
		}
	}

	// Map proto request to application DTO
	updateReq := &request.UpdateMeaningRequest{
		PartOfSpeechID: partOfSpeechID,
		Description:    req.Description,
		Examples:       req.Examples,
	}

	// Call application service
	resp, err := s.entryService.UpdateMeaning(ctx, id, updateReq)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "meaning not found")
		}

		s.logger.Error("failed to update meaning",
			logging.Error(err),
			logging.String("id", req.Id),
		)
		return nil, status.Errorf(codes.Internal, "failed to update meaning: %v", err)
	}

	// Map application response to proto response
	return s.mapMeaningToProto(resp), nil
}

// DeleteMeaning implements proto.DictionaryServiceServer
func (s *DictionaryService) DeleteMeaning(ctx context.Context, req *proto.DeleteMeaningRequest) (*emptypb.Empty, error) {
	s.logger.Debug("gRPC DeleteMeaning called", logging.String("id", req.Id))

	// Parse UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid meaning ID: %v", err)
	}

	// Call application service
	err = s.entryService.DeleteMeaning(ctx, id)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "meaning not found")
		}

		s.logger.Error("failed to delete meaning",
			logging.Error(err),
			logging.String("id", req.Id),
		)
		return nil, status.Errorf(codes.Internal, "failed to delete meaning: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// ListMeanings implements proto.DictionaryServiceServer
func (s *DictionaryService) ListMeanings(ctx context.Context, req *proto.ListMeaningsRequest) (*proto.ListMeaningsResponse, error) {
	s.logger.Debug("gRPC ListMeanings called", logging.String("entryId", req.EntryId))

	// Parse UUID
	entryID, err := uuid.Parse(req.EntryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entry ID: %v", err)
	}

	// Call application service
	resp, err := s.entryService.ListMeanings(ctx, entryID)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "entry not found")
		}

		s.logger.Error("failed to list meanings",
			logging.Error(err),
			logging.String("entryId", req.EntryId),
		)
		return nil, status.Errorf(codes.Internal, "failed to list meanings: %v", err)
	}

	// Map application response to proto response
	protoResp := &proto.ListMeaningsResponse{
		Total: int32(resp.Total),
	}

	// Map meanings
	protoResp.Meanings = make([]*proto.MeaningResponse, len(resp.Meanings))
	for i, meaning := range resp.Meanings {
		protoResp.Meanings[i] = s.mapMeaningToProto(meaning)
	}

	return protoResp, nil
}

// CreateTranslation implements proto.DictionaryServiceServer
func (s *DictionaryService) CreateTranslation(ctx context.Context, req *proto.CreateTranslationRequest) (*proto.TranslationResponse, error) {
	s.logger.Debug("gRPC CreateTranslation called",
		logging.String("meaningId", req.MeaningId),
		logging.String("languageId", req.LanguageId),
	)

	// Parse UUID
	meaningID, err := uuid.Parse(req.MeaningId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid meaning ID: %v", err)
	}

	// Map proto request to application DTO
	createReq := &request.CreateTranslationRequest{
		LanguageID: req.LanguageId,
		Text:       req.Text,
	}

	// Call application service
	resp, err := s.translationService.CreateTranslation(ctx, meaningID, createReq)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "meaning not found")
		}

		s.logger.Error("failed to create translation",
			logging.Error(err),
			logging.String("meaningId", req.MeaningId),
		)
		return nil, status.Errorf(codes.Internal, "failed to create translation: %v", err)
	}

	// Map application response to proto response
	return s.mapTranslationToProto(resp), nil
}

// UpdateTranslation implements proto.DictionaryServiceServer
func (s *DictionaryService) UpdateTranslation(ctx context.Context, req *proto.UpdateTranslationRequest) (*proto.TranslationResponse, error) {
	s.logger.Debug("gRPC UpdateTranslation called", logging.String("id", req.Id))

	// Parse UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid translation ID: %v", err)
	}

	// Map proto request to application DTO
	updateReq := &request.UpdateTranslationRequest{
		Text: req.Text,
	}

	// Call application service
	resp, err := s.translationService.UpdateTranslation(ctx, id, updateReq)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "translation not found")
		}

		s.logger.Error("failed to update translation",
			logging.Error(err),
			logging.String("id", req.Id),
		)
		return nil, status.Errorf(codes.Internal, "failed to update translation: %v", err)
	}

	// Map application response to proto response
	return s.mapTranslationToProto(resp), nil
}

// DeleteTranslation implements proto.DictionaryServiceServer
func (s *DictionaryService) DeleteTranslation(ctx context.Context, req *proto.DeleteTranslationRequest) (*emptypb.Empty, error) {
	s.logger.Debug("gRPC DeleteTranslation called", logging.String("id", req.Id))

	// Parse UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid translation ID: %v", err)
	}

	// Call application service
	err = s.translationService.DeleteTranslation(ctx, id)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "translation not found")
		}

		s.logger.Error("failed to delete translation",
			logging.Error(err),
			logging.String("id", req.Id),
		)
		return nil, status.Errorf(codes.Internal, "failed to delete translation: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// ListTranslations implements proto.DictionaryServiceServer
func (s *DictionaryService) ListTranslations(ctx context.Context, req *proto.ListTranslationsRequest) (*proto.ListTranslationsResponse, error) {
	s.logger.Debug("gRPC ListTranslations called",
		logging.String("meaningId", req.MeaningId),
		logging.String("languageId", req.LanguageId),
	)

	// Parse UUID
	meaningID, err := uuid.Parse(req.MeaningId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid meaning ID: %v", err)
	}

	// Call application service
	resp, err := s.translationService.ListTranslations(ctx, meaningID, req.LanguageId)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "meaning not found")
		}

		s.logger.Error("failed to list translations",
			logging.Error(err),
			logging.String("meaningId", req.MeaningId),
		)
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

// AddComment implements proto.DictionaryServiceServer
func (s *DictionaryService) AddComment(ctx context.Context, req *proto.AddCommentRequest) (*proto.CommentResponse, error) {
	s.logger.Debug("gRPC AddComment called",
		logging.String("targetId", req.TargetId),
		logging.String("targetType", req.TargetType),
	)

	// Parse target UUID
	targetID, err := uuid.Parse(req.TargetId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target ID: %v", err)
	}

	// Get user ID from context (set by auth interceptor in a real implementation)
	// For this implementation, we'll create a mock user ID
	userID := uuid.New()

	// Map proto request to application DTO
	commentReq := &request.CreateCommentRequest{
		Content: req.Content,
		UserID:  userID,
	}

	var resp *response.CommentResponse

	// Call appropriate service based on target type
	switch req.TargetType {
	case "meaning":
		resp, err = s.entryService.AddMeaningComment(ctx, targetID, commentReq)
	case "translation":
		resp, err = s.translationService.AddTranslationComment(ctx, targetID, commentReq)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid target type: must be 'meaning' or 'translation'")
	}

	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "target not found")
		}

		s.logger.Error("failed to add comment",
			logging.Error(err),
			logging.String("targetId", req.TargetId),
			logging.String("targetType", req.TargetType),
		)
		return nil, status.Errorf(codes.Internal, "failed to add comment: %v", err)
	}

	// Map application response to proto response
	return s.mapCommentToProto(resp), nil
}

// ToggleLike implements proto.DictionaryServiceServer
func (s *DictionaryService) ToggleLike(ctx context.Context, req *proto.ToggleLikeRequest) (*emptypb.Empty, error) {
	s.logger.Debug("gRPC ToggleLike called",
		logging.String("targetId", req.TargetId),
		logging.String("targetType", req.TargetType),
	)

	// Parse target UUID
	targetID, err := uuid.Parse(req.TargetId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target ID: %v", err)
	}

	// Get user ID from context (set by auth interceptor in a real implementation)
	// For this implementation, we'll create a mock user ID
	userID := uuid.New()

	// Call appropriate service based on target type
	switch req.TargetType {
	case "meaning":
		err = s.entryService.ToggleMeaningLike(ctx, targetID, userID)
	case "translation":
		err = s.translationService.ToggleTranslationLike(ctx, targetID, userID)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid target type: must be 'meaning' or 'translation'")
	}

	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "target not found")
		}

		s.logger.Error("failed to toggle like",
			logging.Error(err),
			logging.String("targetId", req.TargetId),
			logging.String("targetType", req.TargetType),
		)
		return nil, status.Errorf(codes.Internal, "failed to toggle like: %v", err)
	}

	return &emptypb.Empty{}, nil
}
