package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/valpere/trytrago/interface/api/rest/handler"
	"github.com/valpere/trytrago/test/mocks"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize JWT for auth tests
	auth.InitJWT("test-secret-key", 1*time.Hour)
}

// setupRouter function with completely separate endpoints
func setupRouter(entryHandler handler.EntryHandlerInterface,
	translationHandler handler.TranslationHandlerInterface,
	userHandler handler.UserHandlerInterface) *gin.Engine {

	router := gin.New()

	// For tests, we'll use completely separate endpoints
	// to avoid any parameter conflicts

	// Public routes
	if entryHandler != nil {
		router.GET("/api/v1/entries", entryHandler.ListEntries)
		router.GET("/api/v1/entries/:id", entryHandler.GetEntry)
	}

	// Auth routes
	if userHandler != nil {
		router.POST("/api/v1/auth/login", userHandler.Login)
		router.POST("/api/v1/auth/register", userHandler.CreateUser)
		router.POST("/api/v1/auth/refresh", userHandler.RefreshToken)
	}

	// Mock auth middleware for protected routes
	authMiddleware := func(c *gin.Context) {
		userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
		c.Set("userID", userID)
		c.Set("username", "testuser")
		c.Set("userRole", "USER")
		c.Set("authenticated", true)
		c.Next()
	}

	// Protected routes - each with a unique path
	if entryHandler != nil {
		// Basic entry endpoints
		router.POST("/api/v1/entries", authMiddleware, entryHandler.CreateEntry)
		router.PUT("/api/v1/entries/:id", authMiddleware, entryHandler.UpdateEntry)
		router.DELETE("/api/v1/entries/:id", authMiddleware, entryHandler.DeleteEntry)

		// Meaning endpoints with direct paths
		router.GET("/api/v1/entry-meanings/:entry_id", authMiddleware, entryHandler.ListMeanings)
		router.POST("/api/v1/entry-meanings/:entry_id", authMiddleware, entryHandler.AddMeaning)
		router.GET("/api/v1/entry-meanings/:entry_id/:meaning_id", authMiddleware, entryHandler.GetMeaning)
		router.PUT("/api/v1/entry-meanings/:entry_id/:meaning_id", authMiddleware, entryHandler.UpdateMeaning)
		router.DELETE("/api/v1/entry-meanings/:entry_id/:meaning_id", authMiddleware, entryHandler.DeleteMeaning)
	}

	// Translation endpoints with their own paths
	if translationHandler != nil {
		router.GET("/api/v1/translations/:meaning_id", authMiddleware, translationHandler.ListTranslations)
		router.POST("/api/v1/translations/:meaning_id", authMiddleware, translationHandler.CreateTranslation)
		router.PUT("/api/v1/translations/:meaning_id/:translation_id", authMiddleware, translationHandler.UpdateTranslation)
		router.DELETE("/api/v1/translations/:meaning_id/:translation_id", authMiddleware, translationHandler.DeleteTranslation)
	}

	// User routes
	if userHandler != nil {
		router.GET("/api/v1/users/me", authMiddleware, userHandler.GetCurrentUser)
		router.PUT("/api/v1/users/me", authMiddleware, userHandler.UpdateCurrentUser)
		router.DELETE("/api/v1/users/me", authMiddleware, userHandler.DeleteCurrentUser)
	}

	return router
}

// setupMockEntryService creates a mock entry service
func setupMockEntryService() *MockEntryService {
	return new(MockEntryService)
}

// setupMockTranslationService creates a mock translation service
func setupMockTranslationService() *MockTranslationService {
	return new(MockTranslationService)
}

// setupMockUserService creates a mock user service
func setupMockUserService() *MockUserService {
	return new(MockUserService)
}

