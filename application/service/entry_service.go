package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/application/mapper"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/database/repository"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/domain/model"
)

// entryService implements the EntryService interface
type entryService struct {
	repo   repository.Repository
	logger logging.Logger
}

// NewEntryService creates a new instance of EntryService
func NewEntryService(repo repository.Repository, logger logging.Logger) EntryService {
	return &entryService{
		repo:   repo,
		logger: logger.With(logging.String("service", "entry")),
	}
}

// CreateEntry implements EntryService.CreateEntry
func (s *entryService) CreateEntry(ctx context.Context, req *request.CreateEntryRequest) (*response.EntryResponse, error) {
	s.logger.Debug("creating entry", logging.String("word", req.Word))

	// Create domain model from request
	entry := &database.Entry{
		ID:            uuid.New(),
		Word:          req.Word,
		Type:          database.EntryType(req.Type),
		Pronunciation: req.Pronunciation,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Persist to database
	if err := s.repo.CreateEntry(ctx, entry); err != nil {
		s.logger.Error("failed to create entry", logging.Error(err))
		return nil, fmt.Errorf("failed to create entry: %w", err)
	}

	// Map domain model to response DTO
	resp := mapper.EntryToResponse(entry)
	return resp, nil
}

// GetEntryByID implements EntryService.GetEntryByID
func (s *entryService) GetEntryByID(ctx context.Context, id uuid.UUID) (*response.EntryResponse, error) {
	s.logger.Debug("getting entry by ID", logging.String("id", id.String()))

	// Fetch entry from repository
	entry, err := s.repo.GetEntryByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get entry", logging.Error(err), logging.String("id", id.String()))
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	// Map domain model to response DTO
	resp := mapper.EntryToResponse(entry)
	return resp, nil
}

// UpdateEntry implements EntryService.UpdateEntry
func (s *entryService) UpdateEntry(ctx context.Context, id uuid.UUID, req *request.UpdateEntryRequest) (*response.EntryResponse, error) {
	s.logger.Debug("updating entry", logging.String("id", id.String()))

	// Fetch entry from repository
	entry, err := s.repo.GetEntryByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get entry for update", logging.Error(err), logging.String("id", id.String()))
		return nil, fmt.Errorf("failed to get entry for update: %w", err)
	}

	// Update fields
	if req.Word != "" {
		entry.Word = req.Word
	}

	if req.Type != "" {
		entry.Type = database.EntryType(req.Type)
	}

	if req.Pronunciation != "" {
		entry.Pronunciation = req.Pronunciation
	}

	entry.UpdatedAt = time.Now().UTC()

	// Save changes
	if err := s.repo.UpdateEntry(ctx, entry); err != nil {
		s.logger.Error("failed to update entry", logging.Error(err), logging.String("id", id.String()))
		return nil, fmt.Errorf("failed to update entry: %w", err)
	}

	// Map domain model to response DTO
	resp := mapper.EntryToResponse(entry)
	return resp, nil
}

// DeleteEntry implements EntryService.DeleteEntry
func (s *entryService) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("deleting entry", logging.String("id", id.String()))

	if err := s.repo.DeleteEntry(ctx, id); err != nil {
		s.logger.Error("failed to delete entry", logging.Error(err), logging.String("id", id.String()))
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	return nil
}

