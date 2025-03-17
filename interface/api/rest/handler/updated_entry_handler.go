package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain/errors"
	"github.com/valpere/trytrago/domain/logging"
	restResponse "github.com/valpere/trytrago/interface/api/rest/response"
)

// EntryHandlerImpl implements the EntryHandlerInterface with improved error handling
type EntryHandlerImpl struct {
	service service.EntryService
	logger  logging.Logger
}

// NewEntryHandlerWithErrorHandling creates a new instance of EntryHandlerImpl with improved error handling
func NewEntryHandlerWithErrorHandling(service service.EntryService, logger logging.Logger) EntryHandlerInterface {
	return &EntryHandlerImpl{
		service: service,
		logger:  logger.With(logging.String("component", "entry_handler")),
	}
}

// ListEntries handles GET /api/v1/entries
func (h *EntryHandlerImpl) ListEntries(c *gin.Context) {
	var req request.ListEntriesRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("invalid list entries request", logging.Error(err))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrBadRequest,
			http.StatusBadRequest,
			"bad_request",
			"Invalid request parameters",
			map[string]interface{}{"query_params": err.Error()},
		), h.logger)
		return
	}

	// Set default values if not provided
	if req.Limit == 0 {
		req.Limit = 20
	}

	// Call service
	resp, err := h.service.ListEntries(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to list entries", logging.Error(err))
		restResponse.RespondWithError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetEntry handles GET /api/v1/entries/:id
func (h *EntryHandlerImpl) GetEntry(c *gin.Context) {
	idParam := c.Param("id")

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", idParam))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_id_format",
			"Invalid entry ID format",
			map[string]interface{}{"id": idParam},
		), h.logger)
		return
	}

	// Call service
	resp, err := h.service.GetEntryByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get entry",
			logging.Error(err),
			logging.String("id", idParam),
		)
		restResponse.RespondWithError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateEntry handles POST /api/v1/entries
func (h *EntryHandlerImpl) CreateEntry(c *gin.Context) {
	var req request.CreateEntryRequest

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid create entry request", logging.Error(err))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_request",
			"Invalid request format",
			map[string]interface{}{"validation": err.Error()},
		), h.logger)
		return
	}

	// Call service
	resp, err := h.service.CreateEntry(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create entry", logging.Error(err))
		restResponse.RespondWithError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// UpdateEntry handles PUT /api/v1/entries/:id
func (h *EntryHandlerImpl) UpdateEntry(c *gin.Context) {
	idParam := c.Param("id")

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", idParam))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_id_format",
			"Invalid entry ID format",
			map[string]interface{}{"id": idParam},
		), h.logger)
		return
	}

	var req request.UpdateEntryRequest

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid update entry request", logging.Error(err))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_request",
			"Invalid request format",
			map[string]interface{}{"validation": err.Error()},
		), h.logger)
		return
	}

	// Call service
	resp, err := h.service.UpdateEntry(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update entry",
			logging.Error(err),
			logging.String("id", idParam),
		)
		restResponse.RespondWithError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteEntry handles DELETE /api/v1/entries/:id
func (h *EntryHandlerImpl) DeleteEntry(c *gin.Context) {
	idParam := c.Param("id")

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", idParam))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_id_format",
			"Invalid entry ID format",
			map[string]interface{}{"id": idParam},
		), h.logger)
		return
	}

	// Call service
	err = h.service.DeleteEntry(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to delete entry",
			logging.Error(err),
			logging.String("id", idParam),
		)
		restResponse.RespondWithError(c, err, h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListMeanings handles GET /api/v1/entries/:id/meanings
func (h *EntryHandlerImpl) ListMeanings(c *gin.Context) {
	idParam := c.Param("id")

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", idParam))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_id_format",
			"Invalid entry ID format",
			map[string]interface{}{"id": idParam},
		), h.logger)
		return
	}

	// Call service
	resp, err := h.service.ListMeanings(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to list meanings",
			logging.Error(err),
			logging.String("entryId", idParam),
		)
		restResponse.RespondWithError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMeaning retrieves a specific meaning
func (h *EntryHandlerImpl) GetMeaning(c *gin.Context) {
	// Parse entry ID
	entryIDParam := c.Param("entryId")
	entryID, err := uuid.Parse(entryIDParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", entryIDParam))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_id_format",
			"Invalid entry ID format",
			map[string]interface{}{"id": entryIDParam},
		), h.logger)
		return
	}

	// Parse meaning ID
	meaningIDParam := c.Param("meaningId")
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_id_format",
			"Invalid meaning ID format",
			map[string]interface{}{"id": meaningIDParam},
		), h.logger)
		return
	}

	// Get the entry with its meanings
	entryResp, err := h.service.GetEntryByID(c.Request.Context(), entryID)
	if err != nil {
		h.logger.Error("failed to get entry",
			logging.Error(err),
			logging.String("entryId", entryIDParam),
		)
		restResponse.RespondWithError(c, err, h.logger)
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
		restResponse.RespondWithError(c, errors.ErrMeaningNotFound, h.logger)
		return
	}

	c.JSON(http.StatusOK, meaningResp)
}

