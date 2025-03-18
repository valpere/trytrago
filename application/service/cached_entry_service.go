// application/service/cached_entry_service.go
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/domain/cache"
	"github.com/valpere/trytrago/domain/logging"
)

// Default cache TTLs
const (
	defaultEntryTTL  = 15 * time.Minute
	defaultListTTL   = 5 * time.Minute
	defaultSocialTTL = 2 * time.Minute
)

// cachedEntryService implements the EntryService interface with Redis caching
type cachedEntryService struct {
	baseService EntryService
	cache       cache.CacheService
	logger      logging.Logger
}

// NewCachedEntryService creates a new cached entry service
func NewCachedEntryService(baseService EntryService, cacheService cache.CacheService, logger logging.Logger) EntryService {
	return &cachedEntryService{
		baseService: baseService,
		cache:       cacheService,
		logger:      logger.With(logging.String("service", "cached_entry_service")),
	}
}

// CreateEntry implements EntryService.CreateEntry with cache invalidation
func (s *cachedEntryService) CreateEntry(ctx context.Context, req *request.CreateEntryRequest) (*response.EntryResponse, error) {
	// Call base service to create the entry
	resp, err := s.baseService.CreateEntry(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate any list caches that might contain the new entry
	if err := s.cache.Invalidate(ctx, "entries:list:*"); err != nil {
		s.logger.Warn("failed to invalidate entry list cache after create",
			logging.Error(err),
		)
	}

	return resp, nil
}

// GetEntryByID implements EntryService.GetEntryByID with caching
func (s *cachedEntryService) GetEntryByID(ctx context.Context, id uuid.UUID) (*response.EntryResponse, error) {
	cacheKey := s.cache.GenerateKey("entries", "id", id.String())

	// Try to get from cache first
	var cachedEntry response.EntryResponse
	err := s.cache.Get(ctx, cacheKey, &cachedEntry)
	if err == nil {
		s.logger.Debug("cache hit for entry",
			logging.String("id", id.String()),
		)
		return &cachedEntry, nil
	}

	// On cache miss, get from base service
	entry, err := s.baseService.GetEntryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache for future requests
	if err := s.cache.Set(ctx, cacheKey, entry, defaultEntryTTL); err != nil {
		s.logger.Warn("failed to cache entry",
			logging.String("id", id.String()),
			logging.Error(err),
		)
	}

	return entry, nil
}

// UpdateEntry implements EntryService.UpdateEntry with cache invalidation
func (s *cachedEntryService) UpdateEntry(ctx context.Context, id uuid.UUID, req *request.UpdateEntryRequest) (*response.EntryResponse, error) {
	// Call base service to update the entry
	resp, err := s.baseService.UpdateEntry(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Invalidate specific entry cache
	cacheKey := s.cache.GenerateKey("entries", "id", id.String())
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("failed to invalidate entry cache after update",
			logging.String("id", id.String()),
			logging.Error(err),
		)
	}

	// Invalidate any list caches
	if err := s.cache.Invalidate(ctx, "entries:list:*"); err != nil {
		s.logger.Warn("failed to invalidate entry list cache after update",
			logging.Error(err),
		)
	}

	return resp, nil
}

// DeleteEntry implements EntryService.DeleteEntry with cache invalidation
func (s *cachedEntryService) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	// Call base service to delete the entry
	if err := s.baseService.DeleteEntry(ctx, id); err != nil {
		return err
	}

	// Invalidate specific entry cache
	cacheKey := s.cache.GenerateKey("entries", "id", id.String())
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("failed to invalidate entry cache after delete",
			logging.String("id", id.String()),
			logging.Error(err),
		)
	}

	// Invalidate any list caches
	if err := s.cache.Invalidate(ctx, "entries:list:*"); err != nil {
		s.logger.Warn("failed to invalidate entry list cache after delete",
			logging.Error(err),
		)
	}

	// Invalidate any meaning caches associated with this entry
	if err := s.cache.Invalidate(ctx, fmt.Sprintf("entries:%s:meanings:*", id.String())); err != nil {
		s.logger.Warn("failed to invalidate meaning caches after entry delete",
			logging.String("entryId", id.String()),
			logging.Error(err),
		)
	}

	return nil
}

