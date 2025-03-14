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

// UserHandler implements the UserHandlerInterface
type UserHandler struct {
    service service.UserService
    logger  logging.Logger
}

// NewUserHandler creates a new instance of UserHandler
func NewUserHandler(service service.UserService, logger logging.Logger) *UserHandler {
    return &UserHandler{
        service: service,
        logger:  logger.With(logging.String("component", "user_handler")),
    }
}

// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req request.CreateUserRequest

    // Bind JSON body
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Warn("invalid create user request", logging.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    // Call service
    resp, err := h.service.CreateUser(c.Request.Context(), &req)
    if err != nil {
        if database.IsDuplicateError(err) {
            c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
            return
        }

        h.logger.Error("failed to create user", logging.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, resp)
}

// GetCurrentUser handles GET /api/v1/users/me
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
    // Get user ID from context (set by auth middleware)
    userID, exists := c.Get("userID")
    if !exists {
        h.logger.Error("user ID not found in context")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
        return
    }

    // Call service
    resp, err := h.service.GetUser(c.Request.Context(), userID.(uuid.UUID))
    if err != nil {
        if database.IsNotFoundError(err) {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }

        h.logger.Error("failed to get user",
            logging.Error(err),
            logging.String("userId", userID.(uuid.UUID).String()),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

// UpdateCurrentUser handles PUT /api/v1/users/me
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
    // Get user ID from context (set by auth middleware)
    userID, exists := c.Get("userID")
    if !exists {
        h.logger.Error("user ID not found in context")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
        return
    }

    var req request.UpdateUserRequest

    // Bind JSON body
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Warn("invalid update user request", logging.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    // Call service
    resp, err := h.service.UpdateUser(c.Request.Context(), userID.(uuid.UUID), &req)
    if err != nil {
        if database.IsNotFoundError(err) {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }

        if database.IsDuplicateError(err) {
            c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
            return
        }

        h.logger.Error("failed to update user",
            logging.Error(err),
            logging.String("userId", userID.(uuid.UUID).String()),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

// DeleteCurrentUser handles DELETE /api/v1/users/me
func (h *UserHandler) DeleteCurrentUser(c *gin.Context) {
    // Get user ID from context (set by auth middleware)
    userID, exists := c.Get("userID")
    if !exists {
        h.logger.Error("user ID not found in context")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
        return
    }

    // Call service
    err := h.service.DeleteUser(c.Request.Context(), userID.(uuid.UUID))
    if err != nil {
        if database.IsNotFoundError(err) {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }

        h.logger.Error("failed to delete user",
            logging.Error(err),
            logging.String("userId", userID.(uuid.UUID).String()),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }

    c.Status(http.StatusNoContent)
}

// Login handles POST /api/v1/auth/login
func (h *UserHandler) Login(c *gin.Context) {
    var req request.AuthRequest

    // Bind JSON body
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Warn("invalid login request", logging.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    // Call service
    resp, err := h.service.Authenticate(c.Request.Context(), &req)
    if err != nil {
        h.logger.Warn("authentication failed",
            logging.Error(err),
            logging.String("username", req.Username),
        )
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
    var req request.RefreshTokenRequest

    // Bind JSON body
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Warn("invalid refresh token request", logging.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    // Call service
    resp, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
    if err != nil {
        h.logger.Warn("token refresh failed", logging.Error(err))
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

// ListUserEntries handles GET /api/v1/users/me/entries
func (h *UserHandler) ListUserEntries(c *gin.Context) {
    // Get user ID from context (set by auth middleware)
    userID, exists := c.Get("userID")
    if !exists {
        h.logger.Error("user ID not found in context")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
        return
    }

    var req request.ListEntriesRequest

    // Bind query parameters
    if err := c.ShouldBindQuery(&req); err != nil {
        h.logger.Warn("invalid list entries request", logging.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
        return
    }

    // Set default values if not provided
    if req.Limit == 0 {
        req.Limit = 20
    }

    // Call service
    resp, err := h.service.ListUserEntries(c.Request.Context(), userID.(uuid.UUID), &req)
    if err != nil {
        h.logger.Error("failed to list user entries",
            logging.Error(err),
            logging.String("userId", userID.(uuid.UUID).String()),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve entries"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

// ListUserTranslations handles GET /api/v1/users/me/translations
func (h *UserHandler) ListUserTranslations(c *gin.Context) {
    // Get user ID from context (set by auth middleware)
    userID, exists := c.Get("userID")
    if !exists {
        h.logger.Error("user ID not found in context")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
        return
    }

    var req request.ListTranslationsRequest

    // Bind query parameters
    if err := c.ShouldBindQuery(&req); err != nil {
        h.logger.Warn("invalid list translations request", logging.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
        return
    }

    // Set default values if not provided
    if req.Limit == 0 {
        req.Limit = 20
    }

    // Call service
    resp, err := h.service.ListUserTranslations(c.Request.Context(), userID.(uuid.UUID), &req)
    if err != nil {
        h.logger.Error("failed to list user translations",
            logging.Error(err),
            logging.String("userId", userID.(uuid.UUID).String()),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve translations"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

// ListUserComments handles GET /api/v1/users/me/comments
func (h *UserHandler) ListUserComments(c *gin.Context) {
    // Get user ID from context (set by auth middleware)
    userID, exists := c.Get("userID")
    if !exists {
        h.logger.Error("user ID not found in context")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
        return
    }

    var req request.ListCommentsRequest

    // Bind query parameters
    if err := c.ShouldBindQuery(&req); err != nil {
        h.logger.Warn("invalid list comments request", logging.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
        return
    }

    // Set default values if not provided
    if req.Limit == 0 {
        req.Limit = 20
    }

    // Call service
    resp, err := h.service.ListUserComments(c.Request.Context(), userID.(uuid.UUID), &req)
    if err != nil {
        h.logger.Error("failed to list user comments",
            logging.Error(err),
            logging.String("userId", userID.(uuid.UUID).String()),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve comments"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

// ListUserLikes handles GET /api/v1/users/me/likes
func (h *UserHandler) ListUserLikes(c *gin.Context) {
    // Get user ID from context (set by auth middleware)
    userID, exists := c.Get("userID")
    if !exists {
        h.logger.Error("user ID not found in context")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
        return
    }

    var req request.ListLikesRequest

    // Bind query parameters
    if err := c.ShouldBindQuery(&req); err != nil {
        h.logger.Warn("invalid list likes request", logging.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
        return
    }

    // Set default values if not provided
    if req.Limit == 0 {
        req.Limit = 20
    }

    // Call service
    resp, err := h.service.ListUserLikes(c.Request.Context(), userID.(uuid.UUID), &req)
    if err != nil {
        h.logger.Error("failed to list user likes",
            logging.Error(err),
            logging.String("userId", userID.(uuid.UUID).String()),
        )
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve likes"})
        return
    }

    c.JSON(http.StatusOK, resp)
}
