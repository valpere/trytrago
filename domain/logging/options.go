package logging

import (
	"fmt"
)

// Level represents the logging level
type Level string

// Available logging levels
const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
)

// Format represents the logging output format
type Format string

// Available logging formats
const (
	JSONFormat    Format = "json"
	ConsoleFormat Format = "console"
)

// Options contains all configuration for the logger
type Options struct {
	// Level determines the minimum level of logs that will be output
	Level Level

	// Format specifies how the log entries will be formatted
	Format Format

	// FilePath specifies where to write log files
	// If empty, logs will be written to stderr
	FilePath string

	// ServiceName is used to identify this service in logs
	ServiceName string

	// Environment helps distinguish between different deployment environments
	Environment string

	// AddCaller determines if the logger should add caller information
	AddCaller bool
}

// NewDefaultOptions returns Options with sensible defaults
func NewDefaultOptions() *Options {
	return &Options{
		Level:       InfoLevel,
		Format:      JSONFormat,
		ServiceName: "trytrago",
		Environment: "development",
		AddCaller:   true,
	}
}

// Validate checks if the options are valid
func (o *Options) Validate() error {
	// Validate Level
	switch o.Level {
	case DebugLevel, InfoLevel, WarnLevel, ErrorLevel:
		// valid
	default:
		return fmt.Errorf("invalid log level: %s", o.Level)
	}

	// Validate Format
	switch o.Format {
	case JSONFormat, ConsoleFormat:
		// valid
	default:
		return fmt.Errorf("invalid log format: %s", o.Format)
	}

	// Validate ServiceName
	if o.ServiceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	return nil
}
