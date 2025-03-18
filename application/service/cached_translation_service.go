// application/service/cached_translation_service.go
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/domain/cache"
	"github.com/valpere/trytrago/domain/logging"
)

// cachedTranslationService implements the TranslationService interface with Redis caching
type cachedTranslationService struct {
	baseService TranslationService
	cache       cache.CacheService
	logger      logging.Logger
}

// NewCachedTranslationService creates a new cached translation service
func NewCachedTranslationService(
	baseService TranslationService,
	cacheService cache.CacheService,
	logger logging.Logger,
) TranslationService {
	return &cachedTranslationService{
		baseService: baseService,
		cache:       cacheService,
		logger:      logger.With(logging.String("service", "cached_translation_service")),
	}
}

// CreateTranslation implements TranslationService.CreateTranslation with cache invalidation
func (s *cachedTranslationService) CreateTranslation(
	ctx context.Context,
	meaningID uuid.UUID,
	req *request.CreateTranslationRequest,
) (*response.TranslationResponse, error) {
	// Call base service to create the translation
	resp, err := s.baseService.CreateTranslation(ctx, meaningID, req)
	if err != nil {
		return nil, err
	}

	// Invalidate translations list cache
	listCacheKey := s.cache.GenerateKey("meanings", meaningID.String(), "translations", "list")
	if err := s.cache.Delete(ctx, listCacheKey); err != nil {
		s.logger.Warn("failed to invalidate translations list cache after create",
			logging.String("meaningId", meaningID.String()),
			logging.Error(err),
		)
	}

	// Invalidate language-specific translations cache
	langCacheKey := s.cache.GenerateKey(
		"meanings",
		meaningID.String(),
		"translations",
		"language",
		req.LanguageID,
	)
	if err := s.cache.Delete(ctx, langCacheKey); err != nil {
		s.logger.Warn("failed to invalidate language translations cache after create",
			logging.String("meaningId", meaningID.String()),
			logging.String("languageId", req.LanguageID),
			logging.Error(err),
		)
	}

	// Invalidate the meaning cache as it may contain translations
	meaningCacheKey := s.cache.GenerateKey("meanings", "id", meaningID.String())
	if err := s.cache.Delete(ctx, meaningCacheKey); err != nil {
		s.logger.Warn("failed to invalidate meaning cache after translation create",
			logging.String("meaningId", meaningID.String()),
			logging.Error(err),
		)
	}

	return resp, nil
}

// UpdateTranslation implements TranslationService.UpdateTranslation with cache invalidation
func (s *cachedTranslationService) UpdateTranslation(
	ctx context.Context,
	id uuid.UUID,
	req *request.UpdateTranslationRequest,
) (*response.TranslationResponse, error) {
	// Call base service to update the translation
	resp, err := s.baseService.UpdateTranslation(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Invalidate translation cache
	cacheKey := s.cache.GenerateKey("translations", "id", id.String())
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("failed to invalidate translation cache after update",
			logging.String("translationId", id.String()),
			logging.Error(err),
		)
	}

	// Invalidate meaning cache since it may contain translations
	if resp != nil {
		meaningCacheKey := s.cache.GenerateKey("meanings", "id", resp.MeaningID.String())
		if err := s.cache.Delete(ctx, meaningCacheKey); err != nil {
			s.logger.Warn("failed to invalidate meaning cache after translation update",
				logging.String("meaningId", resp.MeaningID.String()),
				logging.Error(err),
			)
		}

		// Invalidate translations list cache
		listCacheKey := s.cache.GenerateKey("meanings", resp.MeaningID.String(), "translations", "list")
		if err := s.cache.Delete(ctx, listCacheKey); err != nil {
			s.logger.Warn("failed to invalidate translations list cache after update",
				logging.String("meaningId", resp.MeaningID.String()),
				logging.Error(err),
			)
		}

		// Invalidate language-specific translations cache
		langCacheKey := s.cache.GenerateKey(
			"meanings",
			resp.MeaningID.String(),
			"translations",
			"language",
			resp.LanguageID,
		)
		if err := s.cache.Delete(ctx, langCacheKey); err != nil {
			s.logger.Warn("failed to invalidate language translations cache after update",
				logging.String("meaningId", resp.MeaningID.String()),
				logging.String("languageId", resp.LanguageID),
				logging.Error(err),
			)
		}
	}

	return resp, nil
}

