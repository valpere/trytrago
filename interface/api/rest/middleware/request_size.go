package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/valpere/trytrago/domain/logging"
)

// RequestSizeLimiter limits the size of incoming requests
func RequestSizeLimiter(maxSize int64, logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip GET, HEAD, OPTIONS requests as they don't have a body
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

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

		// Set the maximum request size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		// Continue processing the request
		c.Next()

		// Check if there was a request entity too large error
		for _, err := range c.Errors {
			if strings.Contains(err.Error(), "http: request body too large") {
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
