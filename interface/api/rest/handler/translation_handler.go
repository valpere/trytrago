package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/logging"
)

// TranslationHandler handles HTTP requests related to translations
type TranslationHandler struct {
	service service.TranslationService
	logger  logging.Logger
}

// NewTranslationHandler creates a new instance of TranslationHandler
func NewTranslationHandler(service service.TranslationService, logger logging.Logger) *TranslationHandler {
	return &TranslationHandler{
		service: service,
		logger:  logger.With(logging.String("component", "translation_handler")),
	}
}

// ListTranslations handles GET /api/v1/entries/:entryId/meanings/:meaningId/translations
func (h *TranslationHandler) ListTranslations(c *gin.Context) {
	meaningIDParam := c.Param("meaningId")
	languageID := c.Query("language")
	
	// Parse meaning UUID
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meaning ID format"})
		return
	}
	
	// Call service
	resp, err := h.service.ListTranslations(c.Request.Context(), meaningID, languageID)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Meaning not found"})
			return
		}
		
		h.logger.Error("failed to list translations", 
			logging.Error(err), 
			logging.String("meaningId", meaningIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve translations"})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

// CreateTranslation handles POST /api/v1/entries/:entryId/meanings/:meaningId/translations
func (h *TranslationHandler) CreateTranslation(c *gin.Context) {
	meaningIDParam := c.Param("meaningId")
	
	// Parse meaning UUID
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meaning ID format"})
		return
	}
	
	var req request.CreateTranslationRequest
	
	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid create translation request", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	// Call service
	resp, err := h.service.CreateTranslation(c.Request.Context(), meaningID, &req)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Meaning not found"})
			return
		}
		
		h.logger.Error("failed to create translation", 
			logging.Error(err), 
			logging.String("meaningId", meaningIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create translation"})
		return
	}
	
	c.JSON(http.StatusCreated, resp)
}

// UpdateTranslation handles PUT /api/v1/entries/:entryId/meanings/:meaningId/translations/:translationId
func (h *TranslationHandler) UpdateTranslation(c *gin.Context) {
	translationIDParam := c.Param("translationId")
	
	// Parse translation UUID
	translationID, err := uuid.Parse(translationIDParam)
	if err != nil {
		h.logger.Warn("invalid translation ID format", logging.String("id", translationIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid translation ID format"})
		return
	}
	
	var req request.UpdateTranslationRequest
	
	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid update translation request", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	// Call service
	resp, err := h.service.UpdateTranslation(c.Request.Context(), translationID, &req)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Translation not found"})
			return
		}
		
		h.logger.Error("failed to update translation", 
			logging.Error(err), 
			logging.String("translationId", translationIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update translation"})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

// DeleteTranslation handles DELETE /api/v1/entries/:entryId/meanings/:meaningId/translations/:translationId
func (h *TranslationHandler) DeleteTranslation(c *gin.Context) {
	translationIDParam := c.Param("translationId")
	
	// Parse translation UUID
	translationID, err := uuid.Parse(translationIDParam)
	if err != nil {
		h.logger.Warn("invalid translation ID format", logging.String("id", translationIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid translation ID format"})
		return
	}
	
	// Call service
	err = h.service.DeleteTranslation(c.Request.Context(), translationID)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Translation not found"})
			return
		}
		
		h.logger.Error("failed to delete translation", 
			logging.Error(err), 
			logging.String("translationId", translationIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete translation"})
		return
	}
	
	c.Status(http.StatusNoContent)
}

// AddTranslationComment handles POST /api/v1/entries/:entryId/meanings/:meaningId/translations/:translationId/comments
func (h *TranslationHandler) AddTranslationComment(c *gin.Context) {
	translationIDParam := c.Param("translationId")
	
	// Parse translation UUID
	translationID, err := uuid.Parse(translationIDParam)
	if err != nil {
		h.logger.Warn("invalid translation ID format", logging.String("id", translationIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid translation ID format"})
		return
	}
	
	var req request.CreateCommentRequest
	
	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid create comment request", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	// Call service
	resp, err := h.service.AddTranslationComment(c.Request.Context(), translationID, &req)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Translation not found"})
			return
		}
		
		h.logger.Error("failed to add comment to translation", 
			logging.Error(err), 
			logging.String("translationId", translationIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add comment"})
		return
	}
	
	c.JSON(http.StatusCreated, resp)
}

// ToggleTranslationLike handles POST /api/v1/entries/:entryId/meanings/:meaningId/translations/:translationId/likes
func (h *TranslationHandler) ToggleTranslationLike(c *gin.Context) {
	translationIDParam := c.Param("translationId")
	
	// Parse translation UUID
	translationID, err := uuid.Parse(translationIDParam)
	if err != nil {
		h.logger.Warn("invalid translation ID format", logging.String("id", translationIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid translation ID format"})
		return
	}
	
	// Get user ID from the context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("user ID not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		return
	}
	
	// Call service
	err = h.service.ToggleTranslationLike(c.Request.Context(), translationID, userID.(uuid.UUID))
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Translation not found"})
			return
		}
		
		h.logger.Error("failed to toggle like on translation", 
			logging.Error(err), 
			logging.String("translationId", translationIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle like"})
		return
	}
	
	c.Status(http.StatusNoContent)
}
