package mocks

import (
    "github.com/stretchr/testify/mock"
    "github.com/valpere/trytrago/domain/logging"
    "go.uber.org/zap/zapcore"
)

// MockLogger is a mock implementation of the Logger interface
type MockLogger struct {
    mock.Mock
}

// Debug implements the Logger.Debug method
func (m *MockLogger) Debug(msg string, fields ...zapcore.Field) {
    // Create a slice of interface{} for variadic arguments to the mock
    args := make([]interface{}, 0, len(fields)+1)
    args = append(args, msg)
    for _, field := range fields {
        args = append(args, field)
    }
    m.Called(args...)
}

// Info implements the Logger.Info method
func (m *MockLogger) Info(msg string, fields ...zapcore.Field) {
    // Create a slice of interface{} for variadic arguments to the mock
    args := make([]interface{}, 0, len(fields)+1)
    args = append(args, msg)
    for _, field := range fields {
        args = append(args, field)
    }
    m.Called(args...)
}

// Warn implements the Logger.Warn method
func (m *MockLogger) Warn(msg string, fields ...zapcore.Field) {
    // Create a slice of interface{} for variadic arguments to the mock
    args := make([]interface{}, 0, len(fields)+1)
    args = append(args, msg)
    for _, field := range fields {
        args = append(args, field)
    }
    m.Called(args...)
}

// Error implements the Logger.Error method
func (m *MockLogger) Error(msg string, fields ...zapcore.Field) {
    // Create a slice of interface{} for variadic arguments to the mock
    args := make([]interface{}, 0, len(fields)+1)
    args = append(args, msg)
    for _, field := range fields {
        args = append(args, field)
    }
    m.Called(args...)
}

// With implements the Logger.With method
func (m *MockLogger) With(fields ...zapcore.Field) logging.Logger {
    args := m.Called(fields)
    return args.Get(0).(logging.Logger)
}

// Sync implements the Logger.Sync method
func (m *MockLogger) Sync() error {
    args := m.Called()
    return args.Error(0)
}
