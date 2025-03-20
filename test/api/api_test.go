// test/api/api_test.go
package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/infrastructure/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
	auth.InitJWT("test-secret-key", 1*time.Hour)
}

// Mock service implementations
type MockEntryService struct {
	mock.Mock
}

func (m *MockEntryService) CreateEntry(ctx context.Context, req *request.CreateEntryRequest) (*response.EntryResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EntryResponse), args.Error(1)
}

func (m *MockEntryService) GetEntryByID(ctx context.Context, id uuid.UUID) (*response.EntryResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EntryResponse), args.Error(1)
}

func (m *MockEntryService) UpdateEntry(ctx context.Context, id uuid.UUID, req *request.UpdateEntryRequest) (*response.EntryResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EntryResponse), args.Error(1)
}

func (m *MockEntryService) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEntryService) ListEntries(ctx context.Context, req *request.ListEntriesRequest) (*response.EntryListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EntryListResponse), args.Error(1)
}

func (m *MockEntryService) AddMeaning(ctx context.Context, entryID uuid.UUID, req *request.CreateMeaningRequest) (*response.MeaningResponse, error) {
	args := m.Called(ctx, entryID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.MeaningResponse), args.Error(1)
}

func (m *MockEntryService) UpdateMeaning(ctx context.Context, id uuid.UUID, req *request.UpdateMeaningRequest) (*response.MeaningResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.MeaningResponse), args.Error(1)
}

func (m *MockEntryService) DeleteMeaning(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEntryService) ListMeanings(ctx context.Context, entryID uuid.UUID) (*response.MeaningListResponse, error) {
	args := m.Called(ctx, entryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.MeaningListResponse), args.Error(1)
}

func (m *MockEntryService) AddMeaningComment(ctx context.Context, meaningID uuid.UUID, req *request.CreateCommentRequest) (*response.CommentResponse, error) {
	args := m.Called(ctx, meaningID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CommentResponse), args.Error(1)
}

func (m *MockEntryService) ToggleMeaningLike(ctx context.Context, meaningID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, meaningID, userID)
	return args.Error(0)
}

type MockTranslationService struct {
	mock.Mock
}

func (m *MockTranslationService) CreateTranslation(ctx context.Context, meaningID uuid.UUID, req *request.CreateTranslationRequest) (*response.TranslationResponse, error) {
	args := m.Called(ctx, meaningID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.TranslationResponse), args.Error(1)
}

func (m *MockTranslationService) UpdateTranslation(ctx context.Context, id uuid.UUID, req *request.UpdateTranslationRequest) (*response.TranslationResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.TranslationResponse), args.Error(1)
}

func (m *MockTranslationService) DeleteTranslation(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTranslationService) ListTranslations(ctx context.Context, meaningID uuid.UUID, langID string) (*response.TranslationListResponse, error) {
	args := m.Called(ctx, meaningID, langID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.TranslationListResponse), args.Error(1)
}

func (m *MockTranslationService) AddTranslationComment(ctx context.Context, translationID uuid.UUID, req *request.CreateCommentRequest) (*response.CommentResponse, error) {
	args := m.Called(ctx, translationID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CommentResponse), args.Error(1)
}

func (m *MockTranslationService) ToggleTranslationLike(ctx context.Context, translationID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, translationID, userID)
	return args.Error(0)
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req *request.CreateUserRequest) (*response.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id uuid.UUID) (*response.UserResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uuid.UUID, req *request.UpdateUserRequest) (*response.UserResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) Authenticate(ctx context.Context, req *request.AuthRequest) (*response.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AuthResponse), args.Error(1)
}

func (m *MockUserService) RefreshToken(ctx context.Context, refreshToken string) (*response.AuthResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AuthResponse), args.Error(1)
}

func (m *MockUserService) ListUserEntries(ctx context.Context, userID uuid.UUID, req *request.ListEntriesRequest) (*response.EntryListResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EntryListResponse), args.Error(1)
}

func (m *MockUserService) ListUserTranslations(ctx context.Context, userID uuid.UUID, req *request.ListTranslationsRequest) (*response.TranslationListResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.TranslationListResponse), args.Error(1)
}

