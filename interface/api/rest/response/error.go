package response

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valpere/trytrago/domain/errors"
	"github.com/valpere/trytrago/domain/logging"
	"go.uber.org/zap/zapcore"
)

// ErrorResponse defines the standard API error response format
type ErrorResponse struct {
	Status    int                    `json:"status"`               // HTTP status code
	Error     string                 `json:"error"`                // Error type
	Message   string                 `json:"message"`              // User-friendly error message
	Details   map[string]interface{} `json:"details,omitempty"`    // Optional error details
	RequestID string                 `json:"request_id,omitempty"` // Request ID for tracing
	Timestamp time.Time              `json:"timestamp"`            // Timestamp of the error
}

// NewErrorResponse creates a new error response from an error
func NewErrorResponse(err error, requestID string) *ErrorResponse {
	// Convert to an AppError to get status code and type
	appErr := errors.ToAppError(err)

	return &ErrorResponse{
		Status:    appErr.Code,
		Error:     appErr.Type,
		Message:   appErr.Message,
		Details:   appErr.Details,
		RequestID: requestID,
		Timestamp: time.Now().UTC(),
	}
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, err error, logger logging.Logger) {
	// Get request ID from context
	requestID, exists := c.Get("requestID")
	requestIDStr := ""
	if exists {
		requestIDStr = requestID.(string)
	}

	// Create error response
	resp := NewErrorResponse(err, requestIDStr)

	// Log the error with details
	logFields := []zapcore.Field{
		logging.String("request_id", requestIDStr),
		logging.Int("status", resp.Status),
		logging.String("error_type", resp.Error),
		logging.String("client_ip", c.ClientIP()),
		logging.String("method", c.Request.Method),
		logging.String("path", c.Request.URL.Path),
		logging.Error(err),
	}

	// Adjust log level based on status code
	switch {
	case resp.Status >= 500:
		logger.Error("server error", logFields...)
	case resp.Status >= 400:
		logger.Warn("client error", logFields...)
	default:
		logger.Info("error response", logFields...)
	}

	// Send response
	c.JSON(resp.Status, resp)
}
