package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/interface/api/rest/handler"
	"github.com/valpere/trytrago/interface/api/rest/middleware"
)

// SetupRouter configures the Gin router with all routes and middleware
func SetupRouter(
	logger logging.Logger,
	entryService service.EntryService,
	translationService service.TranslationService,
	userService service.UserService,
) *gin.Engine {
	// Create router with default middleware
	router := gin.New()
	
	// Add custom middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(gin.Recovery())
	
	// Create handlers
	entryHandler := handler.NewEntryHandler(entryService, logger)
	translationHandler := handler.NewTranslationHandler(translationService, logger)
	userHandler := handler.NewUserHandler(userService, logger)
	
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
		}
		
		// Dictionary read-only operations are public
		entries := v1.Group("/entries")
		{
			entries.GET("", entryHandler.ListEntries)
			entries.GET("/:id", entryHandler.GetEntry)
			entries.GET("/:id/meanings", entryHandler.ListMeanings)
			
			meanings := entries.Group("/:entryId/meanings")
			{
				meanings.GET("/:meaningId", entryHandler.GetMeaning)
				meanings.GET("/:meaningId/translations", translationHandler.ListTranslations)
			}
		}
		
		// User routes
		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
		}
		
		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.Auth())
		{
			// User management
			protectedUsers := protected.Group("/users")
			{
				protectedUsers.GET("/me", userHandler.GetCurrentUser)
				protectedUsers.PUT("/me", userHandler.UpdateCurrentUser)
				protectedUsers.DELETE("/me", userHandler.DeleteCurrentUser)
				
				// User content listings
				protectedUsers.GET("/me/entries", userHandler.ListUserEntries)
				protectedUsers.GET("/me/translations", userHandler.ListUserTranslations)
				protectedUsers.GET("/me/comments", userHandler.ListUserComments)
				protectedUsers.GET("/me/likes", userHandler.ListUserLikes)
			}
			
			// Entry management (protected)
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
	}
	
	return router
}
