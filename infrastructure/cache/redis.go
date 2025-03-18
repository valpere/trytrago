package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/valpere/trytrago/domain/logging"
)

// Cache defines the interface for caching operations
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Invalidate(ctx context.Context, pattern string) error
	Close() error
}

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	logger logging.Logger
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(cfg RedisConfig, logger logging.Logger) (Cache, error) {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		logger: logger.With(logging.String("component", "redis_cache")),
	}, nil
}

// Get retrieves a value from the cache
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			c.logger.Debug("cache miss", logging.String("key", key))
			return fmt.Errorf("key not found")
		}
		c.logger.Error("error getting from cache",
			logging.String("key", key),
			logging.Error(err),
		)
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	// Unmarshal the JSON value
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		c.logger.Error("error unmarshaling cached value",
			logging.String("key", key),
			logging.Error(err),
		)
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	c.logger.Debug("cache hit", logging.String("key", key))
	return nil
}

// Set stores a value in the cache
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Marshal the value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("error marshaling value for cache",
			logging.String("key", key),
			logging.Error(err),
		)
		return fmt.Errorf("failed to marshal value for cache: %w", err)
	}

	// Store in Redis
	if err := c.client.Set(ctx, key, data, expiration).Err(); err != nil {
		c.logger.Error("error setting cache",
			logging.String("key", key),
			logging.Error(err),
		)
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.logger.Debug("cache set",
		logging.String("key", key),
		logging.Duration("expiration", expiration),
	)
	return nil
}

// Delete removes a value from the cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		c.logger.Error("error deleting from cache",
			logging.String("key", key),
			logging.Error(err),
		)
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	c.logger.Debug("cache delete", logging.String("key", key))
	return nil
}

// Exists checks if a key exists in the cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		c.logger.Error("error checking cache existence",
			logging.String("key", key),
			logging.Error(err),
		)
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}

	return exists == 1, nil
}

// Invalidate removes all keys matching a pattern
func (c *RedisCache) Invalidate(ctx context.Context, pattern string) error {
	// Find all keys matching the pattern
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		c.logger.Error("error finding keys for invalidation",
			logging.String("pattern", pattern),
			logging.Error(err),
		)
		return fmt.Errorf("failed to find keys for invalidation: %w", err)
	}

	// If there are no keys, return
	if len(keys) == 0 {
		return nil
	}

	// Delete all matching keys
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		c.logger.Error("error invalidating cache",
			logging.String("pattern", pattern),
			logging.Error(err),
		)
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	c.logger.Debug("cache invalidated",
		logging.String("pattern", pattern),
		logging.Int("keys_removed", len(keys)),
	)
	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}