func (m *MockUserService) ListUserComments(ctx context.Context, userID uuid.UUID, req *request.ListCommentsRequest) (*response.CommentListResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CommentListResponse), args.Error(1)
}

func (m *MockUserService) ListUserLikes(ctx context.Context, userID uuid.UUID, req *request.ListLikesRequest) (*response.LikeListResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.LikeListResponse), args.Error(1)
}

// Helper functions
func setupMockEntryService() *MockEntryService {
	return new(MockEntryService)
}

func setupMockTranslationService() *MockTranslationService {
	return new(MockTranslationService)
}

func setupMockUserService() *MockUserService {
	return new(MockUserService)
}

// Test cases with inline handlers
func TestListEntries(t *testing.T) {
	mockEntryService := setupMockEntryService()

	mockEntries := &response.EntryListResponse{
		Entries: []*response.EntryResponse{
			{
				ID:            uuid.New(),
				Word:          "test1",
				Type:          "WORD",
				Pronunciation: "test1",
				CreatedAt:     time.Now().UTC(),
				UpdatedAt:     time.Now().UTC(),
			},
			{
				ID:            uuid.New(),
				Word:          "test2",
				Type:          "PHRASE",
				Pronunciation: "test2",
				CreatedAt:     time.Now().UTC(),
				UpdatedAt:     time.Now().UTC(),
			},
		},
		Total:  2,
		Limit:  10,
		Offset: 0,
	}

	mockEntryService.On("ListEntries", mock.Anything, mock.AnythingOfType("*request.ListEntriesRequest")).Return(mockEntries, nil)

	router := gin.New()
	router.GET("/api/v1/entries", func(c *gin.Context) {
		var req request.ListEntriesRequest
		if req.Limit == 0 {
			req.Limit = 20
		}

		resp, err := mockEntryService.ListEntries(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list entries"})
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	req := httptest.NewRequest("GET", "/api/v1/entries", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Body.String(), "Response body should not be empty")

	var resp response.EntryListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Should be valid JSON")

	assert.Equal(t, len(mockEntries.Entries), len(resp.Entries))
	assert.Equal(t, mockEntries.Total, resp.Total)

	mockEntryService.AssertExpectations(t)
}

func TestGetEntry(t *testing.T) {
	mockEntryService := setupMockEntryService()

	entryID := uuid.New()
	mockEntry := &response.EntryResponse{
		ID:            entryID,
		Word:          "test",
		Type:          "WORD",
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	mockEntryService.On("GetEntryByID", mock.Anything, entryID).Return(mockEntry, nil)

	router := gin.New()
	router.GET("/api/v1/entries/:id", func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entry ID format"})
			return
		}

		resp, err := mockEntryService.GetEntryByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get entry"})
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	req := httptest.NewRequest("GET", "/api/v1/entries/"+entryID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Body.String(), "Response body should not be empty")

	var resp response.EntryResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Should be valid JSON")

	assert.Equal(t, mockEntry.ID, resp.ID)
	assert.Equal(t, mockEntry.Word, resp.Word)
	assert.Equal(t, mockEntry.Type, resp.Type)

	mockEntryService.AssertExpectations(t)
}

func TestCreateEntry(t *testing.T) {
	mockEntryService := setupMockEntryService()

	createReq := &request.CreateEntryRequest{
		Word:          "test",
		Type:          "WORD",
		Pronunciation: "test",
	}

	mockEntry := &response.EntryResponse{
		ID:            uuid.New(),
		Word:          createReq.Word,
		Type:          createReq.Type,
		Pronunciation: createReq.Pronunciation,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	mockEntryService.On("CreateEntry", mock.Anything, mock.MatchedBy(func(req *request.CreateEntryRequest) bool {
		return req.Word == createReq.Word && req.Type == createReq.Type
	})).Return(mockEntry, nil)

	router := gin.New()
	router.POST("/api/v1/entries", func(c *gin.Context) {
		var req request.CreateEntryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		resp, err := mockEntryService.CreateEntry(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create entry"})
			return
		}

		c.JSON(http.StatusCreated, resp)
	})

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/entries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	require.Equal(t, http.StatusCreated, w.Code)
	require.NotEmpty(t, w.Body.String(), "Response body should not be empty")

	var resp response.EntryResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Should be valid JSON")

	assert.Equal(t, mockEntry.Word, resp.Word)
	assert.Equal(t, mockEntry.Type, resp.Type)

	mockEntryService.AssertExpectations(t)
}

func TestRegisterUser(t *testing.T) {
	mockUserService := setupMockUserService()

	createReq := &request.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockUser := &response.UserResponse{
		ID:        uuid.New(),
		Username:  createReq.Username,
		Email:     createReq.Email,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	mockUserService.On("CreateUser", mock.Anything, mock.MatchedBy(func(req *request.CreateUserRequest) bool {
		return req.Username == createReq.Username && req.Email == createReq.Email
	})).Return(mockUser, nil)

	router := gin.New()
	router.POST("/api/v1/auth/register", func(c *gin.Context) {
		var req request.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		resp, err := mockUserService.CreateUser(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, resp)
	})

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	require.Equal(t, http.StatusCreated, w.Code)
	require.NotEmpty(t, w.Body.String(), "Response body should not be empty")

	var resp response.UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Should be valid JSON")

	assert.Equal(t, mockUser.Username, resp.Username)
	assert.Equal(t, mockUser.Email, resp.Email)

	mockUserService.AssertExpectations(t)
}

func TestLogin(t *testing.T) {
	mockUserService := setupMockUserService()

	loginReq := &request.AuthRequest{
		Username: "testuser",
		Password: "password123",
	}

	userID := uuid.New()
	mockAuthResp := &response.AuthResponse{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresIn:    3600,
		User: response.UserResponse{
			ID:       userID,
			Username: loginReq.Username,
			Email:    "test@example.com",
		},
	}

	mockUserService.On("Authenticate", mock.Anything, mock.MatchedBy(func(req *request.AuthRequest) bool {
		return req.Username == loginReq.Username && req.Password == loginReq.Password
	})).Return(mockAuthResp, nil)

	router := gin.New()
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req request.AuthRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		resp, err := mockUserService.Authenticate(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Body.String(), "Response body should not be empty")

	var resp response.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Should be valid JSON")

	assert.Equal(t, mockAuthResp.AccessToken, resp.AccessToken)
	assert.Equal(t, mockAuthResp.RefreshToken, resp.RefreshToken)
	assert.Equal(t, mockAuthResp.User.Username, resp.User.Username)

	mockUserService.AssertExpectations(t)
}

func TestRefreshToken(t *testing.T) {
	mockUserService := setupMockUserService()

	refreshReq := &request.RefreshTokenRequest{
		RefreshToken: "mock_refresh_token",
	}

	userID := uuid.New()
	mockAuthResp := &response.AuthResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		ExpiresIn:    3600,
		User: response.UserResponse{
			ID:       userID,
			Username: "testuser",
			Email:    "test@example.com",
		},
	}

	mockUserService.On("RefreshToken", mock.Anything, refreshReq.RefreshToken).Return(mockAuthResp, nil)

	router := gin.New()
	router.POST("/api/v1/auth/refresh", func(c *gin.Context) {
		var req request.RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		resp, err := mockUserService.RefreshToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	reqBody, err := json.Marshal(refreshReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Body.String(), "Response body should not be empty")

	var resp response.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Should be valid JSON")

	assert.Equal(t, mockAuthResp.AccessToken, resp.AccessToken)
	assert.Equal(t, mockAuthResp.RefreshToken, resp.RefreshToken)

	mockUserService.AssertExpectations(t)
}

func TestCurrentUser(t *testing.T) {
	mockUserService := setupMockUserService()

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	mockUser := &response.UserResponse{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUserService.On("GetUser", mock.Anything, userID).Return(mockUser, nil)

	router := gin.New()
	router.GET("/api/v1/users/me", func(c *gin.Context) {
		// Simulate auth middleware
		c.Set("userID", userID)

		id, _ := c.Get("userID")
		resp, err := mockUserService.GetUser(c.Request.Context(), id.(uuid.UUID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Body.String(), "Response body should not be empty")

	var resp response.UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Should be valid JSON")

	assert.Equal(t, mockUser.ID, resp.ID)
	assert.Equal(t, mockUser.Username, resp.Username)
	assert.Equal(t, mockUser.Email, resp.Email)

	mockUserService.AssertExpectations(t)
}
