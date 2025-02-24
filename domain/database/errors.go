package database

// This error system provides several important features:

// 1. Standard error types: Predefined error constants that can be compared with errors.Is(), making error handling consistent across the application.
// 2. Error wrapping: Uses Go's error wrapping to maintain a chain of errors while adding context, allowing you to use errors.Is() and errors.As() to check error types.
// 3. Rich error context: The DatabaseError type contains fields for operation, table, query, and parameters, making it easier to understand what went wrong.
// 4. Error helpers: Utility functions to check error types without directly comparing error values.

import (
	"errors"
	"fmt"
)

// Standard error types that can be used for error comparisons
var (
	// ErrNotFound indicates that a requested resource wasn't found
	ErrNotFound = errors.New("resource not found")

	// ErrEntryNotFound indicates that a dictionary entry wasn't found
	ErrEntryNotFound = fmt.Errorf("%w: entry not found", ErrNotFound)

	// ErrDuplicateEntry indicates that an entry with the same key already exists
	ErrDuplicateEntry = errors.New("duplicate entry")

	// ErrInvalidInput indicates that the provided input is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrDatabaseConnection indicates a failure to connect to the database
	ErrDatabaseConnection = errors.New("database connection failed")

	// ErrTransactionFailed indicates that a database transaction failed
	ErrTransactionFailed = errors.New("database transaction failed")

	// ErrUnsupportedDriver indicates that the requested database driver is not supported
	ErrUnsupportedDriver = errors.New("unsupported database driver")

	// ErrForeignKeyViolation indicates that a foreign key constraint was violated
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")

	// ErrQueryTimeout indicates that a database query timed out
	ErrQueryTimeout = errors.New("database query timeout")
)

// DatabaseError represents a detailed database error with contextual information
type DatabaseError struct {
	// Original is the underlying error
	Original error

	// Operation is the database operation that was being performed (e.g., "create", "query")
	Operation string

	// Table is the database table involved
	Table string

	// Query is the query being executed (may be empty for security reasons)
	Query string

	// Params contains query parameters or relevant data (may be sanitized)
	Params map[string]interface{}
}

// Error implements the error interface
func (e *DatabaseError) Error() string {
	if e.Original == nil {
		return fmt.Sprintf("database error during %s operation on %s", e.Operation, e.Table)
	}
	return fmt.Sprintf("database error during %s operation on %s: %v", e.Operation, e.Table, e.Original)
}

// Unwrap returns the original error for errors.Is/As compatibility
func (e *DatabaseError) Unwrap() error {
	return e.Original
}

// NewDatabaseError creates a new DatabaseError
func NewDatabaseError(err error, operation, table string) *DatabaseError {
	return &DatabaseError{
		Original:  err,
		Operation: operation,
		Table:     table,
	}
}

// NewQueryError creates a new DatabaseError with query information
func NewQueryError(err error, operation, table, query string, params map[string]interface{}) *DatabaseError {
	return &DatabaseError{
		Original:  err,
		Operation: operation,
		Table:     table,
		Query:     query,
		Params:    params,
	}
}

// IsNotFoundError checks if the error is a not-found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsDuplicateError checks if the error is a duplicate entry error
func IsDuplicateError(err error) bool {
	return errors.Is(err, ErrDuplicateEntry)
}

// IsDatabaseConnectionError checks if the error is a connection-related error
func IsDatabaseConnectionError(err error) bool {
	return errors.Is(err, ErrDatabaseConnection)
}
