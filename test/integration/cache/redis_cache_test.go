// test/integration/cache/redis_cache_test.go
package cache_test

import (
    "context"
    "os"
    "strconv"
    "strings"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/valpere/trytrago/domain/logging"
    "github.com/valpere/trytrago/infrastructure/cache"
)

// TestData is a simple struct to use for cache testing
type TestData struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Value int    `json:"value"`
}

// setupRedisCache prepares a real Redis connection for integration testing
func setupRedisCache(t *testing.T) cache.Cache {
    // Get Redis address from environment (with default for local testing)
    redisAddr := os.Getenv("TRYTRAGO_CACHE_ADDRESS")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }

    logger, _ := logging.NewLogger(logging.NewDefaultOptions())

    // Create Redis configuration
    redisCfg := cache.RedisConfig{
        Host: "localhost",
        Port: 6379,
    }

    // If the address is set in the environment variable, parse it to get host and port
    if redisAddr != "" {
        // Parse host and port from the address
        parts := strings.Split(redisAddr, ":")
        if len(parts) == 2 {
            redisCfg.Host = parts[0]
            if port, err := strconv.Atoi(parts[1]); err == nil {
                redisCfg.Port = port
            }
        }
    }

    // Create Redis cache instance
    redisCache, err := cache.NewRedisCache(redisCfg, logger)
    require.NoError(t, err, "Failed to create Redis cache")

    return redisCache
}

// TestRedisCache_Integration tests the Redis cache with a real Redis instance
func TestRedisCache_Integration(t *testing.T) {
    // Skip if environment indicates to skip integration tests
    if testing.Short() || os.Getenv("INTEGRATION_TEST") != "true" {
        t.Skip("Skipping integration tests. Set INTEGRATION_TEST=true to run")
    }

    // Set up Redis cache
    rCache := setupRedisCache(t)
    defer rCache.Close()

    // Test context
    ctx := context.Background()

    // Test data
    testData := &TestData{
        ID:    "test-123",
        Name:  "Integration Test",
        Value: 42,
    }

    // Test Set and Get
    t.Run("Set and Get", func(t *testing.T) {
        key := "test:integration:set-get"

        // First, ensure the key doesn't exist
        _ = rCache.Delete(ctx, key)

        // Set the value
        err := rCache.Set(ctx, key, testData, 1*time.Minute)
        require.NoError(t, err, "Set should succeed")

        // Get the value
        var result TestData
        err = rCache.Get(ctx, key, &result)
        require.NoError(t, err, "Get should succeed")

        // Verify the retrieved value
        assert.Equal(t, testData.ID, result.ID, "ID should match")
        assert.Equal(t, testData.Name, result.Name, "Name should match")
        assert.Equal(t, testData.Value, result.Value, "Value should match")
    })

    // Test Exists
    t.Run("Exists", func(t *testing.T) {
        keyExists := "test:integration:exists"
        keyNotExists := "test:integration:not-exists"

        // First, ensure the keys don't exist
        _ = rCache.Delete(ctx, keyExists)
        _ = rCache.Delete(ctx, keyNotExists)

        // Set a value for the first key
        err := rCache.Set(ctx, keyExists, "exists", 1*time.Minute)
        require.NoError(t, err, "Set should succeed")

        // Check if the keys exist
        exists, err := rCache.Exists(ctx, keyExists)
        require.NoError(t, err, "Exists should succeed")
        assert.True(t, exists, "Key should exist")

        exists, err = rCache.Exists(ctx, keyNotExists)
        require.NoError(t, err, "Exists should succeed")
        assert.False(t, exists, "Key should not exist")
    })

    // Test Delete
    t.Run("Delete", func(t *testing.T) {
        key := "test:integration:delete"

        // Set a value
        err := rCache.Set(ctx, key, "delete-me", 1*time.Minute)
        require.NoError(t, err, "Set should succeed")

        // Verify it exists
        exists, err := rCache.Exists(ctx, key)
        require.NoError(t, err, "Exists should succeed")
        assert.True(t, exists, "Key should exist")

        // Delete it
        err = rCache.Delete(ctx, key)
        require.NoError(t, err, "Delete should succeed")

        // Verify it no longer exists
        exists, err = rCache.Exists(ctx, key)
        require.NoError(t, err, "Exists should succeed")
        assert.False(t, exists, "Key should no longer exist")
    })

    // Test Invalidate
    t.Run("Invalidate", func(t *testing.T) {
        // Create multiple keys with a common pattern
        pattern := "test:integration:invalidate:"
        keys := []string{
            pattern + "key1",
            pattern + "key2",
            pattern + "key3",
        }

        // Set values for all keys
        for i, key := range keys {
            err := rCache.Set(ctx, key, i, 1*time.Minute)
            require.NoError(t, err, "Set should succeed")
        }

        // Verify all keys exist
        for _, key := range keys {
            exists, err := rCache.Exists(ctx, key)
            require.NoError(t, err, "Exists should succeed")
            assert.True(t, exists, "Key should exist")
        }

        // Invalidate keys matching the pattern
        err := rCache.Invalidate(ctx, pattern+"*")
        require.NoError(t, err, "Invalidate should succeed")

        // Verify all keys no longer exist
        for _, key := range keys {
            exists, err := rCache.Exists(ctx, key)
            require.NoError(t, err, "Exists should succeed")
            assert.False(t, exists, "Key should no longer exist after invalidation")
        }
    })

    // Test Expiration
    t.Run("Expiration", func(t *testing.T) {
        key := "test:integration:expiration"

        // Set a value with a very short TTL (100ms)
        err := rCache.Set(ctx, key, "expiring", 100*time.Millisecond)
        require.NoError(t, err, "Set should succeed")

        // Verify it exists immediately
        exists, err := rCache.Exists(ctx, key)
        require.NoError(t, err, "Exists should succeed")
        assert.True(t, exists, "Key should exist before expiration")

        // Wait for the key to expire
        time.Sleep(200 * time.Millisecond)

        // Verify it no longer exists
        exists, err = rCache.Exists(ctx, key)
        require.NoError(t, err, "Exists should succeed")
        assert.False(t, exists, "Key should not exist after expiration")
    })

    // Test JSON handling
    t.Run("Complex JSON", func(t *testing.T) {
        key := "test:integration:complex-json"

        // Complex nested structure
        complexData := map[string]interface{}{
            "string": "value",
            "number": 42,
            "bool":   true,
            "nested": map[string]interface{}{
                "array": []string{"one", "two", "three"},
                "object": map[string]interface{}{
                    "key": "value",
                },
            },
            "nullValue": nil,
        }

        // Set the complex data
        err := rCache.Set(ctx, key, complexData, 1*time.Minute)
        require.NoError(t, err, "Set should succeed with complex data")

        // Get the data back
        var result map[string]interface{}
        err = rCache.Get(ctx, key, &result)
        require.NoError(t, err, "Get should succeed")

        // Verify structure (partial verification)
        assert.Equal(t, "value", result["string"], "String value should match")
        assert.Equal(t, float64(42), result["number"], "Number should match (as float64)")
        assert.Equal(t, true, result["bool"], "Boolean should match")

        // Check nested structures exist
        assert.Contains(t, result, "nested", "Nested field should exist")
        nested, ok := result["nested"].(map[string]interface{})
        require.True(t, ok, "Nested should be a map")

        assert.Contains(t, nested, "array", "Array field should exist")
        assert.Contains(t, nested, "object", "Object field should exist")
    })
}
