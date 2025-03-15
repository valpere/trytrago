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
	// Don't try to match specific calls - this causes problems with variable args
	m.Called()
}

// Info implements the Logger.Info method
func (m *MockLogger) Info(msg string, fields ...zapcore.Field) {
	// Don't try to match specific calls - this causes problems with variable args
	m.Called()
}

// Warn implements the Logger.Warn method
func (m *MockLogger) Warn(msg string, fields ...zapcore.Field) {
	// Don't try to match specific calls - this causes problems with variable args
	m.Called()
}

// Error implements the Logger.Error method
func (m *MockLogger) Error(msg string, fields ...zapcore.Field) {
	// Don't try to match specific calls - this causes problems with variable args
	m.Called()
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