// TestListEntries tests the ListEntries endpoint
func TestListEntries(t *testing.T) {
	// Create mock service
	mockEntryService := setupMockEntryService()

	// Setup test data
	mockEntries := &response.EntryListResponse{
		Entries: []*response.EntryResponse{
			{
				ID:   uuid.New(),
				Word: "test1",
				Type: "WORD",
			},
			{
				ID:   uuid.New(),
				Word: "test2",
				Type: "PHRASE",
			},
		},
		Total:  2,
		Limit:  10,
		Offset: 0,
	}

	// Setup expectations
	mockEntryService.On("ListEntries", mock.Anything, mock.Anything).Return(mockEntries, nil)

	// Create handlers with mock services
	entryHandler := handler.NewEntryHandler(mockEntryService, mocks.SetupLoggerMock())

	// Create empty mock handlers for other interfaces to avoid nil pointers
	mockTransHandler := new(mocks.MockTranslationHandler)
	mockUserHandler := new(mocks.MockUserHandler)

	// Setup router with handlers
	router := setupRouter(entryHandler, mockTransHandler, mockUserHandler)

	// Create request
	req, err := http.NewRequest("GET", "/api/v1/entries", nil)
	require.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp response.EntryListResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Check response data
	assert.Equal(t, len(mockEntries.Entries), len(resp.Entries))
	assert.Equal(t, mockEntries.Total, resp.Total)

	// Verify mock expectations
	mockEntryService.AssertExpectations(t)
}

