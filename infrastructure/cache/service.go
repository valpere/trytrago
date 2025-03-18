package cache

import (
	"context"
	"strings"
	"time"

	"github.com/valpere/trytrago/domain/cache"
	"github.com/valpere/trytrago/domain/logging"
)

// redisCacheService implements the CacheService interface using Redis
type redisCacheService struct {
	redisCache Cache
	logger     logging.Logger
	prefix     string
}

// NewRedisCacheService creates a new Redis cache service
func NewRedisCacheService(redisCache Cache, logger logging.Logger, prefix string) cache.CacheService {
	return &redisCacheService{
		redisCache: redisCache,
		logger:     logger.With(logging.String("component", "redis_cache_service")),
		prefix:     prefix,
	}
}

// Get retrieves a value from the cache
func (s *redisCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := s.prefixKey(key)
	err := s.redisCache.Get(ctx, fullKey, dest)
	if err != nil {
		s.logger.Debug("cache miss",
			logging.String("key", key),
			logging.Error(err),
		)
		return err
	}

	s.logger.Debug("cache hit", logging.String("key", key))
	return nil
}

// Set stores a value in the cache
func (s *redisCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := s.prefixKey(key)
	if err := s.redisCache.Set(ctx, fullKey, value, expiration); err != nil {
		s.logger.Error("failed to set cache key",
			logging.String("key", key),
			logging.Error(err),
		)
		return err
	}

	s.logger.Debug("cache set successful",
		logging.String("key", key),
		logging.Duration("expiration", expiration),
	)
	return nil
}

// Delete removes a value from the cache
func (s *redisCacheService) Delete(ctx context.Context, key string) error {
	fullKey := s.prefixKey(key)
	if err := s.redisCache.Delete(ctx, fullKey); err != nil {
		s.logger.Error("failed to delete cache key",
			logging.String("key", key),
			logging.Error(err),
		)
		return err
	}

	s.logger.Debug("cache delete successful", logging.String("key", key))
	return nil
}

// Exists checks if a key exists in the cache
func (s *redisCacheService) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := s.prefixKey(key)
	exists, err := s.redisCache.Exists(ctx, fullKey)
	if err != nil {
		s.logger.Error("failed to check cache key existence",
			logging.String("key", key),
			logging.Error(err),
		)
		return false, err
	}

	return exists, nil
}

// Invalidate removes all keys matching a pattern
func (s *redisCacheService) Invalidate(ctx context.Context, pattern string) error {
	fullPattern := s.prefixKey(pattern)
	if err := s.redisCache.Invalidate(ctx, fullPattern); err != nil {
		s.logger.Error("failed to invalidate cache keys",
			logging.String("pattern", pattern),
			logging.Error(err),
		)
		return err
	}

	s.logger.Debug("cache invalidation successful", logging.String("pattern", pattern))
	return nil
}

// GenerateKey creates a standardized cache key
func (s *redisCacheService) GenerateKey(keyParts ...string) string {
	return strings.Join(keyParts, ":")
}

// prefixKey adds the application prefix to a cache key
func (s *redisCacheService) prefixKey(key string) string {
	if s.prefix == "" {
		return key
	}
	return s.prefix + ":" + key
}
