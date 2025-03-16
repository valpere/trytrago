package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/infrastructure/auth"

	"time"
)

// RateLimiterConfig defines configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerSecond int           // Number of requests allowed per second
	Burst             int           // Maximum burst size
	CleanupInterval   time.Duration // Interval to clean up old limiters
	ClientTimeout     time.Duration // Time after which a client is considered inactive
}

// jwtAuthMiddleware implements AuthMiddleware using JWT
type jwtAuthMiddleware struct {
	logger logging.Logger
}

// RequireAuth implements middleware that requires authentication
func (m *jwtAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract and validate token
		token, err := m.extractToken(c)
		if err != nil {
			m.logger.Debug("Authentication failed", logging.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Parse user ID string to UUID
		userID, err := uuid.Parse(token.UserID)
		if err != nil {
			m.logger.Error("Invalid user ID in token", logging.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
			return
		}

		// Store user info in context
		c.Set("userID", userID)
		c.Set("username", token.Username)
		c.Set("userRole", token.Role)
		c.Set("authenticated", true)

		c.Next()
	}
}

// RequireAdmin implements middleware that requires admin privileges
func (m *jwtAuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First require authentication
		token, err := m.extractToken(c)
		if err != nil {
			m.logger.Debug("Authentication failed", logging.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Parse user ID string to UUID
		userID, err := uuid.Parse(token.UserID)
		if err != nil {
			m.logger.Error("Invalid user ID in token", logging.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
			return
		}

		// Store user info in context
		c.Set("userID", userID)
		c.Set("username", token.Username)
		c.Set("userRole", token.Role)
		c.Set("authenticated", true)

		// Check role
		if token.Role != "ADMIN" {
			m.logger.Debug("Admin access denied",
				logging.String("role", token.Role),
				logging.String("path", c.Request.URL.Path),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		c.Next()
	}
}

// OptionalAuth implements middleware that makes authentication optional
func (m *jwtAuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to extract token but don't abort on failure
		token, err := m.extractToken(c)
		if err == nil {
			// Parse user ID string to UUID
			userID, err := uuid.Parse(token.UserID)
			if err == nil {
				// Store user info in context
				c.Set("userID", userID)
				c.Set("username", token.Username)
				c.Set("userRole", token.Role)
				c.Set("authenticated", true)
			} else {
				// Mark as unauthenticated
				c.Set("authenticated", false)
			}
		} else {
			// Mark as unauthenticated
			c.Set("authenticated", false)
		}

		c.Next()
	}
}

// extractToken extracts and validates the JWT token from the request
func (m *jwtAuthMiddleware) extractToken(c *gin.Context) (*auth.TokenClaims, error) {
	// Get Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	// Check Bearer format
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("invalid authorization format")
	}

	// Extract token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return nil, errors.New("empty token")
	}

	// Validate token
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
