// test/mocks/redis_cache_mock.go
package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockRedisCache is a mock implementation of the infrastructure/cache.Cache interface
type MockRedisCache struct {
	mock.Mock
}

// Get mocks the Get method
func (m *MockRedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

// Set mocks the Set method
func (m *MockRedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockRedisCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Exists mocks the Exists method
func (m *MockRedisCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

// Invalidate mocks the Invalidate method
func (m *MockRedisCache) Invalidate(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

// Close mocks the Close method
func (m *MockRedisCache) Close() error {
	args := m.Called()
	return args.Error(0)
}
