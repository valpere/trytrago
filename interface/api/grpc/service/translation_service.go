package service

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/interface/api/grpc/proto"
)

// The translation-related methods for the DictionaryService are defined in entry_service.go
// This file contains helper methods specifically for translation operations

// CreateTranslation implements the gRPC method
func (s *DictionaryService) handleCreateTranslation(ctx context.Context, req *proto.CreateTranslationRequest) (*proto.TranslationResponse, error) {
	s.logger.Debug("Creating translation",
		logging.String("meaningID", req.MeaningId),
		logging.String("languageID", req.LanguageId),
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
		s.logger.Error("failed to create translation",
			logging.Error(err),
			logging.String("meaningID", req.MeaningId),
		)

		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "meaning not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to create translation: %v", err)
	}

	// Map application response to proto response
	return s.mapTranslationToProto(resp), nil
}

// UpdateTranslation implements the gRPC method
func (s *DictionaryService) handleUpdateTranslation(ctx context.Context, req *proto.UpdateTranslationRequest) (*proto.TranslationResponse, error) {
	s.logger.Debug("Updating translation", logging.String("id", req.Id))

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
		s.logger.Error("failed to update translation",
			logging.Error(err),
			logging.String("id", req.Id),
		)

		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "translation not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to update translation: %v", err)
	}

	// Map application response to proto response
	return s.mapTranslationToProto(resp), nil
}

// DeleteTranslation implements the gRPC method
func (s *DictionaryService) handleDeleteTranslation(ctx context.Context, req *proto.DeleteTranslationRequest) (*emptypb.Empty, error) {
	s.logger.Debug("Deleting translation", logging.String("id", req.Id))

	// Parse UUID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid translation ID: %v", err)
	}

	// Call application service
	err = s.translationService.DeleteTranslation(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete translation",
			logging.Error(err),
			logging.String("id", req.Id),
		)

		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "translation not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to delete translation: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// ListTranslations implements the gRPC method
func (s *DictionaryService) handleListTranslations(ctx context.Context, req *proto.ListTranslationsRequest) (*proto.ListTranslationsResponse, error) {
	s.logger.Debug("Listing translations",
		logging.String("meaningID", req.MeaningId),
		logging.String("languageID", req.LanguageId),
	)

	// Parse UUID
	meaningID, err := uuid.Parse(req.MeaningId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid meaning ID: %v", err)
	}

	// Call application service
	resp, err := s.translationService.ListTranslations(ctx, meaningID, req.LanguageId)
	if err != nil {
		s.logger.Error("failed to list translations",
			logging.Error(err),
			logging.String("meaningID", req.MeaningId),
		)

		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "meaning not found")
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

// AddTranslationComment implements the gRPC method
func (s *DictionaryService) handleAddTranslationComment(ctx context.Context, translationID uuid.UUID, content string, userID uuid.UUID) (*proto.CommentResponse, error) {
	s.logger.Debug("Adding comment to translation",
		logging.String("translationID", translationID.String()),
	)

	// Map to application DTO
	req := &request.CreateCommentRequest{
		Content: content,
		UserID:  userID,
	}

	// Call application service
	resp, err := s.translationService.AddTranslationComment(ctx, translationID, req)
	if err != nil {
		s.logger.Error("failed to add comment to translation",
			logging.Error(err),
			logging.String("translationID", translationID.String()),
		)

		if database.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "translation not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to add comment: %v", err)
	}

	// Map application response to proto response
	return s.mapCommentToProto(resp), nil
}

// ToggleTranslationLike implements the gRPC method
func (s *DictionaryService) handleToggleTranslationLike(ctx context.Context, translationID uuid.UUID, userID uuid.UUID) error {
	s.logger.Debug("Toggling like on translation",
		logging.String("translationID", translationID.String()),
		logging.String("userID", userID.String()),
	)

	// Call application service
	err := s.translationService.ToggleTranslationLike(ctx, translationID, userID)
	if err != nil {
		s.logger.Error("failed to toggle like on translation",
			logging.Error(err),
			logging.String("translationID", translationID.String()),
		)

		if database.IsNotFoundError(err) {
			return status.Errorf(codes.NotFound, "translation not found")
		}

		return status.Errorf(codes.Internal, "failed to toggle like: %v", err)
	}

	return nil
}
