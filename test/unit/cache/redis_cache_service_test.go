// test/unit/cache/redis_cache_service_test.go
package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/infrastructure/cache"
	"github.com/valpere/trytrago/test/mocks"
)

// TestData is a simple struct to use for cache testing
type TestData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// setupMocks prepares mock objects for testing
func setupMocks() (*mocks.MockRedisCache, logging.Logger) {
	mockCache := &mocks.MockRedisCache{}
	logger, _ := logging.NewLogger(logging.NewDefaultOptions())
	return mockCache, logger
}

// TestNewRedisCacheService tests the creation of the Redis cache service
func TestNewRedisCacheService(t *testing.T) {
	mockCache, logger := setupMocks()
	
	service := cache.NewRedisCacheService(mockCache, logger, "test")
	assert.NotNil(t, service, "Cache service should not be nil")
}

// TestCacheService_Get tests the Get method
func TestCacheService_Get(t *testing.T) {
	mockCache, logger := setupMocks()
	service := cache.NewRedisCacheService(mockCache, logger, "test")
	
	ctx := context.Background()
	testKey := "test:key"
	expectedData := TestData{ID: "123", Name: "Test", Value: 42}
	
	// Test successful retrieval
	mockCache.On("Get", ctx, "test:"+testKey, mock.AnythingOfType("*cache_test.TestData")).
		Run(func(args mock.Arguments) {
			// Simulate successful retrieval by setting the destination
			dest := args.Get(2).(*TestData)
			*dest = expectedData
		}).
		Return(nil)
	
	var result TestData
	err := service.Get(ctx, testKey, &result)
	
	assert.NoError(t, err, "Get should succeed")
	assert.Equal(t, expectedData, result, "Data should match expected value")
	mockCache.AssertExpectations(t)
	
	// Test cache miss
	mockCache.ExpectedCalls = nil
	cacheErr := errors.New("cache miss")
	mockCache.On("Get", ctx, "test:"+testKey, mock.AnythingOfType("*cache_test.TestData")).
		Return(cacheErr)
	
	err = service.Get(ctx, testKey, &result)
	
	assert.Error(t, err, "Get should return error on cache miss")
	assert.Equal(t, cacheErr, err, "Error should be propagated")
	mockCache.AssertExpectations(t)
}

// TestCacheService_Set tests the Set method
func TestCacheService_Set(t *testing.T) {
	mockCache, logger := setupMocks()
	service := cache.NewRedisCacheService(mockCache, logger, "test")
	
	ctx := context.Background()
	testKey := "test:key"
	testData := TestData{ID: "123", Name: "Test", Value: 42}
	expiration := 5 * time.Minute
	
	// Test successful set
	mockCache.On("Set", ctx, "test:"+testKey, testData, expiration).Return(nil)
	
	err := service.Set(ctx, testKey, testData, expiration)
	
	assert.NoError(t, err, "Set should succeed")
	mockCache.AssertExpectations(t)
	
	// Test error
	mockCache.ExpectedCalls = nil
	setErr := errors.New("set error")
	mockCache.On("Set", ctx, "test:"+testKey, testData, expiration).Return(setErr)
	
	err = service.Set(ctx, testKey, testData, expiration)
	
	assert.Error(t, err, "Set should return error")
	assert.Equal(t, setErr, err, "Error should be propagated")
	mockCache.AssertExpectations(t)
}

// TestCacheService_Delete tests the Delete method
func TestCacheService_Delete(t *testing.T) {
	mockCache, logger := setupMocks()
	service := cache.NewRedisCacheService(mockCache, logger, "test")
	
	ctx := context.Background()
	testKey := "test:key"
	
	// Test successful delete
	mockCache.On("Delete", ctx, "test:"+testKey).Return(nil)
	
	err := service.Delete(ctx, testKey)
	
	assert.NoError(t, err, "Delete should succeed")
	mockCache.AssertExpectations(t)
	
	// Test error
	mockCache.ExpectedCalls = nil
	deleteErr := errors.New("delete error")
	mockCache.On("Delete", ctx, "test:"+testKey).Return(deleteErr)
	
	err = service.Delete(ctx, testKey)
	
	assert.Error(t, err, "Delete should return error")
	assert.Equal(t, deleteErr, err, "Error should be propagated")
	mockCache.AssertExpectations(t)
}

// TestCacheService_Exists tests the Exists method
func TestCacheService_Exists(t *testing.T) {
	mockCache, logger := setupMocks()
	service := cache.NewRedisCacheService(mockCache, logger, "test")
	
	ctx := context.Background()
	testKey := "test:key"
	
	// Test key exists
	mockCache.On("Exists", ctx, "test:"+testKey).Return(true, nil)
	
	exists, err := service.Exists(ctx, testKey)
	
	assert.NoError(t, err, "Exists should succeed")
	assert.True(t, exists, "Key should exist")
	mockCache.AssertExpectations(t)
	
	// Test key does not exist
	mockCache.ExpectedCalls = nil
	mockCache.On("Exists", ctx, "test:"+testKey).Return(false, nil)
	
	exists, err = service.Exists(ctx, testKey)
	
	assert.NoError(t, err, "Exists should succeed")
	assert.False(t, exists, "Key should not exist")
	mockCache.AssertExpectations(t)
	
	// Test error
	mockCache.ExpectedCalls = nil
	existsErr := errors.New("exists error")
	mockCache.On("Exists", ctx, "test:"+testKey).Return(false, existsErr)
	
	exists, err = service.Exists(ctx, testKey)
	
	assert.Error(t, err, "Exists should return error")
	assert.False(t, exists, "Exists should return false on error")
	assert.Equal(t, existsErr, err, "Error should be propagated")
	mockCache.AssertExpectations(t)
}

// TestCacheService_Invalidate tests the Invalidate method
func TestCacheService_Invalidate(t *testing.T) {
	mockCache, logger := setupMocks()
	service := cache.NewRedisCacheService(mockCache, logger, "test")
	
	ctx := context.Background()
	pattern := "test:pattern:*"
	
	// Test successful invalidation
	mockCache.On("Invalidate", ctx, "test:"+pattern).Return(nil)
	
	err := service.Invalidate(ctx, pattern)
	
	assert.NoError(t, err, "Invalidate should succeed")
	mockCache.AssertExpectations(t)
	
	// Test error
	mockCache.ExpectedCalls = nil
	invalidateErr := errors.New("invalidate error")
	mockCache.On("Invalidate", ctx, "test:"+pattern).Return(invalidateErr)
	
	err = service.Invalidate(ctx, pattern)
	
	assert.Error(t, err, "Invalidate should return error")
	assert.Equal(t, invalidateErr, err, "Error should be propagated")
	mockCache.AssertExpectations(t)
}

// TestCacheService_GenerateKey tests the GenerateKey method
func TestCacheService_GenerateKey(t *testing.T) {
	mockCache, logger := setupMocks()
	service := cache.NewRedisCacheService(mockCache, logger, "test")
	
	// Test with single part
	key := service.GenerateKey("part1")
	assert.Equal(t, "part1", key, "Single part key should match")
	
	// Test with multiple parts
	key = service.GenerateKey("part1", "part2", "part3")
	assert.Equal(t, "part1:part2:part3", key, "Multiple parts key should match")
	
	// Test with empty parts
	key = service.GenerateKey("", "", "part3")
	assert.Equal(t, "::part3", key, "Empty parts should be included as empty strings")
	
	// Test with no parts
	key = service.GenerateKey()
	assert.Equal(t, "", key, "No parts should return empty string")
}
