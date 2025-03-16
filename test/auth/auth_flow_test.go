package auth_test

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
	"github.com/stretchr/testify/suite"

	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/infrastructure/auth"
	"github.com/valpere/trytrago/interface/api/rest/handler"
	"github.com/valpere/trytrago/interface/api/rest/middleware"
	"github.com/valpere/trytrago/test/mocks"
)

// AuthFlowTestSuite contains tests for the authentication flow
type AuthFlowTestSuite struct {
	suite.Suite
	engine         *gin.Engine
	userService    *MockUserService
	logger         *mocks.MockLogger
	authMiddleware middleware.AuthMiddleware
	userHandler    *handler.UserHandler
	testUserID     uuid.UUID
	accessToken    string
	refreshToken   string
}

// MockUserService for testing authentication
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

// Implementing other required methods
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

func (s *AuthFlowTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize JWT
	auth.InitJWT("test-secret-key-for-auth-flow", 1*time.Hour)

	// Create a test user ID
	s.testUserID = uuid.New()

	// Generate real JWT tokens for testing
	var err error
	s.accessToken, err = auth.GenerateToken(s.testUserID.String(), "testuser", "USER")
	require.NoError(s.T(), err, "Failed to generate access token")

	s.refreshToken, err = auth.GenerateRefreshToken(s.testUserID.String())
	require.NoError(s.T(), err, "Failed to generate refresh token")
}

func (s *AuthFlowTestSuite) SetupTest() {
	// Create logger
	s.logger = mocks.SetupLoggerMock()

	// Create mock services
	s.userService = new(MockUserService)

	// Create auth middleware
	s.authMiddleware = middleware.NewAuthMiddleware(s.logger)

	// Create handlers
	s.userHandler = handler.NewUserHandler(s.userService, s.logger)

	// Setup router
	s.engine = s.setupRouter()
}

// setupRouter configures a test router with authentication endpoints and protected routes
func (s *AuthFlowTestSuite) setupRouter() *gin.Engine {
	router := gin.New()

	// Public auth routes
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", s.userHandler.CreateUser)
		auth.POST("/login", s.userHandler.Login)
		auth.POST("/refresh", s.userHandler.RefreshToken)
	}

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(s.authMiddleware.RequireAuth())
	{
		protected.GET("/users/me", s.userHandler.GetCurrentUser)
	}

	// Admin routes
	admin := router.Group("/api/v1/admin")
	admin.Use(s.authMiddleware.RequireAdmin())
	{
		admin.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "admin access granted"})
		})
	}

	return router
}

// TestRegistration tests the user registration flow
func (s *AuthFlowTestSuite) TestRegistration() {
	// Create registration request
	registerReq := &request.CreateUserRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
	}

	// Mock user service response
	mockUser := &response.UserResponse{
		ID:        uuid.New(),
		Username:  registerReq.Username,
		Email:     registerReq.Email,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	s.userService.On("CreateUser", mock.Anything, mock.MatchedBy(func(req *request.CreateUserRequest) bool {
		return req.Username == registerReq.Username && req.Email == registerReq.Email
	})).Return(mockUser, nil).Once()

	// Create request body
	reqBody, err := json.Marshal(registerReq)
	require.NoError(s.T(), err)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	s.engine.ServeHTTP(w, req)

	// Check response
	assert.Equal(s.T(), http.StatusCreated, w.Code)

	// Parse response
	var resp response.UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(s.T(), err)

	// Verify response
	assert.Equal(s.T(), mockUser.Username, resp.Username)
	assert.Equal(s.T(), mockUser.Email, resp.Email)

	// Verify mock calls
	s.userService.AssertExpectations(s.T())
}

// TestLoginSuccess tests successful user login
func (s *AuthFlowTestSuite) TestLoginSuccess() {
	// Create login request
	loginReq := &request.AuthRequest{
		Username: "testuser",
		Password: "password123",
	}

	// Mock authentication response
	mockAuthResp := &response.AuthResponse{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    3600,
		User: response.UserResponse{
			ID:       s.testUserID,
			Username: loginReq.Username,
			Email:    "test@example.com",
		},
	}

	s.userService.On("Authenticate", mock.Anything, mock.MatchedBy(func(req *request.AuthRequest) bool {
		return req.Username == loginReq.Username && req.Password == loginReq.Password
	})).Return(mockAuthResp, nil).Once()

	// Create request body
	reqBody, err := json.Marshal(loginReq)
	require.NoError(s.T(), err)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	s.engine.ServeHTTP(w, req)

	// Check response
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// Parse response
	var resp response.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(s.T(), err)

	// Verify response
	assert.NotEmpty(s.T(), resp.AccessToken)
	assert.NotEmpty(s.T(), resp.RefreshToken)
	assert.Equal(s.T(), mockAuthResp.User.Username, resp.User.Username)

	// Verify mock calls
	s.userService.AssertExpectations(s.T())
}

