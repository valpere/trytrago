package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/valpere/trytrago/domain/logging"
)

// AuthMiddleware defines the interface for authentication middleware
type AuthMiddleware interface {
    // RequireAuth returns middleware that requires authentication
    RequireAuth() gin.HandlerFunc

    // RequireAdmin returns middleware that requires admin privileges
    RequireAdmin() gin.HandlerFunc

    // OptionalAuth returns middleware that makes authentication optional
    OptionalAuth() gin.HandlerFunc
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(logger logging.Logger) AuthMiddleware {
    return &jwtAuthMiddleware{
        logger: logger.With(logging.String("component", "auth_middleware")),
    }
}