// DeleteTranslation implements TranslationService.DeleteTranslation with cache invalidation
func (s *cachedTranslationService) DeleteTranslation(ctx context.Context, id uuid.UUID) error {
	// First, get the translation to retrieve its meaning ID and language ID for cache invalidation
	// In a real implementation, you might want to include this info in the request context
	// Here we'll assume we don't have that information and proceed with broader invalidation

	// Call base service to delete the translation
	if err := s.baseService.DeleteTranslation(ctx, id); err != nil {
		return err
	}

	// Invalidate translation cache
	cacheKey := s.cache.GenerateKey("translations", "id", id.String())
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("failed to invalidate translation cache after delete",
			logging.String("translationId", id.String()),
			logging.Error(err),
		)
	}

	// Invalidate all meaning caches that might contain this translation
	// In a real implementation, this would be more targeted
	if err := s.cache.Invalidate(ctx, "meanings:*"); err != nil {
		s.logger.Warn("failed to invalidate meaning caches after translation delete",
			logging.Error(err),
		)
	}

	// Invalidate all translations list caches
	if err := s.cache.Invalidate(ctx, "meanings:*:translations:*"); err != nil {
		s.logger.Warn("failed to invalidate translations list caches after delete",
			logging.Error(err),
		)
	}

	return nil
}

// ListTranslations implements TranslationService.ListTranslations with caching
func (s *cachedTranslationService) ListTranslations(
	ctx context.Context,
	meaningID uuid.UUID,
	langID string,
) (*response.TranslationListResponse, error) {
	var cacheKey string

	// Generate appropriate cache key based on whether language filter is provided
	if langID != "" {
		cacheKey = s.cache.GenerateKey(
			"meanings",
			meaningID.String(),
			"translations",
			"language",
			langID,
		)
	} else {
		cacheKey = s.cache.GenerateKey(
			"meanings",
			meaningID.String(),
			"translations",
			"list",
		)
	}

	// Try to get from cache first
	var cachedList response.TranslationListResponse
	err := s.cache.Get(ctx, cacheKey, &cachedList)
	if err == nil {
		s.logger.Debug("cache hit for translations list",
			logging.String("meaningId", meaningID.String()),
			logging.String("languageId", langID),
		)
		return &cachedList, nil
	}

	// On cache miss, get from base service
	list, err := s.baseService.ListTranslations(ctx, meaningID, langID)
	if err != nil {
		return nil, err
	}

	// Store in cache for future requests
	if err := s.cache.Set(ctx, cacheKey, list, defaultListTTL); err != nil {
		s.logger.Warn("failed to cache translations list",
			logging.String("meaningId", meaningID.String()),
			logging.String("languageId", langID),
			logging.Error(err),
		)
	}

	return list, nil
}

// AddTranslationComment implements TranslationService.AddTranslationComment with cache invalidation
func (s *cachedTranslationService) AddTranslationComment(
	ctx context.Context,
	translationID uuid.UUID,
	req *request.CreateCommentRequest,
) (*response.CommentResponse, error) {
	// Call base service to add the comment
	resp, err := s.baseService.AddTranslationComment(ctx, translationID, req)
	if err != nil {
		return nil, err
	}

	// Invalidate translation cache
	translationCacheKey := s.cache.GenerateKey("translations", "id", translationID.String())
	if err := s.cache.Delete(ctx, translationCacheKey); err != nil {
		s.logger.Warn("failed to invalidate translation cache after adding comment",
			logging.String("translationId", translationID.String()),
			logging.Error(err),
		)
	}

	// Invalidate comments list cache
	commentsCacheKey := s.cache.GenerateKey("translations", translationID.String(), "comments")
	if err := s.cache.Delete(ctx, commentsCacheKey); err != nil {
		s.logger.Warn("failed to invalidate comments cache",
			logging.String("translationId", translationID.String()),
			logging.Error(err),
		)
	}

	return resp, nil
}

// ToggleTranslationLike implements TranslationService.ToggleTranslationLike with cache invalidation
func (s *cachedTranslationService) ToggleTranslationLike(
	ctx context.Context,
	translationID uuid.UUID,
	userID uuid.UUID,
) error {
	// Call base service to toggle the like
	if err := s.baseService.ToggleTranslationLike(ctx, translationID, userID); err != nil {
		return err
	}

	// Invalidate translation cache
	translationCacheKey := s.cache.GenerateKey("translations", "id", translationID.String())
	if err := s.cache.Delete(ctx, translationCacheKey); err != nil {
		s.logger.Warn("failed to invalidate translation cache after toggling like",
			logging.String("translationId", translationID.String()),
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
