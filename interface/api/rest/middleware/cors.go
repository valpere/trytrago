package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valpere/trytrago/domain/logging"
)

// CORSConfig defines configuration for CORS middleware
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed to make cross-domain requests
	AllowedOrigins []string

	// AllowedMethods is a list of methods that are allowed for cross-domain requests
	AllowedMethods []string

	// AllowedHeaders is a list of headers that are allowed for cross-domain requests
	AllowedHeaders []string

	// ExposedHeaders is a list of headers that are exposed to the client
	ExposedHeaders []string

	// AllowCredentials indicates whether the request can include user credentials
	AllowCredentials bool

	// MaxAge indicates how long the results of a preflight request can be cached
	MaxAge time.Duration
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}

// ProductionCORSConfig returns a stricter CORS configuration suitable for production
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	return CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposedHeaders:   []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           1 * time.Hour,
	}
}

// CORS returns a middleware that handles CORS
func CORS(config CORSConfig, logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request, proceed with the request
			c.Next()
			return
		}

		// Check if the origin is allowed
		allowed := false
		for _, allowedOrigin := range config.AllowedOrigins {
			if allowedOrigin == "*" {
				allowed = true
				break
			}

			if allowedOrigin == origin {
				allowed = true
				break
			}

			// Support for wildcard subdomains
			if strings.HasPrefix(allowedOrigin, "*.") {
				domain := allowedOrigin[2:] // Remove "*."
				if strings.HasSuffix(origin, domain) {
					allowed = true
					break
				}
			}
		}

		if !allowed {
			logger.Warn("blocked CORS request from unauthorized origin",
				logging.String("origin", origin),
				logging.String("path", c.Request.URL.Path),
				logging.String("method", c.Request.Method),
				logging.String("ip", c.ClientIP()),
			)
			// Proceed without setting CORS headers
			c.Next()
			return
		}

		// Set CORS headers
		if origin == "*" && len(config.AllowedOrigins) > 0 {
			// If wildcard is specified but we have specific allowed origins,
			// we respond with the request's origin if it's in the allowed list
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin != "" && origin != "*" {
			// Set the specific origin
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Fallback to the first allowed origin or wildcard
			if len(config.AllowedOrigins) > 0 {
				c.Writer.Header().Set("Access-Control-Allow-Origin", config.AllowedOrigins[0])
			} else {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			}
		}

		// Set other CORS headers
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
		
		if config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			c.Writer.Header().Set("Access-Control-Max-Age", 
				strings.TrimSpace(config.MaxAge.String()))
		}

		// If it's a preflight request, respond with 204 No Content
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// Proceed with the request
		c.Next()
	}
}

// CORSMiddleware creates a CORS middleware with the given configuration
func CORSMiddleware(config CORSConfig, logger logging.Logger) gin.HandlerFunc {
	// Validate configuration
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{"*"}
	}
	
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	}
	
	return CORS(config, logger)
}
