package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is our custom logging interface that wraps zap.Logger
type Logger interface {
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	With(fields ...zapcore.Field) Logger
	Sync() error
}

// logger implements the Logger interface
type logger struct {
	*zap.Logger
}

// newLogger creates a new logger instance with the given options
func NewLogger(opts *Options) (Logger, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger options: %w", err)
	}

	// Create the base configuration
	config := zap.NewProductionConfig()

	// Configure the logging level
	switch opts.Level {
	case DebugLevel:
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case InfoLevel:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case WarnLevel:
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case ErrorLevel:
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}

	// Configure output format
	if opts.Format == ConsoleFormat {
		config.Encoding = "console"
		config.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		config.Encoding = "json"
		config.EncoderConfig = zap.NewProductionEncoderConfig()
	}

	// Configure output destination
	if opts.FilePath != "" {
		// Ensure the log directory exists
		if err := os.MkdirAll(filepath.Dir(opts.FilePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		config.OutputPaths = []string{opts.FilePath}
		config.ErrorOutputPaths = []string{opts.FilePath}
	}

	// Add common fields that will appear in every log entry
	config.InitialFields = map[string]interface{}{
		"service":     opts.ServiceName,
		"environment": opts.Environment,
	}

	// Create the logger
	zapLogger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &logger{zapLogger}, nil
}

// Implementation of Logger interface methods
func (l *logger) With(fields ...zapcore.Field) Logger {
	return &logger{l.Logger.With(fields...)}
}

// Let's also create some helper functions for creating commonly used fields
func String(key, value string) zapcore.Field {
	return zap.String(key, value)
}

func Int(key string, value int) zapcore.Field {
	return zap.Int(key, value)
}

func Int64(key string, value int64) zapcore.Field {
	return zap.Int64(key, value)
}

func Bool(key string, value bool) zapcore.Field {
	return zap.Bool(key, value)
}

func Error(err error) zapcore.Field {
	return zap.Error(err)
}

func Duration(key string, value time.Duration) zapcore.Field {
	return zap.Duration(key, value)
}
