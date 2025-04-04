package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/domain/utils"
	"github.com/valpere/trytrago/interface/api/rest/middleware"
)

// EntryHandler implements the EntryHandlerInterface
type EntryHandler struct {
	service service.EntryService
	logger  logging.Logger
}

// NewEntryHandler creates a new instance of EntryHandler
func NewEntryHandler(service service.EntryService, logger logging.Logger) *EntryHandler {
	return &EntryHandler{
		service: service,
		logger:  logger.With(logging.String("component", "entry_handler")),
	}
}

// ListEntries handles GET /api/v1/entries
func (h *EntryHandler) ListEntries(c *gin.Context) {
	var req request.ListEntriesRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("invalid list entries request", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Sanitize input parameters - strip HTML tags properly
	if req.WordFilter != "" {
		req.WordFilter = utils.SanitizeString(req.WordFilter)
	}

	// Rest of the implementation...
}

// GetEntry handles GET /api/v1/entries/:id
func (h *EntryHandler) GetEntry(c *gin.Context) {
	// Get sanitized parameter from middleware
	idParam := middleware.GetSanitizedParam(c, "id")

	// Validate UUID format
	if !utils.IsValidUUID(idParam) {
		h.logger.Warn("invalid entry ID format", logging.String("id", idParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry ID format"})
		return
	}

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", idParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry ID format"})
		return
	}

	// Call service
	resp, err := h.service.GetEntryByID(c.Request.Context(), id)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
			return
		}

		h.logger.Error("failed to get entry", logging.Error(err), logging.String("id", idParam))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve entry"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateEntry handles PUT /api/v1/entries/:id
func (h *EntryHandler) UpdateEntry(c *gin.Context) {
	idParam := c.Param("id")

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", idParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry ID format"})
		return
	}

	var req request.UpdateEntryRequest

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid update entry request", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Call service
	resp, err := h.service.UpdateEntry(c.Request.Context(), id, &req)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
			return
		}

		h.logger.Error("failed to update entry", logging.Error(err), logging.String("id", idParam))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update entry"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteEntry handles DELETE /api/v1/entries/:id
func (h *EntryHandler) DeleteEntry(c *gin.Context) {
	idParam := c.Param("id")

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", idParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry ID format"})
		return
	}

	// Call service
	err = h.service.DeleteEntry(c.Request.Context(), id)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
			return
		}

		h.logger.Error("failed to delete entry", logging.Error(err), logging.String("id", idParam))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete entry"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListMeanings handles GET /api/v1/entries/:id/meanings
func (h *EntryHandler) ListMeanings(c *gin.Context) {
	idParam := c.Param("id")

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", idParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry ID format"})
		return
	}

	// Call service
	resp, err := h.service.ListMeanings(c.Request.Context(), id)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
			return
		}

		h.logger.Error("failed to list meanings", logging.Error(err), logging.String("entryId", idParam))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve meanings"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMeaning retrieves a specific meaning
func (h *EntryHandler) GetMeaning(c *gin.Context) {
	meaningIDParam := c.Param("meaningId")

	// Parse UUID
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meaning ID format"})
		return
	}

	// This endpoint is not directly implemented in the service interface
	// We need to find the meaning in the context of its entry

	// Get the entry ID from the path
	entryIDParam := c.Param("entryId")
	entryID, err := uuid.Parse(entryIDParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", entryIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry ID format"})
		return
	}

	// Get the entry with its meanings
	entryResp, err := h.service.GetEntryByID(c.Request.Context(), entryID)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
			return
		}

		h.logger.Error("failed to get entry",
			logging.Error(err),
			logging.String("entryId", entryIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve entry"})
		return
	}

	// Find the meaning with the given ID
	var meaningResp *response.MeaningResponse
	for _, meaning := range entryResp.Meanings {
		if meaning.ID == meaningID {
			meaningResp = &meaning
			break
		}
	}

	if meaningResp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Meaning not found"})
		return
	}

	c.JSON(http.StatusOK, meaningResp)
}