// ListEntries implements EntryService.ListEntries
func (s *entryService) ListEntries(ctx context.Context, req *request.ListEntriesRequest) (*response.EntryListResponse, error) {
	s.logger.Debug("listing entries",
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

	// Add filters if specified
	if req.WordFilter != "" {
		params.Filters["word LIKE ?"] = "%" + req.WordFilter + "%"
	}

	if req.Type != "" {
		params.Filters["type = ?"] = req.Type
	}

	// Execute query
	entries, err := s.repo.ListEntries(ctx, params)
	if err != nil {
		s.logger.Error("failed to list entries", logging.Error(err))
		return nil, fmt.Errorf("failed to list entries: %w", err)
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

// AddMeaning implements EntryService.AddMeaning
func (s *entryService) AddMeaning(ctx context.Context, entryID uuid.UUID, req *request.CreateMeaningRequest) (*response.MeaningResponse, error) {
	s.logger.Debug("adding meaning to entry",
		logging.String("entryID", entryID.String()),
		logging.String("partOfSpeech", req.PartOfSpeechID.String()),
	)

	// Fetch the entry to ensure it exists
	entry, err := s.repo.GetEntryByID(ctx, entryID)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, database.ErrEntryNotFound
		}
		s.logger.Error("failed to get entry for adding meaning",
			logging.Error(err),
			logging.String("entryID", entryID.String()),
		)
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	// Create a new meaning
	now := time.Now().UTC()
	meaning := database.Meaning{
		ID:             uuid.New(),
		EntryID:        entryID,
		PartOfSpeechId: req.PartOfSpeechID,
		Description:    req.Description,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Add examples if provided
	if len(req.Examples) > 0 {
		meaning.Examples = make([]database.Example, len(req.Examples))
		for i, exampleText := range req.Examples {
			meaning.Examples[i] = database.Example{
				ID:        uuid.New(),
				MeaningID: meaning.ID,
				Text:      exampleText,
				CreatedAt: now,
				UpdatedAt: now,
			}
		}
	}

	// Add the meaning to the entry
	entry.Meanings = append(entry.Meanings, meaning)

	// Update the entry with the new meaning
	if err := s.repo.UpdateEntry(ctx, entry); err != nil {
		s.logger.Error("failed to update entry with new meaning",
			logging.Error(err),
			logging.String("entryID", entryID.String()),
		)
		return nil, fmt.Errorf("failed to save meaning: %w", err)
	}

	// Retrieve the updated entry to ensure we have the proper data
	updatedEntry, err := s.repo.GetEntryByID(ctx, entryID)
	if err != nil {
		s.logger.Error("failed to retrieve updated entry",
			logging.Error(err),
			logging.String("entryID", entryID.String()),
		)
		return nil, fmt.Errorf("failed to retrieve updated entry: %w", err)
	}

	// Find the new meaning in the updated entry
	var newMeaning *database.Meaning
	for i := range updatedEntry.Meanings {
		if updatedEntry.Meanings[i].ID == meaning.ID {
			newMeaning = &updatedEntry.Meanings[i]
			break
		}
	}

	if newMeaning == nil {
		s.logger.Error("newly added meaning not found in updated entry",
			logging.String("meaningID", meaning.ID.String()),
		)
		return nil, fmt.Errorf("newly added meaning not found in updated entry")
	}

	// Map to response
	resp := mapper.MeaningToResponse(newMeaning)
	return resp, nil
}

// UpdateMeaning implements EntryService.UpdateMeaning
func (s *entryService) UpdateMeaning(ctx context.Context, id uuid.UUID, req *request.UpdateMeaningRequest) (*response.MeaningResponse, error) {
	s.logger.Debug("updating meaning", logging.String("meaningID", id.String()))

	// Find the meaning by ID
	// Note: This approach searches across all entries to find the meaning
	// In a real implementation, you would have a direct repository method to get a meaning by ID
	params := repository.ListParams{
		Limit: 100, // Reasonable limit to search through entries
	}
	entries, err := s.repo.ListEntries(ctx, params)
	if err != nil {
		s.logger.Error("failed to list entries to find meaning",
			logging.Error(err),
			logging.String("meaningID", id.String()),
		)
		return nil, fmt.Errorf("failed to find meaning: %w", err)
	}

	// Find the meaning and its parent entry
	var foundMeaning *database.Meaning
	var foundEntry *database.Entry

	for i := range entries {
		entry := &entries[i]
		for j := range entry.Meanings {
			meaning := &entry.Meanings[j]
			if meaning.ID == id {
				foundMeaning = meaning
				foundEntry = entry
				break
			}
		}
		if foundMeaning != nil {
			break
		}
	}

	if foundMeaning == nil {
		return nil, database.ErrEntryNotFound
	}

	// Update meaning fields
	if req.PartOfSpeechID != uuid.Nil {
		foundMeaning.PartOfSpeechId = req.PartOfSpeechID
	}

	if req.Description != "" {
		foundMeaning.Description = req.Description
	}

	foundMeaning.UpdatedAt = time.Now().UTC()

	// Handle examples if provided
	if len(req.Examples) > 0 {
		// For simplicity, we'll replace all examples
		// In a real implementation, you might want to handle more granular updates
		foundMeaning.Examples = make([]database.Example, len(req.Examples))
		for i, exampleText := range req.Examples {
			foundMeaning.Examples[i] = database.Example{
				ID:        uuid.New(), // New example gets a new ID
				MeaningID: foundMeaning.ID,
				Text:      exampleText,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
		}
	}

	// Save the updated entry
	if err := s.repo.UpdateEntry(ctx, foundEntry); err != nil {
		s.logger.Error("failed to update meaning",
			logging.Error(err),
			logging.String("meaningID", id.String()),
		)
		return nil, fmt.Errorf("failed to update meaning: %w", err)
	}

	// Retrieve the updated entry to ensure we have the latest data
	updatedEntry, err := s.repo.GetEntryByID(ctx, foundEntry.ID)
	if err != nil {
		s.logger.Error("failed to retrieve updated entry",
			logging.Error(err),
			logging.String("entryID", foundEntry.ID.String()),
		)
		return nil, fmt.Errorf("failed to retrieve updated entry: %w", err)
	}

	// Find the updated meaning
	var updatedMeaning *database.Meaning
	for i := range updatedEntry.Meanings {
		if updatedEntry.Meanings[i].ID == id {
			updatedMeaning = &updatedEntry.Meanings[i]
			break
		}
	}

	if updatedMeaning == nil {
		s.logger.Error("updated meaning not found in entry",
			logging.String("meaningID", id.String()),
		)
		return nil, fmt.Errorf("updated meaning not found in entry")
	}

	// Map to response
	resp := mapper.MeaningToResponse(updatedMeaning)
	return resp, nil
}

// DeleteMeaning implements EntryService.DeleteMeaning
func (s *entryService) DeleteMeaning(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("deleting meaning", logging.String("meaningID", id.String()))

	// Find the meaning by ID
	params := repository.ListParams{
		Limit: 100, // Reasonable limit to search through entries
	}
	entries, err := s.repo.ListEntries(ctx, params)
	if err != nil {
		s.logger.Error("failed to list entries to find meaning",
			logging.Error(err),
			logging.String("meaningID", id.String()),
		)
		return fmt.Errorf("failed to find meaning: %w", err)
	}

	// Find the meaning and its parent entry
	var foundMeaning *database.Meaning
	var foundEntry *database.Entry
	var meaningIndex int

	for i := range entries {
		entry := &entries[i]
		for j := range entry.Meanings {
			meaning := &entry.Meanings[j]
			if meaning.ID == id {
				foundMeaning = meaning
				foundEntry = entry
				meaningIndex = j
				break
			}
		}
		if foundMeaning != nil {
			break
		}
	}

	if foundMeaning == nil {
		return database.ErrEntryNotFound
	}

	// Remove the meaning from the entry
	foundEntry.Meanings = append(
		foundEntry.Meanings[:meaningIndex],
		foundEntry.Meanings[meaningIndex+1:]...,
	)

	// Save the updated entry
	if err := s.repo.UpdateEntry(ctx, foundEntry); err != nil {
		s.logger.Error("failed to update entry after deleting meaning",
			logging.Error(err),
			logging.String("meaningID", id.String()),
		)
		return fmt.Errorf("failed to delete meaning: %w", err)
	}

	return nil
}

// ListMeanings implements EntryService.ListMeanings
func (s *entryService) ListMeanings(ctx context.Context, entryID uuid.UUID) (*response.MeaningListResponse, error) {
	s.logger.Debug("listing meanings for entry", logging.String("entryID", entryID.String()))

	// Fetch the entry to get its meanings
	entry, err := s.repo.GetEntryByID(ctx, entryID)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, database.ErrEntryNotFound
		}
		s.logger.Error("failed to get entry for listing meanings",
			logging.Error(err),
			logging.String("entryID", entryID.String()),
		)
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	// Prepare response
	resp := &response.MeaningListResponse{
		Meanings: make([]*response.MeaningResponse, len(entry.Meanings)),
		Total:    len(entry.Meanings),
	}

	// Map meanings to response
	for i, meaning := range entry.Meanings {
		resp.Meanings[i] = mapper.MeaningToResponse(&meaning)
	}

	return resp, nil
}

// AddMeaningComment implements EntryService.AddMeaningComment
func (s *entryService) AddMeaningComment(ctx context.Context, meaningID uuid.UUID, req *request.CreateCommentRequest) (*response.CommentResponse, error) {
	s.logger.Debug("adding comment to meaning",
		logging.String("meaningID", meaningID.String()),
		logging.String("userID", req.UserID.String()),
	)

	// Find the meaning by ID
	params := repository.ListParams{
		Limit: 100, // Reasonable limit to search through entries
	}
	entries, err := s.repo.ListEntries(ctx, params)
	if err != nil {
		s.logger.Error("failed to list entries to find meaning",
			logging.Error(err),
			logging.String("meaningID", meaningID.String()),
		)
		return nil, fmt.Errorf("failed to find meaning: %w", err)
	}

	// Find the meaning
	var foundMeaning *database.Meaning

	for i := range entries {
		entry := &entries[i]
		for j := range entry.Meanings {
			meaning := &entry.Meanings[j]
			if meaning.ID == meaningID {
				foundMeaning = meaning
				break
			}
		}
		if foundMeaning != nil {
			break
		}
	}

	if foundMeaning == nil {
		return nil, database.ErrEntryNotFound
	}

	// Create a new comment
	comment := model.Comment{
		ID:         uuid.New(),
		UserID:     req.UserID,
		TargetType: "meaning",
		TargetID:   meaningID,
		Content:    req.Content,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// In a real implementation, you would save this comment to a dedicated comments table
	// For now, we'll create a mock response

	// Create user for the comment (would come from a user repository in a real implementation)
	user := &model.User{
		ID:       req.UserID,
		Username: "user" + req.UserID.String()[0:8], // Mock username
	}

	// Create response
	resp := &response.CommentResponse{
		ID:      comment.ID,
		Content: comment.Content,
		User: response.UserSummary{
			ID:       user.ID,
			Username: user.Username,
		},
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}

	return resp, nil
}

// ToggleMeaningLike implements EntryService.ToggleMeaningLike
func (s *entryService) ToggleMeaningLike(ctx context.Context, meaningID uuid.UUID, userID uuid.UUID) error {
	s.logger.Debug("toggling like on meaning",
		logging.String("meaningID", meaningID.String()),
		logging.String("userID", userID.String()),
	)

	// Find the meaning by ID
	params := repository.ListParams{
		Limit: 100, // Reasonable limit to search through entries
	}
	entries, err := s.repo.ListEntries(ctx, params)
	if err != nil {
		s.logger.Error("failed to list entries to find meaning",
			logging.Error(err),
			logging.String("meaningID", meaningID.String()),
		)
		return fmt.Errorf("failed to find meaning: %w", err)
	}

	// Find the meaning
	var foundMeaning *database.Meaning

	for i := range entries {
		entry := &entries[i]
		for j := range entry.Meanings {
			meaning := &entry.Meanings[j]
			if meaning.ID == meaningID {
				foundMeaning = meaning
				break
			}
		}
		if foundMeaning != nil {
			break
		}
	}

	if foundMeaning == nil {
		return database.ErrEntryNotFound
	}

	// In a real implementation, you would:
	// 1. Check if the user has already liked this meaning
	// 2. If yes, remove the like
	// 3. If no, add a new like
	// 4. Update the likes count for the meaning

	// Create or toggle like (would be saved to a database in a real implementation)
	like := model.Like{
		ID:         uuid.New(),
		UserID:     userID,
		TargetType: "meaning",
		TargetID:   meaningID,
		CreatedAt:  time.Now().UTC(),
	}

	// Placeholder for logging - in a real implementation, this would be saved
	s.logger.Info("like processed",
		logging.String("likeID", like.ID.String()),
		logging.String("meaningID", meaningID.String()),
		logging.String("userID", userID.String()),
	)

	return nil
}
