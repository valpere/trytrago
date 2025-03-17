package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/valpere/trytrago/domain/errors"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/interface/api/rest/response"
)

// ErrorHandler middleware catches panics and errors
func ErrorHandler(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a logger with request info
		reqLogger := logger.With(
			logging.String("path", c.Request.URL.Path),
			logging.String("method", c.Request.Method),
			logging.String("client_ip", c.ClientIP()),
			logging.String("request_id", GetRequestID(c)),
		)

		// Recover from any panics
		defer func() {
			if err := recover(); err != nil {
				// Log stack trace
				stack := string(debug.Stack())
				reqLogger.Error("panic recovered in API handler",
					logging.String("panic", fmt.Sprintf("%v", err)),
					logging.String("stack", stack),
				)

				// Send 500 error response
				appErr := errors.New(
					errors.ErrInternalServer,
					http.StatusInternalServerError,
					"server_error",
					"An unexpected error occurred",
				)
				
				response.RespondWithError(c, appErr, reqLogger)
				c.Abort()
			}
		}()

		// Process the request
		c.Next()

		// Handle errors set by API handlers
		if len(c.Errors) > 0 {
			// Get the first error
			err := c.Errors[0].Err

			// Log the error
			reqLogger.Error("error occurred during request handling", 
				logging.Error(err),
			)

			// Generate response based on the error
			response.RespondWithError(c, err, reqLogger)
		}
	}
}
