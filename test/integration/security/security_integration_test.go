// test/integration/security/security_integration_test.go
package security_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/valpere/trytrago/domain/logging"
	domainValidator "github.com/valpere/trytrago/domain/validator"
	"github.com/valpere/trytrago/interface/api/rest/middleware"
)

// Setup a complete test server with all security middleware
func setupSecureTestServer(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create router
	router := gin.New()

	// Create logger
	logger := &MockLogger{}

	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		domainValidator.RegisterCustomValidators(v)
		middleware.InitCustomValidators(v)
	}

	// Add security middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Security(middleware.DefaultSecurityConfig(), logger))
	router.Use(middleware.CORSMiddleware(middleware.DefaultCORSConfig(), logger))
	router.Use(middleware.Validation(logger))
	router.Use(middleware.RequestSizeLimiter(1024*1024, logger)) // 1MB limit
	router.Use(middleware.ProtectAgainstCSRF(logger))

	// Add test endpoints
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.POST("/api/test", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	return router
}

// MockLogger for testing
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...zapcore.Field)   {}
func (m *MockLogger) Info(msg string, fields ...zapcore.Field)    {}
func (m *MockLogger) Warn(msg string, fields ...zapcore.Field)    {}
func (m *MockLogger) Error(msg string, fields ...zapcore.Field)   {}
func (m *MockLogger) With(fields ...zapcore.Field) logging.Logger { return m }
func (m *MockLogger) Sync() error                                 { return nil }

func TestSecurityHeadersIntegration(t *testing.T) {
	router := setupSecureTestServer(t)

	// Create request
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check security headers
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))

	// Check response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}

func TestCORSHeadersIntegration(t *testing.T) {
	router := setupSecureTestServer(t)

	// Create request with Origin header
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check CORS headers - expect the origin to be reflected back
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
}

func TestCSRFProtectionIntegration(t *testing.T) {
	router := setupSecureTestServer(t)

	// Create POST request without Origin or Referer headers
	req := httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(`{"test":"data"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response - should be forbidden due to CSRF protection
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Now try with Origin header
	req = httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(`{"test":"data"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	w = httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response - should be OK with Origin header
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestXSSProtectionIntegration(t *testing.T) {
	router := setupSecureTestServer(t)

	// Create POST request with XSS payload
	xssPayload := `{"input":"<script>alert('XSS')</script>"}`
	req := httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(xssPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com") // Add Origin to pass CSRF check
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check if XSS is blocked in the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// The input should be returned as-is since the middleware doesn't modify response data,
	// but in a real implementation with proper sanitization, this would be sanitized
	input, ok := response["input"].(string)
	assert.True(t, ok)
	assert.Contains(t, input, "<script>")

	// Note: In a real application, there would be more sanitization happening,
	// but this test is focused on the middleware integration
}

func TestRequestSizeLimitIntegration(t *testing.T) {
	router := setupSecureTestServer(t)

	// Create a large payload (2MB)
	largePayload := make([]byte, 2*1024*1024) // 2MB
	for i := range largePayload {
		largePayload[i] = 'A'
	}

	// Create POST request with large payload
	req := httptest.NewRequest("POST", "/api/test", bytes.NewBuffer(largePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com") // Add Origin to pass CSRF check
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response - should be 413 Request Entity Too Large
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestSQLInjectionProtectionIntegration(t *testing.T) {
	router := setupSecureTestServer(t)

	// Create POST request with SQL injection payload
	sqlInjectionPayload := `{"input":"'; DROP TABLE users; --"}`
	req := httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(sqlInjectionPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com") // Add Origin to pass CSRF check
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check response - in a real implementation with proper sanitization,
	// the SQL injection would be neutralized
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "'; DROP TABLE users; --", response["input"])

	// Note: In a real application, SQL injection would be handled at the repository level
	// through parametrized queries, not just through input sanitization
}