func TestGetEntry(t *testing.T) {
	// Create mock service
	mockEntryService := setupMockEntryService()

	// Setup test data
	entryID := uuid.New()
	mockEntry := &response.EntryResponse{
		ID:            entryID,
		Word:          "test",
		Type:          "WORD",
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Setup expectations
	mockEntryService.On("GetEntryByID", mock.Anything, entryID).Return(mockEntry, nil)

	// Create handlers with mock services
	entryHandler := handler.NewEntryHandler(mockEntryService, mocks.SetupLoggerMock())

	// Create empty mock handlers for other interfaces to avoid nil pointers
	mockTransHandler := new(mocks.MockTranslationHandler)
	mockUserHandler := new(mocks.MockUserHandler)

	// Setup router with handlers
	router := setupRouter(entryHandler, mockTransHandler, mockUserHandler)

	// Create request
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/entries/%s", entryID), nil)
	require.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp response.EntryResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Check response data
	assert.Equal(t, mockEntry.ID, resp.ID)
	assert.Equal(t, mockEntry.Word, resp.Word)
	assert.Equal(t, mockEntry.Type, resp.Type)

	// Verify mock expectations
	mockEntryService.AssertExpectations(t)
}

func TestCreateEntry(t *testing.T) {
	// Create mock service
	mockEntryService := setupMockEntryService()

	// Setup test data
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

	// Setup expectations
	mockEntryService.On("CreateEntry", mock.Anything, mock.MatchedBy(func(req *request.CreateEntryRequest) bool {
		return req.Word == createReq.Word && req.Type == createReq.Type
	})).Return(mockEntry, nil)

	// Create handlers with mock services
	entryHandler := handler.NewEntryHandler(mockEntryService, mocks.SetupLoggerMock())

	// Create empty mock handlers for other interfaces to avoid nil pointers
	mockTransHandler := new(mocks.MockTranslationHandler)
	mockUserHandler := new(mocks.MockUserHandler)

	// Setup router with handlers
	router := setupRouter(entryHandler, mockTransHandler, mockUserHandler)

	// Create request body
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	// Create request - use the updated path
	req, err := http.NewRequest("POST", "/api/v1/entries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Check response
	require.Equal(t, http.StatusCreated, w.Code)

	// Parse response
	var resp response.EntryResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Check response data
	assert.Equal(t, mockEntry.ID, resp.ID)
	assert.Equal(t, mockEntry.Word, resp.Word)
	assert.Equal(t, mockEntry.Type, resp.Type)

	// Verify mock expectations
	mockEntryService.AssertExpectations(t)
}

func TestRegisterUser(t *testing.T) {
	// Create mock service
	mockUserService := setupMockUserService()

	// Setup test data
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

	// Setup expectations
	mockUserService.On("CreateUser", mock.Anything, mock.MatchedBy(func(req *request.CreateUserRequest) bool {
		return req.Username == createReq.Username && req.Email == createReq.Email
	})).Return(mockUser, nil)

	// Create handlers with mock services
	userHandler := handler.NewUserHandler(mockUserService, mocks.SetupLoggerMock())

	// Create empty mock handlers for other interfaces to avoid nil pointers
	mockEntryHandler := new(mocks.MockEntryHandler)
	mockTransHandler := new(mocks.MockTranslationHandler)

	// Setup router with handlers
	router := setupRouter(mockEntryHandler, mockTransHandler, userHandler)

	// Create request body
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	// Create request
	req, err := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Check response
	require.Equal(t, http.StatusCreated, w.Code)

	// Parse response
	var resp response.UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Check response data
	assert.Equal(t, mockUser.Username, resp.Username)
	assert.Equal(t, mockUser.Email, resp.Email)

	// Verify mock expectations
	mockUserService.AssertExpectations(t)
}

func TestLogin(t *testing.T) {
	// Create mock service
	mockUserService := setupMockUserService()

	// Setup test data
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

	// Setup expectations
	mockUserService.On("Authenticate", mock.Anything, mock.MatchedBy(func(req *request.AuthRequest) bool {
		return req.Username == loginReq.Username && req.Password == loginReq.Password
	})).Return(mockAuthResp, nil)

	// Create handlers with mock services
	userHandler := handler.NewUserHandler(mockUserService, mocks.SetupLoggerMock())

	// Create empty mock handlers for other interfaces to avoid nil pointers
	mockEntryHandler := new(mocks.MockEntryHandler)
	mockTransHandler := new(mocks.MockTranslationHandler)

	// Setup router with handlers
	router := setupRouter(mockEntryHandler, mockTransHandler, userHandler)

	// Create request body
	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	// Create request
	req, err := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp response.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Check response data
	assert.Equal(t, mockAuthResp.AccessToken, resp.AccessToken)
	assert.Equal(t, mockAuthResp.RefreshToken, resp.RefreshToken)
	assert.Equal(t, mockAuthResp.User.Username, resp.User.Username)

	// Verify mock expectations
	mockUserService.AssertExpectations(t)
}

func TestRefreshToken(t *testing.T) {
	// Create mock service
	mockUserService := setupMockUserService()

	// Setup test data
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

	// Setup expectations
	mockUserService.On("RefreshToken", mock.Anything, refreshReq.RefreshToken).Return(mockAuthResp, nil)

	// Create handlers with mock services
	userHandler := handler.NewUserHandler(mockUserService, mocks.SetupLoggerMock())

	// Create empty mock handlers for other interfaces to avoid nil pointers
	mockEntryHandler := new(mocks.MockEntryHandler)
	mockTransHandler := new(mocks.MockTranslationHandler)

	// Setup router with handlers
	router := setupRouter(mockEntryHandler, mockTransHandler, userHandler)

	// Create request body
	reqBody, err := json.Marshal(refreshReq)
	require.NoError(t, err)

	// Create request
	req, err := http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp response.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Check response data
	assert.Equal(t, mockAuthResp.AccessToken, resp.AccessToken)
	assert.Equal(t, mockAuthResp.RefreshToken, resp.RefreshToken)

	// Verify mock expectations
	mockUserService.AssertExpectations(t)
}

func TestCurrentUser(t *testing.T) {
	// Create mock service
	mockUserService := setupMockUserService()

	// Setup test data
	userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001") // Match the ID set in auth middleware mock
	mockUser := &response.UserResponse{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Setup expectations
	mockUserService.On("GetUser", mock.Anything, userID).Return(mockUser, nil)

	// Create handlers with mock services
	userHandler := handler.NewUserHandler(mockUserService, mocks.SetupLoggerMock())

	// Create empty mock handlers for other interfaces to avoid nil pointers
	mockEntryHandler := new(mocks.MockEntryHandler)
	mockTransHandler := new(mocks.MockTranslationHandler)

	// Setup router with handlers
	router := setupRouter(mockEntryHandler, mockTransHandler, userHandler)

	// Create request
	req, err := http.NewRequest("GET", "/api/v1/users/me", nil)
	require.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp response.UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Check response data
	assert.Equal(t, mockUser.ID, resp.ID)
	assert.Equal(t, mockUser.Username, resp.Username)
	assert.Equal(t, mockUser.Email, resp.Email)

	// Verify mock expectations
	mockUserService.AssertExpectations(t)
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
