package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valpere/trytrago/domain/logging"
)

// responseWriter is a custom gin.ResponseWriter that captures the response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body
func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logger creates a middleware that logs HTTP requests
func Logger(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Read request body
		var requestBody []byte
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			if c.Request.Body != nil {
				// Read request body
				requestBodyBytes, err := io.ReadAll(c.Request.Body)
				if err == nil {
					requestBody = requestBodyBytes
					// Restore request body
					c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))
				}
			}
		}

		// Create custom writer to capture response
		responseBodyBuffer := &bytes.Buffer{}
		blw := responseWriter{
			ResponseWriter: c.Writer,
			body:           responseBodyBuffer,
		}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Calculate request size
		requestSize := c.Request.ContentLength

		// Get status
		status := c.Writer.Status()

		// Log request details
		switch {
		case status >= 500:
			// Server error
			logger.Error("server error",
				logging.String("method", c.Request.Method),
				logging.String("path", path),
				logging.String("query", raw),
				logging.String("ip", c.ClientIP()),
				logging.Int("status", status),
				logging.Int64("size", requestSize),
				logging.Duration("duration", duration),
				logging.String("error", c.Errors.String()),
			)
			// Optionally log request body for debugging, but be careful with sensitive data
			if len(requestBody) > 0 && len(requestBody) < 10000 {
				logger.Debug("request body",
					logging.String("body", string(requestBody)),
				)
			}
			// Log response for debugging
			if responseBodyBuffer.Len() > 0 && responseBodyBuffer.Len() < 10000 {
				logger.Debug("response body",
					logging.String("body", responseBodyBuffer.String()),
				)
			}
		case status >= 400:
			// Client error
			logger.Warn("client error",
				logging.String("method", c.Request.Method),
				logging.String("path", path),
				logging.String("query", raw),
				logging.String("ip", c.ClientIP()),
				logging.Int("status", status),
				logging.Int64("size", requestSize),
				logging.Duration("duration", duration),
				logging.String("error", c.Errors.String()),
			)
		default:
			// Success
			logger.Info("request completed",
				logging.String("method", c.Request.Method),
				logging.String("path", path),
				logging.String("query", raw),
				logging.String("ip", c.ClientIP()),
				logging.Int("status", status),
				logging.Int64("size", requestSize),
				logging.Duration("duration", duration),
			)
		}
	}
}

// Recovery creates a middleware that recovers from panics
func Recovery(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger.Error("request handler panic",
					logging.String("method", c.Request.Method),
					logging.String("path", c.Request.URL.Path),
					logging.String("error", err.(error).Error()),
				)

				// Return 500 error
				c.AbortWithStatus(500)
			}
		}()

		c.Next()
	}
}
