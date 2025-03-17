package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header name for request IDs
	RequestIDHeader = "X-Request-ID"

	// RequestIDContextKey is the context key for request IDs
	RequestIDContextKey = "requestID"
)

// RequestID is a middleware that injects a request ID into the context
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get request ID from header first
		requestID := c.GetHeader(RequestIDHeader)

		// If not found, generate a new one
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context
		c.Set(RequestIDContextKey, requestID)

		// Set request ID in response headers
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDContextKey); exists {
		return requestID.(string)
	}
	return ""
}
