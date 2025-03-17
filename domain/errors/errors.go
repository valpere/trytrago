package errors

import (
	"errors"
	"fmt"
)

// Standard error types for the application
var (
	// ErrNotFound indicates a resource wasn't found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput indicates invalid input parameters
	ErrInvalidInput = errors.New("invalid input")

	// ErrDuplicate indicates a conflict with existing data
	ErrDuplicate = errors.New("resource already exists")

	// ErrUnauthorized indicates lack of valid authentication
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates lack of sufficient permissions
	ErrForbidden = errors.New("forbidden")

	// ErrInternalServer indicates an unexpected server-side error
	ErrInternalServer = errors.New("internal server error")

	// ErrBadRequest indicates a malformed request
	ErrBadRequest = errors.New("bad request")

	// ErrValidation indicates validation failures
	ErrValidation = errors.New("validation error")

	// ErrUnavailable indicates a service is currently unavailable
	ErrUnavailable = errors.New("service unavailable")
)

// Domain-specific errors - derive from standard types
var (
	// Entry errors
	ErrEntryNotFound = fmt.Errorf("%w: entry not found", ErrNotFound)
	ErrDuplicateEntry = fmt.Errorf("%w: duplicate entry", ErrDuplicate)

	// Meaning errors
	ErrMeaningNotFound = fmt.Errorf("%w: meaning not found", ErrNotFound)

	// Translation errors
	ErrTranslationNotFound = fmt.Errorf("%w: translation not found", ErrNotFound)

	// User errors
	ErrUserNotFound = fmt.Errorf("%w: user not found", ErrNotFound)
	ErrDuplicateUsername = fmt.Errorf("%w: username already exists", ErrDuplicate)
	ErrDuplicateEmail = fmt.Errorf("%w: email already exists", ErrDuplicate)
	ErrInvalidCredentials = fmt.Errorf("%w: invalid username or password", ErrUnauthorized)
	ErrInvalidToken = fmt.Errorf("%w: invalid authentication token", ErrUnauthorized)
	ErrExpiredToken = fmt.Errorf("%w: expired authentication token", ErrUnauthorized)
	ErrInsufficientPermissions = fmt.Errorf("%w: insufficient permissions", ErrForbidden)

	// Comment/social errors
	ErrCommentNotFound = fmt.Errorf("%w: comment not found", ErrNotFound)
	ErrLikeNotFound = fmt.Errorf("%w: like not found", ErrNotFound)
)

// AppError represents a structured application error
type AppError struct {
	// Original is the underlying error
	Original error

	// Code is the HTTP status code
	Code int

	// Type is the error type (for categorization)
	Type string

	// Message is a human-readable error message
	Message string

	// Details contains additional error information
	Details map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Original)
	}
	return e.Message
}

// Unwrap returns the original error
func (e *AppError) Unwrap() error {
	return e.Original
}

// New creates a new AppError
func New(original error, code int, errType, message string) *AppError {
	return &AppError{
		Original: original,
		Code:     code,
		Type:     errType,
		Message:  message,
		Details:  make(map[string]interface{}),
	}
}

// NewWithDetails creates a new AppError with details
func NewWithDetails(original error, code int, errType, message string, details map[string]interface{}) *AppError {
	return &AppError{
		Original: original,
		Code:     code,
		Type:     errType,
		Message:  message,
		Details:  details,
	}
}

// AddDetail adds a detail to an existing AppError
func (e *AppError) AddDetail(key string, value interface{}) *AppError {
	e.Details[key] = value
	return e
}

// WithDetails returns a copy of the error with details
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	result := &AppError{
		Original: e.Original,
		Code:     e.Code,
		Type:     e.Type,
		Message:  e.Message,
		Details:  make(map[string]interface{}),
	}

	// Copy existing details
	for k, v := range e.Details {
		result.Details[k] = v
	}

	// Add new details
	for k, v := range details {
		result.Details[k] = v
	}

	return result
}

// Is checks if an error is of a specific error type (compatibility with errors.Is)
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// GetErrorCode returns the HTTP status code for an error or 500 if not classifiable
func GetErrorCode(err error) int {
	// Check if it's already an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}

	// Match against known error types
	switch {
	case errors.Is(err, ErrNotFound):
		return 404
	case errors.Is(err, ErrInvalidInput) || errors.Is(err, ErrBadRequest) || errors.Is(err, ErrValidation):
		return 400
	case errors.Is(err, ErrUnauthorized):
		return 401
	case errors.Is(err, ErrForbidden):
		return 403
	case errors.Is(err, ErrDuplicate):
		return 409
	case errors.Is(err, ErrUnavailable):
		return 503
	default:
		return 500
	}
}

// GetErrorType returns a string representation of the error type
func GetErrorType(err error) string {
	// Check if it's already an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type
	}

	// Match against known error types
	switch {
	case errors.Is(err, ErrNotFound):
		return "not_found"
	case errors.Is(err, ErrInvalidInput):
		return "invalid_input"
	case errors.Is(err, ErrBadRequest):
		return "bad_request"
	case errors.Is(err, ErrValidation):
		return "validation_error"
	case errors.Is(err, ErrUnauthorized):
		return "unauthorized"
	case errors.Is(err, ErrForbidden):
		return "forbidden"
	case errors.Is(err, ErrDuplicate):
		return "conflict"
	case errors.Is(err, ErrUnavailable):
		return "service_unavailable"
	default:
		return "internal_server_error"
	}
}

// GetErrorMessage returns a user-friendly error message
func GetErrorMessage(err error) string {
	// Check if it's an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}

	// Use the error string but remove implementation details
	return errors.Unwrap(err).Error()
}

// ToAppError converts any error to an AppError with appropriate HTTP status code
func ToAppError(err error) *AppError {
	// Check if error is already an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Convert standard error to AppError
	code := GetErrorCode(err)
	errType := GetErrorType(err)
	message := err.Error()

	return &AppError{
		Original: err,
		Code:     code,
		Type:     errType,
		Message:  message,
		Details:  make(map[string]interface{}),
	}
}

// IsNotFoundError checks if an error is a "not found" error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsDuplicateError checks if an error is a duplicate/conflict error
func IsDuplicateError(err error) bool {
	return errors.Is(err, ErrDuplicate)
}

// IsUnauthorizedError checks if an error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbiddenError checks if an error is a forbidden error
func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation)
}
