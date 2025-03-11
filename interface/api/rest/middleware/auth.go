package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/infrastructure/auth"
)

// AuthConfig defines configuration for the Auth middleware
type AuthConfig struct {
	JWTSecret    string
	TokenExpiry  int
	AllowedPaths []string // Paths that can be accessed without auth
}

// Auth creates a middleware for authentication using JWT tokens
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check for Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format, expected 'Bearer {token}'"})
			c.Abort()
			return
		}
		tokenString := parts[1]

		// Parse and validate the token
		token, err := auth.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			}
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*auth.CustomClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Convert string ID to UUID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", userID)
		c.Set("username", claims.Username)
		c.Set("userRole", claims.Role)

		// Continue to the next handler
		c.Next()
	}
}

// OptionalAuth middleware that doesn't abort if no token is present
// but adds user information to the context if a valid token exists
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		
		// If no auth header, just continue to the next handler
		if authHeader == "" {
			c.Next()
			return
		}

		// Check for Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format but don't abort
			c.Next()
			return
		}
		tokenString := parts[1]

		// Parse and validate the token
		token, err := auth.ValidateToken(tokenString)
		if err != nil {
			// Invalid token but don't abort
			c.Next()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*auth.CustomClaims)
		if !ok || !token.Valid {
			// Invalid claims but don't abort
			c.Next()
			return
		}

		// Convert string ID to UUID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			// Invalid user ID but don't abort
			c.Next()
			return
		}

		// Set user info in context
		c.Set("userID", userID)
		c.Set("username", claims.Username)
		c.Set("userRole", claims.Role)
		c.Set("authenticated", true)

		// Continue to the next handler
		c.Next()
	}
}

// RequireRole middleware that checks if the user has the required role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by Auth middleware)
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Check if the user has the required role
		if userRole.(string) != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		// Continue to the next handler
		c.Next()
	}
}