// AddMeaning adds a new meaning to an entry
func (h *EntryHandler) AddMeaning(c *gin.Context) {
	entryIDParam := c.Param("entryId")

	// Parse UUID
	entryID, err := uuid.Parse(entryIDParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", entryIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry ID format"})
		return
	}

	var req request.CreateMeaningRequest

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid create meaning request", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Call service
	resp, err := h.service.AddMeaning(c.Request.Context(), entryID, &req)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
			return
		}

		h.logger.Error("failed to add meaning",
			logging.Error(err),
			logging.String("entryId", entryIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add meaning"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// UpdateMeaning updates an existing meaning
func (h *EntryHandler) UpdateMeaning(c *gin.Context) {
	meaningIDParam := c.Param("meaningId")

	// Parse UUID
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meaning ID format"})
		return
	}

	var req request.UpdateMeaningRequest

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid update meaning request", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Call service
	resp, err := h.service.UpdateMeaning(c.Request.Context(), meaningID, &req)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Meaning not found"})
			return
		}

		h.logger.Error("failed to update meaning",
			logging.Error(err),
			logging.String("meaningId", meaningIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update meaning"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteMeaning deletes a meaning
func (h *EntryHandler) DeleteMeaning(c *gin.Context) {
	meaningIDParam := c.Param("meaningId")

	// Parse UUID
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meaning ID format"})
		return
	}

	// Call service
	err = h.service.DeleteMeaning(c.Request.Context(), meaningID)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Meaning not found"})
			return
		}

		h.logger.Error("failed to delete meaning",
			logging.Error(err),
			logging.String("meaningId", meaningIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete meaning"})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddMeaningComment adds a comment to a meaning
func (h *EntryHandler) AddMeaningComment(c *gin.Context) {
	meaningIDParam := c.Param("meaningId")

	// Parse UUID
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meaning ID format"})
		return
	}

	var req request.CreateCommentRequest

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid create comment request", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("user ID not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		return
	}

	// Add user ID to request from the authenticated user
	req.UserID = userID.(uuid.UUID)

	// Call service
	resp, err := h.service.AddMeaningComment(c.Request.Context(), meaningID, &req)
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Meaning not found"})
			return
		}

		h.logger.Error("failed to add comment to meaning",
			logging.Error(err),
			logging.String("meaningId", meaningIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add comment"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ToggleMeaningLike toggles a like on a meaning
func (h *EntryHandler) ToggleMeaningLike(c *gin.Context) {
	meaningIDParam := c.Param("meaningId")

	// Parse UUID
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meaning ID format"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("user ID not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		return
	}

	// Call service
	err = h.service.ToggleMeaningLike(c.Request.Context(), meaningID, userID.(uuid.UUID))
	if err != nil {
		if database.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Meaning not found"})
			return
		}

		h.logger.Error("failed to toggle like on meaning",
			logging.Error(err),
			logging.String("meaningId", meaningIDParam),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle like"})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateEntry handles POST /api/v1/entries
func (h *EntryHandler) CreateEntry(c *gin.Context) {
	var req request.CreateEntryRequest

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid create entry request", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Sanitize and validate input
	req.Word = utils.SanitizeString(req.Word)
	req.Pronunciation = utils.SanitizeString(req.Pronunciation)

	// Validate entry type
	if !utils.IsValidEntryType(req.Type) {
		h.logger.Warn("invalid entry type", logging.String("type", req.Type))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry type. Must be one of: WORD, COMPOUND_WORD, PHRASE"})
		return
	}

	// Call service
	resp, err := h.service.CreateEntry(c.Request.Context(), &req)
	if err != nil {
		if database.IsDuplicateError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Entry already exists"})
			return
		}

		h.logger.Error("failed to create entry", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create entry"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}