// TestLoginFailure tests login with invalid credentials
func (s *AuthFlowTestSuite) TestLoginFailure() {
	// Create login request with invalid password
	loginReq := &request.AuthRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	// Mock authentication error
	authErr := fmt.Errorf("invalid credentials")

	s.userService.On("Authenticate", mock.Anything, mock.MatchedBy(func(req *request.AuthRequest) bool {
		return req.Username == loginReq.Username && req.Password == loginReq.Password
	})).Return(nil, authErr).Once()

	// Create request body
	reqBody, err := json.Marshal(loginReq)
	require.NoError(s.T(), err)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	s.engine.ServeHTTP(w, req)

	// Check response
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)

	// Verify mock calls
	s.userService.AssertExpectations(s.T())
}

// TestRefreshToken tests the token refresh flow
func (s *AuthFlowTestSuite) TestRefreshToken() {
	// Create refresh token request
	refreshReq := &request.RefreshTokenRequest{
		RefreshToken: s.refreshToken,
	}

	// Mock refresh token response
	mockAuthResp := &response.AuthResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresIn:    3600,
		User: response.UserResponse{
			ID:       s.testUserID,
			Username: "testuser",
			Email:    "test@example.com",
		},
	}

	s.userService.On("RefreshToken", mock.Anything, s.refreshToken).Return(mockAuthResp, nil).Once()

	// Create request body
	reqBody, err := json.Marshal(refreshReq)
	require.NoError(s.T(), err)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	s.engine.ServeHTTP(w, req)

	// Check response
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// Parse response
	var resp response.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(s.T(), err)

	// Verify response
	assert.NotEmpty(s.T(), resp.AccessToken)
	assert.NotEmpty(s.T(), resp.RefreshToken)

	// Verify mock calls
	s.userService.AssertExpectations(s.T())
}

// TestInvalidRefreshToken tests using an invalid refresh token
func (s *AuthFlowTestSuite) TestInvalidRefreshToken() {
	// Create refresh token request with invalid token
	refreshReq := &request.RefreshTokenRequest{
		RefreshToken: "invalid-token",
	}

	// Mock refresh token error
	refreshErr := fmt.Errorf("invalid refresh token")

	s.userService.On("RefreshToken", mock.Anything, "invalid-token").Return(nil, refreshErr).Once()

	// Create request body
	reqBody, err := json.Marshal(refreshReq)
	require.NoError(s.T(), err)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	s.engine.ServeHTTP(w, req)

	// Check response
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)

	// Verify mock calls
	s.userService.AssertExpectations(s.T())
}

// TestAccessProtectedEndpoint tests accessing a protected endpoint with a valid token
func (s *AuthFlowTestSuite) TestAccessProtectedEndpoint() {
	// Mock user service response
	mockUser := &response.UserResponse{
		ID:       s.testUserID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	// We expect the handler to extract the user ID from the JWT token
	// and then call GetUser with that ID
	s.userService.On("GetUser", mock.Anything, s.testUserID).Return(mockUser, nil).Once()

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	s.engine.ServeHTTP(w, req)

	// Check response
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// Parse response
	var resp response.UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(s.T(), err)

	// Verify response
	assert.Equal(s.T(), mockUser.Username, resp.Username)
	assert.Equal(s.T(), mockUser.Email, resp.Email)

	// Verify mock calls
	s.userService.AssertExpectations(s.T())
}

// TestAccessProtectedEndpointWithoutToken tests accessing a protected endpoint without a token
func (s *AuthFlowTestSuite) TestAccessProtectedEndpointWithoutToken() {
	// Create request with no token
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	s.engine.ServeHTTP(w, req)

	// Check response - should be unauthorized
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

// TestAccessProtectedEndpointWithInvalidToken tests accessing a protected endpoint with an invalid token
func (s *AuthFlowTestSuite) TestAccessProtectedEndpointWithInvalidToken() {
	// Create request with invalid token
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	s.engine.ServeHTTP(w, req)

	// Check response - should be unauthorized
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

// TestAccessAdminEndpointWithUserToken tests accessing an admin endpoint with a regular user token
func (s *AuthFlowTestSuite) TestAccessAdminEndpointWithUserToken() {
	// Create request with user token
	req := httptest.NewRequest("GET", "/api/v1/admin/status", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	s.engine.ServeHTTP(w, req)

	// Check response - should be forbidden
	assert.Equal(s.T(), http.StatusForbidden, w.Code)
}

// TestAccessAdminEndpointWithAdminToken tests accessing an admin endpoint with an admin token
func (s *AuthFlowTestSuite) TestAccessAdminEndpointWithAdminToken() {
	// Generate admin token
	adminToken, err := auth.GenerateToken(s.testUserID.String(), "adminuser", "ADMIN")
	require.NoError(s.T(), err)

	// Create request with admin token
	req := httptest.NewRequest("GET", "/api/v1/admin/status", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	s.engine.ServeHTTP(w, req)

	// Check response - should be OK
	assert.Equal(s.T(), http.StatusOK, w.Code)
}

// TestAuthFlowIntegration runs the AuthFlowTestSuite
func TestAuthFlowIntegration(t *testing.T) {
	suite.Run(t, new(AuthFlowTestSuite))
}
