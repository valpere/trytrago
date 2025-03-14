package rest

import (
    "github.com/gin-gonic/gin"
    "github.com/valpere/trytrago/domain"
    "github.com/valpere/trytrago/domain/logging"
    "github.com/valpere/trytrago/interface/api/rest/handler"
    "github.com/valpere/trytrago/interface/api/rest/middleware"
)

// Router defines the interface for the REST router
type Router interface {
    // Handler returns the HTTP handler
    Handler() *gin.Engine

    // Config returns the current configuration
    Config() domain.Config
}

// ginRouter implements the Router interface using Gin
type ginRouter struct {
    engine  *gin.Engine
    config  domain.Config
    logger  logging.Logger
}

// NewRouter creates a new Router instance
func NewRouter(
    config domain.Config,
    logger logging.Logger,
    entryHandler *handler.EntryHandler,
    translationHandler *handler.TranslationHandler,
    userHandler *handler.UserHandler,
    authMiddleware middleware.AuthMiddleware,
) Router {
    // Set Gin mode based on environment
    if config.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    // Create router with middleware
    router := gin.New()

    // Add middleware
    router.Use(middleware.Logger(logger))
    router.Use(middleware.Recovery(logger))
    router.Use(middleware.RateLimiter(logger, middleware.RateLimiterConfig{
        RequestsPerSecond: 10,
        Burst:             20,
    }))

    // Health check endpoint
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
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
                    "status": "ok",
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

// Handler returns the HTTP handler
func (r *ginRouter) Handler() *gin.Engine {
    return r.engine
}

// Config returns the current configuration
func (r *ginRouter) Config() domain.Config {
    return r.config
}