// ListEntries implements EntryService.ListEntries with caching
func (s *cachedEntryService) ListEntries(ctx context.Context, req *request.ListEntriesRequest) (*response.EntryListResponse, error) {
	// Generate a cache key based on the request parameters
	cacheKey := s.generateListCacheKey(req)

	// Try to get from cache first
	var cachedList response.EntryListResponse
	err := s.cache.Get(ctx, cacheKey, &cachedList)
	if err == nil {
		s.logger.Debug("cache hit for entry list",
			logging.String("cacheKey", cacheKey),
		)
		return &cachedList, nil
	}

	// On cache miss, get from base service
	list, err := s.baseService.ListEntries(ctx, req)
	if err != nil {
		return nil, err
	}

	// Store in cache for future requests
	if err := s.cache.Set(ctx, cacheKey, list, defaultListTTL); err != nil {
		s.logger.Warn("failed to cache entry list",
			logging.String("cacheKey", cacheKey),
			logging.Error(err),
		)
	}

	return list, nil
}

// generateListCacheKey creates a cache key for list requests based on parameters
func (s *cachedEntryService) generateListCacheKey(req *request.ListEntriesRequest) string {
	key := s.cache.GenerateKey("entries", "list",
		fmt.Sprintf("limit:%d", req.Limit),
		fmt.Sprintf("offset:%d", req.Offset),
	)

	if req.SortBy != "" {
		key = s.cache.GenerateKey(key, fmt.Sprintf("sort:%s", req.SortBy))
	}

	if req.SortDesc {
		key = s.cache.GenerateKey(key, "desc")
	}

	if req.WordFilter != "" {
		key = s.cache.GenerateKey(key, fmt.Sprintf("filter:%s", req.WordFilter))
	}

	if req.Type != "" {
		key = s.cache.GenerateKey(key, fmt.Sprintf("type:%s", req.Type))
	}

	return key
}

// AddMeaning implements EntryService.AddMeaning with cache invalidation
func (s *cachedEntryService) AddMeaning(ctx context.Context, entryID uuid.UUID, req *request.CreateMeaningRequest) (*response.MeaningResponse, error) {
	// Call base service to add the meaning
	resp, err := s.baseService.AddMeaning(ctx, entryID, req)
	if err != nil {
		return nil, err
	}

	// Invalidate entry cache since it includes meanings
	entryCacheKey := s.cache.GenerateKey("entries", "id", entryID.String())
	if err := s.cache.Delete(ctx, entryCacheKey); err != nil {
		s.logger.Warn("failed to invalidate entry cache after adding meaning",
			logging.String("entryId", entryID.String()),
			logging.Error(err),
		)
	}

	// Invalidate meanings list cache
	meaningListCacheKey := s.cache.GenerateKey("entries", entryID.String(), "meanings", "list")
	if err := s.cache.Delete(ctx, meaningListCacheKey); err != nil {
		s.logger.Warn("failed to invalidate meanings list cache after adding meaning",
			logging.String("entryId", entryID.String()),
			logging.Error(err),
		)
	}

	return resp, nil
}

