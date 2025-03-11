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

// AddMeaning, UpdateMeaning, DeleteMeaning, ListMeanings method implementations...

// Social operations for meanings implementations...
func (s *entryService) AddMeaning(ctx context.Context, entryID uuid.UUID, req *request.CreateMeaningRequest) (*response.MeaningResponse, error) {
	// Implementation to be added
	return nil, fmt.Errorf("not implemented")
}

func (s *entryService) UpdateMeaning(ctx context.Context, id uuid.UUID, req *request.UpdateMeaningRequest) (*response.MeaningResponse, error) {
	// Implementation to be added
	return nil, fmt.Errorf("not implemented")
}

func (s *entryService) DeleteMeaning(ctx context.Context, id uuid.UUID) error {
	// Implementation to be added
	return fmt.Errorf("not implemented")
}

func (s *entryService) ListMeanings(ctx context.Context, entryID uuid.UUID) (*response.MeaningListResponse, error) {
	// Implementation to be added
	return nil, fmt.Errorf("not implemented")
}

func (s *entryService) AddMeaningComment(ctx context.Context, meaningID uuid.UUID, req *request.CreateCommentRequest) (*response.CommentResponse, error) {
	// Implementation to be added
	return nil, fmt.Errorf("not implemented")
}

func (s *entryService) ToggleMeaningLike(ctx context.Context, meaningID uuid.UUID, userID uuid.UUID) error {
	// Implementation to be added
	return fmt.Errorf("not implemented")
}
