package rest

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/valpere/trytrago/domain"
	"github.com/valpere/trytrago/domain/logging"
	domainValidator "github.com/valpere/trytrago/domain/validator"
	"github.com/valpere/trytrago/interface/api/rest/docs"
	"github.com/valpere/trytrago/interface/api/rest/handler"
	"github.com/valpere/trytrago/interface/api/rest/middleware"
)

// NewRouterWithErrorHandling creates a Router with improved error handling
// and security enhancements
func NewRouterWithErrorHandling(
	config domain.Config,
	logger logging.Logger,
	entryHandler handler.EntryHandlerInterface,
	translationHandler handler.TranslationHandlerInterface,
	userHandler handler.UserHandlerInterface,
	authMiddleware middleware.AuthMiddleware,
) Router {
	// Set Gin mode based on environment
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router with middleware
	router := gin.New()

	// Register custom validators to Gin's validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		domainValidator.RegisterCustomValidators(v)
		middleware.InitCustomValidators(v)
	}

	// Configure CORS based on environment
	var corsConfig middleware.CORSConfig
	if config.Environment == "production" {
		// In production, use a stricter CORS policy with allowed origins
		allowedOrigins := []string{
			"https://trytrago.com",
			"https://www.trytrago.com",
			"https://api.trytrago.com",
		}
		// Add any additional origins from config
		if len(config.Server.AllowedOrigins) > 0 {
			allowedOrigins = append(allowedOrigins, config.Server.AllowedOrigins...)
		}
		corsConfig = middleware.ProductionCORSConfig(allowedOrigins)
	} else {
		// In development, use more permissive CORS settings
		corsConfig = middleware.DefaultCORSConfig()
	}

	// Security headers configuration
	securityConfig := middleware.DefaultSecurityConfig()
	if config.Environment == "production" {
		// In production, use stricter CSP policy
		securityConfig.ContentSecurityPolicy = "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; font-src 'self'; object-src 'none'; media-src 'self'; frame-src 'none';"
	}

	// Add global middleware for all requests (order matters)
	router.Use(middleware.RequestID())
	router.Use(middleware.ErrorHandler(logger))
	router.Use(middleware.CORSMiddleware(corsConfig, logger))
	router.Use(middleware.Security(securityConfig, logger))
	router.Use(middleware.Validation(logger))
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.RateLimiter(logger, middleware.RateLimiterConfig{
		RequestsPerSecond: 10,
		Burst:             20,
		CleanupInterval:   5 * time.Minute,
		ClientTimeout:     10 * time.Minute,
	}))

	// Add CSRF protection for mutating endpoints
	router.Use(middleware.ProtectAgainstCSRF(logger))

	// Register Swagger documentation endpoints
	docs.RegisterSwaggerEndpoints(router)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": config.Version,
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
			auth.POST("/register", userHandler.CreateUser)
		}

		// Public dictionary routes
		entries := v1.Group("/entries")
		{
			entries.GET("", entryHandler.ListEntries)
			entries.GET("/:id", entryHandler.GetEntry)
			entries.GET("/:id/meanings", entryHandler.ListMeanings)

			// Meanings
			meanings := entries.Group("/:entryId/meanings")
			{
				meanings.GET("/:meaningId", entryHandler.GetMeaning)
				meanings.GET("/:meaningId/translations", translationHandler.ListTranslations)
			}
		}

		// Protected routes - require authentication
		protected := v1.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			// User management
			users := protected.Group("/users")
			{
				users.GET("/me", userHandler.GetCurrentUser)
				users.PUT("/me", userHandler.UpdateCurrentUser)
				users.DELETE("/me", userHandler.DeleteCurrentUser)

				// User content
				users.GET("/me/entries", userHandler.ListUserEntries)
				users.GET("/me/translations", userHandler.ListUserTranslations)
				users.GET("/me/comments", userHandler.ListUserComments)
				users.GET("/me/likes", userHandler.ListUserLikes)
			}

			// Entry management
			protectedEntries := protected.Group("/entries")
			{
				protectedEntries.POST("", entryHandler.CreateEntry)
				protectedEntries.PUT("/:id", entryHandler.UpdateEntry)
				protectedEntries.DELETE("/:id", entryHandler.DeleteEntry)

				// Meaning management
				protectedMeanings := protectedEntries.Group("/:entryId/meanings")
				{
					protectedMeanings.POST("", entryHandler.AddMeaning)
					protectedMeanings.PUT("/:meaningId", entryHandler.UpdateMeaning)
					protectedMeanings.DELETE("/:meaningId", entryHandler.DeleteMeaning)
					protectedMeanings.POST("/:meaningId/comments", entryHandler.AddMeaningComment)
					protectedMeanings.POST("/:meaningId/likes", entryHandler.ToggleMeaningLike)

					// Translation management
					protectedTranslations := protectedMeanings.Group("/:meaningId/translations")
					{
						protectedTranslations.POST("", translationHandler.CreateTranslation)
						protectedTranslations.PUT("/:translationId", translationHandler.UpdateTranslation)
						protectedTranslations.DELETE("/:translationId", translationHandler.DeleteTranslation)
						protectedTranslations.POST("/:translationId/comments", translationHandler.AddTranslationComment)
						protectedTranslations.POST("/:translationId/likes", translationHandler.ToggleTranslationLike)
					}
				}
			}
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(authMiddleware.RequireAdmin())
		{
			// Admin routes go here
			admin.GET("/stats", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"status":  "ok",
					"message": "Admin stats endpoint",
				})
			})
		}
	}

	return &ginRouter{
		engine: router,
		config: config,
		logger: logger,
	}
}
