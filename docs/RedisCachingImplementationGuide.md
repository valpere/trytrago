# Redis Caching Implementation Guide

## Overview

TryTraGo uses Redis for caching to improve performance, especially for frequently accessed dictionary entries and translations. This document describes the Redis caching implementation, configuration, and best practices.

## Configuration

Redis caching is configured in the `config.yaml` file:

```yaml
# Cache configuration
cache:
  # Enable caching
  enabled: true
  # Redis host
  host: localhost
  # Redis port
  port: 6379
  # Redis password (optional)
  password: ""
  # Redis database number
  db: 0
  # Default TTL for cached items
  ttl: 10m
  # Cache TTLs for specific data types
  entry_ttl: 15m
  list_ttl: 5m
  social_ttl: 2m
  translation_ttl: 15m
  # Key prefix (optional, defaults to "trytrago:<environment>")
  key_prefix: ""
```

## Architecture

The caching implementation follows these patterns:

1. **Interface-based design**: Both the cache client and service are interface-based, allowing for easy mocking in tests.
2. **Decorator pattern**: Services are wrapped with caching capabilities using the decorator pattern.
3. **Error resilience**: Cache errors never propagate to the client; the system degrades gracefully to direct database access.

## Key Components

### 1. Cache Interface

The `Cache` interface in `infrastructure/cache/redis.go`:

```go
type Cache interface {
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    Invalidate(ctx context.Context, pattern string) error
    Close() error
}
```

### 2. Redis Implementation

The `RedisCache` struct implements the `Cache` interface with Redis-specific functionality.

### 3. Cache Service

The `CacheService` interface in `domain/cache/cache.go`:

```go
type CacheService interface {
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    Invalidate(ctx context.Context, pattern string) error
    GenerateKey(keyParts ...string) string
}
```

### 4. Cached Services

Cached versions of application services:
- `CachedEntryService` - Caches dictionary entries
- `CachedTranslationService` - Caches translations

## TTL (Time-To-Live) Values

Default TTL values are defined in `application/service/cached_entry_service.go`:

```go
const (
    defaultEntryTTL  = 15 * time.Minute
    defaultListTTL   = 5 * time.Minute
    defaultSocialTTL = 2 * time.Minute
)
```

## Caching Strategies

### 1. Entry Caching

- Individual entries are cached with their ID as the key
- Lists of entries are cached with parameters included in the key
- Cache invalidation occurs on create/update/delete operations

### 2. Translation Caching

- Translations are cached by meaning ID and language
- Cache is invalidated when translations are modified
- Translation comments and likes invalidate related caches

## Error Handling

The caching layer never propagates errors to clients. If a cache operation fails:
1. The error is logged
2. The system falls back to the underlying service

## Cache Key Structure

Keys are constructed using the `GenerateKey` method following this pattern:
- Prefix: `trytrago:<environment>`
- Entity type: `entries`, `meanings`, `translations`
- ID: The entity's unique identifier
- Action/subtype: `list`, `comments`, `likes`
- Parameters: Any query parameters included in the key

Example: `trytrago:development:entries:list:limit:20:offset:0:type:WORD`

## Integration Testing

Tests in the `test/integration/cache` directory verify Redis functionality against a real Redis instance. To run these tests:

```bash
INTEGRATION_TEST=true go test -v ./test/integration/cache/...
```

You can specify a custom Redis address using the `TRYTRAGO_CACHE_ADDRESS` environment variable:

```bash
TRYTRAGO_CACHE_ADDRESS=localhost:6380 INTEGRATION_TEST=true go test -v ./test/integration/cache/...
```

## Best Practices

1. **Always check cache first**: Cache operations should occur before database queries.
2. **Handle cache invalidation**: When data changes, invalidate related caches.
3. **Use appropriate TTLs**: Set shorter TTLs for frequently changing data.
4. **Log cache operations**: Log cache hits, misses, and errors for monitoring.
5. **Graceful degradation**: The system should continue to function if Redis is unavailable.

## Server Integration

The server initializes Redis in the `initializeCache` method of `AppServer`:
1. Checks if caching is enabled and Redis host is configured
2. Creates a Redis connection with the provided configuration
3. Creates a cache service with the appropriate prefix
4. Wraps application services with cached versions

When the server shuts down, it properly closes the Redis connection to prevent leaks.
