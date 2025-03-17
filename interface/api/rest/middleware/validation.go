package middleware

import (
	"errors"
	"fmt"
	"net/http"
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
							Value:  fmt.Sprintf("%v", fieldErr.Value()),
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
	default:
		return fmt.Sprintf("Failed validation for rule: %s", fieldErr.Tag())
	}
}