// UpdateMeaning implements EntryService.UpdateMeaning with cache invalidation
func (s *cachedEntryService) UpdateMeaning(ctx context.Context, id uuid.UUID, req *request.UpdateMeaningRequest) (*response.MeaningResponse, error) {
	// Call base service to update the meaning
	resp, err := s.baseService.UpdateMeaning(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Invalidate meaning cache
	meaningCacheKey := s.cache.GenerateKey("meanings", "id", id.String())
	if err := s.cache.Delete(ctx, meaningCacheKey); err != nil {
		s.logger.Warn("failed to invalidate meaning cache after update",
			logging.String("meaningId", id.String()),
			logging.Error(err),
		)
	}

	// Invalidate entry cache that contains this meaning
	if resp != nil {
		entryCacheKey := s.cache.GenerateKey("entries", "id", resp.EntryID.String())
		if err := s.cache.Delete(ctx, entryCacheKey); err != nil {
			s.logger.Warn("failed to invalidate entry cache after meaning update",
				logging.String("entryId", resp.EntryID.String()),
				logging.Error(err),
			)
		}
	}

	return resp, nil
}

// DeleteMeaning implements EntryService.DeleteMeaning with cache invalidation
func (s *cachedEntryService) DeleteMeaning(ctx context.Context, id uuid.UUID) error {
	// We need to find the entry ID before deleting the meaning for cache invalidation
	// This requires an additional database query
	// In a real implementation, you might want to get this information from the request context

	// Call base service to delete the meaning
	if err := s.baseService.DeleteMeaning(ctx, id); err != nil {
		return err
	}

	// Invalidate meaning cache
	meaningCacheKey := s.cache.GenerateKey("meanings", "id", id.String())
	if err := s.cache.Delete(ctx, meaningCacheKey); err != nil {
		s.logger.Warn("failed to invalidate meaning cache after delete",
			logging.String("meaningId", id.String()),
			logging.Error(err),
		)
	}

	// Invalidate all potential entry caches
	if err := s.cache.Invalidate(ctx, "entries:id:*"); err != nil {
		s.logger.Warn("failed to invalidate entry caches after meaning delete",
			logging.Error(err),
		)
	}

	return nil
}

// ListMeanings implements EntryService.ListMeanings with caching
func (s *cachedEntryService) ListMeanings(ctx context.Context, entryID uuid.UUID) (*response.MeaningListResponse, error) {
	cacheKey := s.cache.GenerateKey("entries", entryID.String(), "meanings", "list")

	// Try to get from cache first
	var cachedList response.MeaningListResponse
	err := s.cache.Get(ctx, cacheKey, &cachedList)
	if err == nil {
		s.logger.Debug("cache hit for meaning list",
			logging.String("entryId", entryID.String()),
		)
		return &cachedList, nil
	}

	// On cache miss, get from base service
	list, err := s.baseService.ListMeanings(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Store in cache for future requests
	if err := s.cache.Set(ctx, cacheKey, list, defaultListTTL); err != nil {
		s.logger.Warn("failed to cache meaning list",
			logging.String("entryId", entryID.String()),
			logging.Error(err),
		)
	}

	return list, nil
}

// AddMeaningComment implements EntryService.AddMeaningComment with cache invalidation
func (s *cachedEntryService) AddMeaningComment(ctx context.Context, meaningID uuid.UUID, req *request.CreateCommentRequest) (*response.CommentResponse, error) {
	// Call base service to add the comment
	resp, err := s.baseService.AddMeaningComment(ctx, meaningID, req)
	if err != nil {
		return nil, err
	}

	// Invalidate meaning cache
	meaningCacheKey := s.cache.GenerateKey("meanings", "id", meaningID.String())
	if err := s.cache.Delete(ctx, meaningCacheKey); err != nil {
		s.logger.Warn("failed to invalidate meaning cache after adding comment",
			logging.String("meaningId", meaningID.String()),
			logging.Error(err),
		)
	}

	// Invalidate comments list cache
	commentsCacheKey := s.cache.GenerateKey("meanings", meaningID.String(), "comments")
	if err := s.cache.Delete(ctx, commentsCacheKey); err != nil {
		s.logger.Warn("failed to invalidate comments cache",
			logging.String("meaningId", meaningID.String()),
			logging.Error(err),
		)
	}

	return resp, nil
}

// ToggleMeaningLike implements EntryService.ToggleMeaningLike with cache invalidation
func (s *cachedEntryService) ToggleMeaningLike(ctx context.Context, meaningID uuid.UUID, userID uuid.UUID) error {
	// Call base service to toggle the like
	if err := s.baseService.ToggleMeaningLike(ctx, meaningID, userID); err != nil {
		return err
	}

	// Invalidate meaning cache
	meaningCacheKey := s.cache.GenerateKey("meanings", "id", meaningID.String())
	if err := s.cache.Delete(ctx, meaningCacheKey); err != nil {
		s.logger.Warn("failed to invalidate meaning cache after toggling like",
			logging.String("meaningId", meaningID.String()),
			logging.Error(err),
		)
	}

	// Invalidate user's likes cache
	userLikesCacheKey := s.cache.GenerateKey("users", userID.String(), "likes")
	if err := s.cache.Delete(ctx, userLikesCacheKey); err != nil {
		s.logger.Warn("failed to invalidate user likes cache",
			logging.String("userId", userID.String()),
			logging.Error(err),
		)
	}

	return nil
}
