package middleware

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	domainErrors "github.com/valpere/trytrago/domain/errors"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/interface/api/rest/response"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field  string `json:"field"`
	Rule   string `json:"rule"`
	Value  string `json:"value,omitempty"`
	Reason string `json:"reason"`
}

// Validation middleware for request validation
func Validation(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize input parameters before binding
		sanitizeQueryParams(c)
		sanitizePathParams(c)

		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Filter validation errors
			var validationErrors []ValidationError

			for _, err := range c.Errors {
				// Check if it's a validator.ValidationErrors
				var valErrs validator.ValidationErrors
				if errors.As(err.Err, &valErrs) {
					for _, fieldErr := range valErrs {
						validationError := ValidationError{
							Field:  fieldErr.Field(),
							Rule:   fieldErr.Tag(),
							Value:  sanitizeValue(fmt.Sprintf("%v", fieldErr.Value())),
							Reason: getValidationErrorMessage(fieldErr),
						}
						validationErrors = append(validationErrors, validationError)
					}
				}
			}

			// Handle validation errors differently
			if len(validationErrors) > 0 {
				// Create fields map for error response
				fields := make(map[string]string)
				for _, ve := range validationErrors {
					fields[ve.Field] = ve.Reason
				}

				// Create validation error
				appErr := domainErrors.NewWithDetails(
					domainErrors.ErrValidation,
					http.StatusBadRequest,
					"validation_error",
					"Validation failed for the request",
					map[string]interface{}{
						"fields": fields,
					},
				)

				// Log validation error
				logger.Warn("validation failed",
					logging.String("path", c.Request.URL.Path),
					logging.String("method", c.Request.Method),
					logging.String("client_ip", c.ClientIP()),
					logging.String("request_id", GetRequestID(c)),
				)

				// Respond with validation error
				response.RespondWithError(c, appErr, logger)
				c.Abort()
			}
		}
	}
}

// getValidationErrorMessage returns a user-friendly error message for a validation error
func getValidationErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		if strings.Contains(fieldErr.Type().String(), "string") {
			return fmt.Sprintf("Must be at least %s characters long", fieldErr.Param())
		}
		return fmt.Sprintf("Must be at least %s", fieldErr.Param())
	case "max":
		if strings.Contains(fieldErr.Type().String(), "string") {
			return fmt.Sprintf("Must be at most %s characters long", fieldErr.Param())
		}
		return fmt.Sprintf("Must be at most %s", fieldErr.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", fieldErr.Param())
	case "uuid":
		return "Must be a valid UUID"
	case "alphanum":
		return "Must contain only alphanumeric characters"
	case "alphaunicode":
		return "Must contain only alphabetic characters"
	case "ascii":
		return "Must contain only ASCII characters"
	case "boolean":
		return "Must be a boolean value"
	case "numeric":
		return "Must be a numeric value"
	case "url":
		return "Must be a valid URL"
	default:
		return fmt.Sprintf("Failed validation for rule: %s", fieldErr.Tag())
	}
}

// sanitizeQueryParams sanitizes query parameters
func sanitizeQueryParams(c *gin.Context) {
	// Get all query parameters
	queryParams := c.Request.URL.Query()
	
	// Create a new sanitized query values
	sanitizedQuery := make(map[string][]string)
	
	// Sanitize each query parameter
	for key, values := range queryParams {
		sanitizedValues := make([]string, len(values))
		for i, value := range values {
			sanitizedValues[i] = sanitizeInput(value)
		}
		sanitizedQuery[key] = sanitizedValues
	}
	
	// Update the request with sanitized values
	for key, values := range sanitizedQuery {
		c.Request.URL.Query()[key] = values
	}
}

// sanitizePathParams sanitizes path parameters
func sanitizePathParams(c *gin.Context) {
	// Get all path parameters
	params := c.Params
	
	// Sanitize each path parameter
	for _, param := range params {
		// We can't modify gin.Params directly, so we set a context value with the sanitized version
		sanitizedValue := sanitizeInput(param.Value)
		c.Set("sanitized_"+param.Key, sanitizedValue)
	}
}

// sanitizeInput sanitizes input string to prevent XSS and injection attacks
func sanitizeInput(input string) string {
	// Escape HTML special characters
	sanitized := html.EscapeString(input)
	
	// Remove any HTML tags that might have slipped through
	sanitized = stripTags(sanitized)
	
	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)
	
	return sanitized
}

// stripTags removes any HTML or script tags from the input
func stripTags(input string) string {
	// Remove HTML tags
	tagPattern := regexp.MustCompile(`<[^>]*>`)
	return tagPattern.ReplaceAllString(input, "")
}

// sanitizeValue sanitizes a value for display in error messages
func sanitizeValue(value string) string {
	// For sensitive fields, redact the value
	if strings.Contains(strings.ToLower(value), "password") ||
		strings.Contains(strings.ToLower(value), "token") ||
		strings.Contains(strings.ToLower(value), "secret") {
		return "[REDACTED]"
	}
	
	// For other values, sanitize to prevent XSS
	return sanitizeInput(value)
}

// GetSanitizedParam gets a sanitized path parameter value
func GetSanitizedParam(c *gin.Context, paramName string) string {
	// First check if we have a sanitized version in the context
	if sanitized, exists := c.Get("sanitized_" + paramName); exists {
		return sanitized.(string)
	}
	
	// If not, sanitize the raw parameter value
	return sanitizeInput(c.Param(paramName))
}

// Custom validator functions

// ValidateAlphaNumericWithDash validates that a string contains only alphanumeric characters and dashes
func ValidateAlphaNumericWithDash(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	match, _ := regexp.MatchString("^[a-zA-Z0-9-]+$", value)
	return match
}

// ValidateNoSQL validates that a string doesn't contain SQL injection attempts
func ValidateNoSQL(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// Check for common SQL injection patterns
	sqlPatterns := []string{
		"--",
		";",
		"/*",
		"*/",
		"UNION",
		"SELECT",
		"DROP",
		"DELETE",
		"UPDATE",
		"INSERT",
		"EXEC",
	}
	
	// Convert to lowercase for case-insensitive matching
	valueLower := strings.ToLower(value)
	
	for _, pattern := range sqlPatterns {
		if strings.Contains(valueLower, strings.ToLower(pattern)) {
			return false
		}
	}
	
	return true
}

// ValidateNoScript validates that a string doesn't contain script injection attempts
func ValidateNoScript(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	
	// Check for common script injection patterns
	scriptPatterns := []string{
		"<script",
		"</script",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
		"alert(",
		"eval(",
	}
	
	// Convert to lowercase for case-insensitive matching
	valueLower := strings.ToLower(value)
	
	for _, pattern := range scriptPatterns {
		if strings.Contains(valueLower, strings.ToLower(pattern)) {
			return false
		}
	}
	
	return true
}

// InitCustomValidators initializes custom validators
func InitCustomValidators(v *validator.Validate) {
	v.RegisterValidation("alphanumdash", ValidateAlphaNumericWithDash)
	v.RegisterValidation("nosql", ValidateNoSQL)
	v.RegisterValidation("noscript", ValidateNoScript)
}
