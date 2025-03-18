package cache

import (
	"context"
	"time"
)

// CacheService defines the interface for caching operations across the application
type CacheService interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string, dest interface{}) error

	// Set stores a value in the cache with expiration time
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Invalidate removes all keys matching a pattern
	Invalidate(ctx context.Context, pattern string) error

	// GenerateKey creates a standardized cache key
	GenerateKey(keyParts ...string) string
}