// AddMeaning adds a new meaning to an entry
func (h *EntryHandlerImpl) AddMeaning(c *gin.Context) {
	entryIDParam := c.Param("entryId")

	// Parse UUID
	entryID, err := uuid.Parse(entryIDParam)
	if err != nil {
		h.logger.Warn("invalid entry ID format", logging.String("id", entryIDParam))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_id_format",
			"Invalid entry ID format",
			map[string]interface{}{"id": entryIDParam},
		), h.logger)
		return
	}

	var req request.CreateMeaningRequest

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid create meaning request", logging.Error(err))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_request",
			"Invalid request format",
			map[string]interface{}{"validation": err.Error()},
		), h.logger)
		return
	}

	// Call service
	resp, err := h.service.AddMeaning(c.Request.Context(), entryID, &req)
	if err != nil {
		h.logger.Error("failed to add meaning",
			logging.Error(err),
			logging.String("entryId", entryIDParam),
		)
		restResponse.RespondWithError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// UpdateMeaning updates an existing meaning
func (h *EntryHandlerImpl) UpdateMeaning(c *gin.Context) {
	meaningIDParam := c.Param("meaningId")

	// Parse UUID
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_id_format",
			"Invalid meaning ID format",
			map[string]interface{}{"id": meaningIDParam},
		), h.logger)
		return
	}

	var req request.UpdateMeaningRequest

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid update meaning request", logging.Error(err))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_request",
			"Invalid request format",
			map[string]interface{}{"validation": err.Error()},
		), h.logger)
		return
	}

	// Call service
	resp, err := h.service.UpdateMeaning(c.Request.Context(), meaningID, &req)
	if err != nil {
		h.logger.Error("failed to update meaning",
			logging.Error(err),
			logging.String("meaningId", meaningIDParam),
		)
		restResponse.RespondWithError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteMeaning handles deleting a meaning
func (h *EntryHandlerImpl) DeleteMeaning(c *gin.Context) {
	meaningIDParam := c.Param("meaningId")

	// Parse UUID
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_id_format",
			"Invalid meaning ID format",
			map[string]interface{}{"id": meaningIDParam},
		), h.logger)
		return
	}

	// Call service
	err = h.service.DeleteMeaning(c.Request.Context(), meaningID)
	if err != nil {
		h.logger.Error("failed to delete meaning",
			logging.Error(err),
			logging.String("meaningId", meaningIDParam),
		)
		restResponse.RespondWithError(c, err, h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

// AddMeaningComment adds a comment to a meaning
func (h *EntryHandlerImpl) AddMeaningComment(c *gin.Context) {
	meaningIDParam := c.Param("meaningId")

	// Parse UUID
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_id_format",
			"Invalid meaning ID format",
			map[string]interface{}{"id": meaningIDParam},
		), h.logger)
		return
	}

	var req request.CreateCommentRequest

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid create comment request", logging.Error(err))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_request",
			"Invalid request format",
			map[string]interface{}{"validation": err.Error()},
		), h.logger)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("user ID not found in context")
		restResponse.RespondWithError(c, errors.New(
			errors.ErrUnauthorized,
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		), h.logger)
		return
	}

	// Add user ID to request
	req.UserID = userID.(uuid.UUID)

	// Call service
	resp, err := h.service.AddMeaningComment(c.Request.Context(), meaningID, &req)
	if err != nil {
		h.logger.Error("failed to add comment to meaning",
			logging.Error(err),
			logging.String("meaningId", meaningIDParam),
		)
		restResponse.RespondWithError(c, err, h.logger)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ToggleMeaningLike toggles a like on a meaning
func (h *EntryHandlerImpl) ToggleMeaningLike(c *gin.Context) {
	meaningIDParam := c.Param("meaningId")

	// Parse UUID
	meaningID, err := uuid.Parse(meaningIDParam)
	if err != nil {
		h.logger.Warn("invalid meaning ID format", logging.String("id", meaningIDParam))
		restResponse.RespondWithError(c, errors.NewWithDetails(
			errors.ErrInvalidInput,
			http.StatusBadRequest,
			"invalid_id_format",
			"Invalid meaning ID format",
			map[string]interface{}{"id": meaningIDParam},
		), h.logger)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Error("user ID not found in context")
		restResponse.RespondWithError(c, errors.New(
			errors.ErrUnauthorized,
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		), h.logger)
		return
	}

	// Call service
	err = h.service.ToggleMeaningLike(c.Request.Context(), meaningID, userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("failed to toggle like on meaning",
			logging.Error(err),
			logging.String("meaningId", meaningIDParam),
		)
		restResponse.RespondWithError(c, err, h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}
