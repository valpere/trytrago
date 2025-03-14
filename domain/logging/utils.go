package logging

import (
	"sync"
	
	"go.uber.org/zap"
)

var (
	defaultLogger Logger
	once          sync.Once
)

// GetLogger returns the default logger instance
// Creates it if it doesn't exist yet
func GetLogger() Logger {
	once.Do(func() {
		opts := NewDefaultOptions()
		var err error
		defaultLogger, err = NewLogger(opts)
		if err != nil {
			// Fall back to a simple logger if there's an error
			zapLogger := zap.NewExample()
			defaultLogger = &logger{
				Logger: zapLogger,
			}
		}
	})
	return defaultLogger
}

// SetDefaultLogger sets the default logger instance
func SetDefaultLogger(l Logger) {
	defaultLogger = l
}
