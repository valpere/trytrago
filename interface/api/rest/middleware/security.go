package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/valpere/trytrago/domain/logging"
)

// SecurityConfig defines configuration for security middleware
type SecurityConfig struct {
	// XSSProtection enables X-XSS-Protection header
	XSSProtection bool

	// ContentTypeNosniff enables X-Content-Type-Options header
	ContentTypeNosniff bool

	// XFrameOptions sets the X-Frame-Options header
	XFrameOptions string

	// ContentSecurityPolicy sets the Content-Security-Policy header
	ContentSecurityPolicy string

	// ReferrerPolicy sets the Referrer-Policy header
	ReferrerPolicy string

	// StrictTransportSecurity sets the Strict-Transport-Security header
	StrictTransportSecurity string

	// PermissionsPolicy sets the Permissions-Policy header
	PermissionsPolicy string
}

// DefaultSecurityConfig returns a default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		XSSProtection:         true,
		ContentTypeNosniff:    true,
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=()",
	}
}

// Security returns a middleware that adds security headers
func Security(config SecurityConfig, logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-XSS-Protection
		if config.XSSProtection {
			c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		}

		// X-Content-Type-Options
		if config.ContentTypeNosniff {
			c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		}

		// X-Frame-Options
		if config.XFrameOptions != "" {
			c.Writer.Header().Set("X-Frame-Options", config.XFrameOptions)
		}

		// Content-Security-Policy
		if config.ContentSecurityPolicy != "" {
			c.Writer.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
		}

		// Referrer-Policy
		if config.ReferrerPolicy != "" {
			c.Writer.Header().Set("Referrer-Policy", config.ReferrerPolicy)
		}

		// Strict-Transport-Security (HSTS)
		// Only set this header over HTTPS connections
		if config.StrictTransportSecurity != "" && c.Request.TLS != nil {
			c.Writer.Header().Set("Strict-Transport-Security", config.StrictTransportSecurity)
		}

		// Permissions-Policy
		if config.PermissionsPolicy != "" {
			c.Writer.Header().Set("Permissions-Policy", config.PermissionsPolicy)
		}

		// Cache-Control for API endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Writer.Header().Set("Pragma", "no-cache")
			c.Writer.Header().Set("Expires", "0")
		}

		c.Next()
	}
}

// ProtectAgainstCSRF returns a middleware that protects against CSRF attacks
func ProtectAgainstCSRF(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to mutating methods
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Method != "OPTIONS" {
			// Check for the Origin or Referer header
			origin := c.Request.Header.Get("Origin")
			referer := c.Request.Header.Get("Referer")
			
			// If neither is present for a mutating request, it's potentially suspicious
			if origin == "" && referer == "" {
				logger.Warn("potential CSRF attempt: missing origin and referer headers",
					logging.String("path", c.Request.URL.Path),
					logging.String("method", c.Request.Method),
					logging.String("ip", c.ClientIP()),
				)
				
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "Forbidden: potential CSRF attempt",
				})
				return
			}
		}
		
		c.Next()
	}
}

// PreventClickjacking adds X-Frame-Options header to prevent clickjacking
func PreventClickjacking() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Next()
	}
}

// PreventMimeSniffing adds X-Content-Type-Options header to prevent MIME sniffing
func PreventMimeSniffing() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Next()
	}
}

// XSSProtection adds X-XSS-Protection header to prevent XSS attacks in older browsers
func XSSProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Next()
	}
}

// ContentSecurityPolicy adds CSP header to prevent various attacks
func ContentSecurityPolicy(policy string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use a default policy if none is provided
		if policy == "" {
			policy = "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'"
		}
		
		c.Writer.Header().Set("Content-Security-Policy", policy)
		c.Next()
	}
}

// ReferrerPolicy controls what information is sent in the Referer header
func ReferrerPolicy(policy string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use a default policy if none is provided
		if policy == "" {
			policy = "strict-origin-when-cross-origin"
		}
		
		c.Writer.Header().Set("Referrer-Policy", policy)
		c.Next()
	}
}

// SecureHeadersBundle adds all common security headers in one middleware
func SecureHeadersBundle() gin.HandlerFunc {
	config := DefaultSecurityConfig()
	
	return func(c *gin.Context) {
		// X-XSS-Protection
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// X-Content-Type-Options
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		
		// X-Frame-Options
		c.Writer.Header().Set("X-Frame-Options", config.XFrameOptions)
		
		// Content-Security-Policy
		c.Writer.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
		
		// Referrer-Policy
		c.Writer.Header().Set("Referrer-Policy", config.ReferrerPolicy)
		
		// Strict-Transport-Security
		if c.Request.TLS != nil {
			c.Writer.Header().Set("Strict-Transport-Security", config.StrictTransportSecurity)
		}
		
		c.Next()
	}
}
