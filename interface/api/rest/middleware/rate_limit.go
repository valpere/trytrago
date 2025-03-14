package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valpere/trytrago/domain/logging"
	"golang.org/x/time/rate"
)

// ClientRateLimiter represents rate limiting configuration for a client
type ClientRateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter middleware that limits request rate per client IP
func RateLimiter(logger logging.Logger, cfg RateLimiterConfig) gin.HandlerFunc {
	// Use default values if not provided
	if cfg.RequestsPerSecond <= 0 {
		cfg.RequestsPerSecond = 10
	}
	if cfg.Burst <= 0 {
		cfg.Burst = 20
	}
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = 5 * time.Minute
	}
	if cfg.ClientTimeout <= 0 {
		cfg.ClientTimeout = 10 * time.Minute
	}

	// Create a map of client limiters
	var (
		limiters = make(map[string]*ClientRateLimiter)
		mu       sync.Mutex // Mutex to protect the map
	)

	// Start a goroutine to periodically clean up old limiters
	go func() {
		ticker := time.NewTicker(cfg.CleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			for ip, client := range limiters {
				if time.Since(client.lastSeen) > cfg.ClientTimeout {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	// Return the middleware function
	return func(c *gin.Context) {
		// Get client IP
		ip := c.ClientIP()

		// Get or create limiter for this client
		mu.Lock()
		client, exists := limiters[ip]
		if !exists {
			client = &ClientRateLimiter{
				limiter:  rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.Burst),
				lastSeen: time.Now(),
			}
			limiters[ip] = client
		} else {
			client.lastSeen = time.Now()
		}
		mu.Unlock()

		// Check if request is allowed
		if !client.limiter.Allow() {
			// Log the rate limit exceeded
			logger.Warn("rate limit exceeded",
				logging.String("ip", ip),
				logging.String("path", c.Request.URL.Path),
				logging.String("method", c.Request.Method),
			)

			// Return 429 Too Many Requests
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		// Request is allowed, continue
		c.Next()
	}
}
