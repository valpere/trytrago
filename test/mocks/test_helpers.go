package mocks

import (
	"github.com/stretchr/testify/mock"
)

// SetupLoggerMock configures a MockLogger with default expectations for common methods
func SetupLoggerMock() *MockLogger {
	mockLogger := new(MockLogger)

	// Setup common methods that might be called in most tests
	mockLogger.On("With", mock.Anything).Return(mockLogger)

	// Configure various log level methods with any number of arguments
	// Using AtLeast(0) means the call is optional
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockLogger.On("Info", mock.Anything).Return().Maybe()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockLogger.On("Warn", mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockLogger.On("Error", mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	return mockLogger
}

// VerifyLoggerMock verifies that all expected logger calls were made
func VerifyLoggerMock(mockLogger *MockLogger, t mock.TestingT) {
	if mockLogger != nil {
		mockLogger.AssertExpectations(t)
	}
}

// ExpectDebug adds a specific expectation for a Debug log with a message pattern
func ExpectDebug(mockLogger *MockLogger, msgPattern string) *mock.Call {
	return mockLogger.On("Debug", mock.MatchedBy(func(msg string) bool {
		return msg == msgPattern
	}), mock.Anything)
}

// ExpectInfo adds a specific expectation for an Info log with a message pattern
func ExpectInfo(mockLogger *MockLogger, msgPattern string) *mock.Call {
	return mockLogger.On("Info", mock.MatchedBy(func(msg string) bool {
		return msg == msgPattern
	}), mock.Anything)
}

// ExpectWarn adds a specific expectation for a Warn log with a message pattern
func ExpectWarn(mockLogger *MockLogger, msgPattern string) *mock.Call {
	return mockLogger.On("Warn", mock.MatchedBy(func(msg string) bool {
		return msg == msgPattern
	}), mock.Anything)
}

// ExpectError adds a specific expectation for an Error log with a message pattern
func ExpectError(mockLogger *MockLogger, msgPattern string) *mock.Call {
	return mockLogger.On("Error", mock.MatchedBy(func(msg string) bool {
		return msg == msgPattern
	}), mock.Anything)
}
