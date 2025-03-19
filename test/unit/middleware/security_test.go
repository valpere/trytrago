package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/interface/api/rest/middleware"
	"go.uber.org/zap/zapcore"
)

// mockLogger is a simple mock for the logging.Logger interface
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...zapcore.Field)   {}
func (m *mockLogger) Info(msg string, fields ...zapcore.Field)    {}
func (m *mockLogger) Warn(msg string, fields ...zapcore.Field)    {}
func (m *mockLogger) Error(msg string, fields ...zapcore.Field)   {}
func (m *mockLogger) With(fields ...zapcore.Field) logging.Logger { return m }
func (m *mockLogger) Sync() error                                 { return nil }

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestSecurityMiddleware(t *testing.T) {
	// Create a test router
	router := setupRouter()

	// Create a mock logger
	logger := &mockLogger{}

	// Configure security middleware
	config := middleware.DefaultSecurityConfig()
	router.Use(middleware.Security(config, logger))

	// Add a test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())

	// Check security headers
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		method         string
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "No Origin",
			origin:         "",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedHeader: "",
		},
		{
			name:           "Allowed Origin",
			origin:         "https://example.com",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedHeader: "https://example.com",
		},
		{
			name:           "Preflight Request",
			origin:         "https://example.com",
			method:         "OPTIONS",
			expectedStatus: http.StatusNoContent,
			expectedHeader: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test router
			router := setupRouter()

			// Create a mock logger
			logger := &mockLogger{}

			// Configure CORS middleware with an allowed origin
			config := middleware.CORSConfig{
				AllowedOrigins:   []string{"https://example.com"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Origin", "Content-Type", "Accept"},
				ExposedHeaders:   []string{"Content-Length"},
				AllowCredentials: true,
			}
			router.Use(middleware.CORSMiddleware(config, logger))

			// Add a test endpoint
			router.Any("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "success")
			})

			// Create a test request
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check CORS headers
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, w.Header().Get("Access-Control-Allow-Origin"))
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
			}
		})
	}
}

func RequestSizeLimiter(maxSize int64, logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip GET, HEAD, OPTIONS requests
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Set the maximum request size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		// Check Content-Length header as an early indicator
		if c.Request.ContentLength > maxSize {
			logger.Warn("request body too large (detected via Content-Length)",
				logging.String("ip", c.ClientIP()),
				logging.String("method", c.Request.Method),
				logging.String("path", c.Request.URL.Path),
				logging.Int64("content_length", c.Request.ContentLength),
				logging.Int64("max_size", maxSize),
			)

			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":          "Request body too large",
				"max_size_bytes": maxSize,
			})
			return
		}

		c.Next()

		// Check for body read errors after processing
		for _, err := range c.Errors {
			if err.Error() == "http: request body too large" ||
				strings.Contains(err.Error(), "body size exceeds") {
				logger.Warn("request body too large",
					logging.String("ip", c.ClientIP()),
					logging.String("method", c.Request.Method),
					logging.String("path", c.Request.URL.Path),
					logging.Int64("max_size", maxSize),
				)

				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
					"error":          "Request body too large",
					"max_size_bytes": maxSize,
				})
				return
			}
		}
	}
}

func TestValidationMiddleware(t *testing.T) {
	// This is a simplified test as validation middleware is complex
	// Create a test router
	router := setupRouter()

	// Create a mock logger
	logger := &mockLogger{}

	// Add validation middleware
	router.Use(middleware.Validation(logger))

	// Add a test endpoint that causes a validation error
	router.GET("/test", func(c *gin.Context) {
		// Simulate a validation error
		c.Error(gin.Error{
			Err:  &gin.Error{Err: &gin.Error{Err: nil, Meta: nil}, Meta: nil},
			Type: gin.ErrorTypeBind,
			Meta: nil,
		})

		// The middleware should handle the error before this point
		// but we'll set a response in case it doesn't
		c.String(http.StatusOK, "success")
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Since we're simulating a simplified validation error,
	// we're primarily testing that the middleware doesn't crash
	// A real test would need to mock validator.ValidationErrors
}

func TestRequestIDMiddleware(t *testing.T) {
	// Create a test router
	router := setupRouter()

	// Add request ID middleware
	router.Use(middleware.RequestID())

	// Add a test endpoint that checks the request ID
	router.GET("/test", func(c *gin.Context) {
		requestID := middleware.GetRequestID(c)
		c.String(http.StatusOK, requestID)
	})

	// Test cases
	tests := []struct {
		name       string
		requestID  string
		shouldHave bool
	}{
		{
			name:       "No Request ID",
			requestID:  "",
			shouldHave: true, // Middleware should generate one
		},
		{
			name:       "With Request ID",
			requestID:  "test-request-id",
			shouldHave: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.requestID != "" {
				req.Header.Set("X-Request-ID", tt.requestID)
			}
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, http.StatusOK, w.Code)

			// Check request ID in response headers
			responseID := w.Header().Get("X-Request-ID")
			if tt.shouldHave {
				assert.NotEmpty(t, responseID)

				// If we provided a request ID, it should be the same
				if tt.requestID != "" {
					assert.Equal(t, tt.requestID, responseID)
				}

				// Response body should contain the request ID
				assert.Equal(t, responseID, w.Body.String())
			} else {
				assert.Empty(t, responseID)
			}
		})
	}
}
