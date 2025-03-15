package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/model"
	"github.com/valpere/trytrago/infrastructure/auth"
	"github.com/valpere/trytrago/test/mocks"
)

func init() {
	// Initialize JWT auth for all tests
	auth.InitJWT("test-secret-key-for-unit-tests", 1*time.Hour)
}

// setupUserService creates a new instance of UserService with mocks
func setupUserService(t *testing.T) (service.UserService, *mocks.MockRepository, *mocks.MockLogger) {
	mockRepo := new(mocks.MockRepository)
	mockLogger := mocks.SetupLoggerMock()

	userService := service.NewUserService(mockRepo, mockLogger)

	return userService, mockRepo, mockLogger
}

// TestCreateUser tests the CreateUser function
func TestCreateUser(t *testing.T) {
	// Setup fixtures
	testUsername := "testuser"
	testEmail := "test@example.com"
	testPassword := "password123"

	// Create request
	createUserReq := &request.CreateUserRequest{
		Username: testUsername,
		Email:    testEmail,
		Password: testPassword,
	}

	// Test cases
	testCases := []struct {
		name          string
		setupMocks    func(*mocks.MockRepository)
		expectedError bool
		errorContains string
	}{
		{
			name: "Success",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				// In the actual implementation, CreateUser would be called with a valid user model
				mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *model.User) bool {
					return user.Username == testUsername &&
						user.Email == testEmail &&
						user.Password != testPassword // Password should be hashed
				})).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "RepositoryError",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				expectedError := errors.New("database error")
				mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(expectedError)
			},
			expectedError: true,
			errorContains: "failed to create user",
		},
		{
			name: "DuplicateUsername",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(database.ErrDuplicateEntry)
			},
			expectedError: true,
			errorContains: "duplicate entry",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup service and mocks
			userService, mockRepo, _ := setupUserService(t)

			// Setup specific test case expectations
			tc.setupMocks(mockRepo)

			// Call service
			resp, err := userService.CreateUser(context.Background(), createUserReq)

			// Assert expectations
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, testUsername, resp.Username)
				assert.Equal(t, testEmail, resp.Email)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestAuthenticate tests the Authenticate function
func TestAuthenticate(t *testing.T) {
	// Initialize JWT
	auth.InitJWT("test-secret-key", 1*time.Hour)

	// Create password and hash
	plainPassword := "testpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.MinCost)
	require.NoError(t, err)

	// Setup fixtures
	testID := uuid.New()
	testUsername := "testuser"
	testUser := &model.User{
		ID:        testID,
		Username:  testUsername,
		Password:  string(hashedPassword),
		Email:     "test@example.com",
		Role:      model.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Create request for successful auth
	validAuthReq := &request.AuthRequest{
		Username: testUsername,
		Password: plainPassword,
	}

	// Create request for invalid password
	invalidPasswordReq := &request.AuthRequest{
		Username: testUsername,
		Password: "wrongpassword",
	}

	// Test cases
	testCases := []struct {
		name          string
		request       *request.AuthRequest
		setupMocks    func(*mocks.MockRepository, *mocks.MockLogger)
		expectedError bool
		errorContains string
	}{
		{
			name:    "Success",
			request: validAuthReq,
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("GetUserByUsername", mock.Anything, testUsername).Return(testUser, nil).Once()
			},
			expectedError: false,
		},
		{
			name:    "UserNotFound",
			request: validAuthReq,
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("GetUserByUsername", mock.Anything, testUsername).Return(nil, database.ErrNotFound).Once()
				mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
			},
			expectedError: true,
			errorContains: "invalid credentials",
		},
		{
			name:    "InvalidPassword",
			request: invalidPasswordReq,
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("GetUserByUsername", mock.Anything, testUsername).Return(testUser, nil).Once()
				mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
			},
			expectedError: true,
			errorContains: "invalid credentials",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup service and mocks
			userService, mockRepo, mockLogger := setupUserService(t)

			// Setup specific test case expectations
			tc.setupMocks(mockRepo, mockLogger)

			// Call service
			resp, err := userService.Authenticate(context.Background(), tc.request)

			// Assert expectations
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
				assert.Equal(t, testUsername, resp.User.Username)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetUser tests the GetUser function
func TestGetUser(t *testing.T) {
	// Setup fixtures
	testID := uuid.New()
	testUser := &model.User{
		ID:        testID,
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      model.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Test cases
	testCases := []struct {
		name          string
		setupMocks    func(*mocks.MockRepository)
		expectedError bool
		errorContains string
	}{
		{
			name: "Success",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(testUser, nil).Once()
			},
			expectedError: false,
		},
		{
			name: "UserNotFound",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(nil, database.ErrNotFound).Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "DatabaseError",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				dbErr := errors.New("database error")
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(nil, dbErr).Once()
			},
			expectedError: true,
			errorContains: "failed to get user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup service and mocks
			userService, mockRepo, _ := setupUserService(t)

			// Setup specific test case expectations
			tc.setupMocks(mockRepo)

			// Call service
			resp, err := userService.GetUser(context.Background(), testID)

			// Assert expectations
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, testUser.Username, resp.Username)
				assert.Equal(t, testUser.Email, resp.Email)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUpdateUser tests the UpdateUser function
func TestUpdateUser(t *testing.T) {
	// Setup fixtures
	testID := uuid.New()
	originalUsername := "originaluser"
	originalEmail := "original@example.com"
	newUsername := "newuser"
	newEmail := "new@example.com"
	testUser := &model.User{
		ID:        testID,
		Username:  originalUsername,
		Email:     originalEmail,
		Password:  "hashedpassword",
		Role:      model.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now().UTC().Add(-24 * time.Hour),
		UpdatedAt: time.Now().UTC(),
	}

	// Create update request
	updateReq := &request.UpdateUserRequest{
		Username: newUsername,
		Email:    newEmail,
	}

	// Test cases
	testCases := []struct {
		name          string
		setupMocks    func(*mocks.MockRepository)
		expectedError bool
		errorContains string
	}{
		{
			name: "Success",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(testUser, nil).Once()
				mockRepo.On("UpdateUser", mock.Anything, mock.MatchedBy(func(user *model.User) bool {
					return user.ID == testID &&
						user.Username == newUsername &&
						user.Email == newEmail
				})).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name: "UserNotFound",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(nil, database.ErrNotFound).Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "DuplicateUsername",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(testUser, nil).Once()
				mockRepo.On("UpdateUser", mock.Anything, mock.Anything).Return(database.ErrDuplicateEntry).Once()
			},
			expectedError: true,
			errorContains: "duplicate entry",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup service and mocks
			userService, mockRepo, _ := setupUserService(t)

			// Setup specific test case expectations
			tc.setupMocks(mockRepo)

			// Call service
			resp, err := userService.UpdateUser(context.Background(), testID, updateReq)

			// Assert expectations
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, newUsername, resp.Username)
				assert.Equal(t, newEmail, resp.Email)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestDeleteUser tests the DeleteUser function
func TestDeleteUser(t *testing.T) {
	// Setup fixtures
	testID := uuid.New()

	// Test cases
	testCases := []struct {
		name          string
		setupMocks    func(*mocks.MockRepository)
		expectedError bool
		errorContains string
	}{
		{
			name: "Success",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.On("DeleteUser", mock.Anything, testID).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name: "UserNotFound",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.On("DeleteUser", mock.Anything, testID).Return(database.ErrNotFound).Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "DatabaseError",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				expectedError := errors.New("database error")
				mockRepo.On("DeleteUser", mock.Anything, testID).Return(expectedError).Once()
			},
			expectedError: true,
			errorContains: "failed to delete user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup service and mocks
			userService, mockRepo, _ := setupUserService(t)

			// Configure expectations
			tc.setupMocks(mockRepo)

			// Call service
			err := userService.DeleteUser(context.Background(), testID)

			// Assert expectations
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRefreshToken tests the RefreshToken function
func TestRefreshToken(t *testing.T) {
	// Setup fixtures
	testID := uuid.New()
	testUser := &model.User{
		ID:        testID,
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      model.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Generate a valid refresh token
	refreshToken, err := auth.GenerateRefreshToken(testID.String())
	require.NoError(t, err, "Failed to generate test refresh token")

	// Test cases
	testCases := []struct {
		name          string
		token         string
		setupMocks    func(*mocks.MockRepository)
		expectedError bool
		errorContains string
	}{
		{
			name:  "Success",
			token: refreshToken,
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(testUser, nil).Once()
			},
			expectedError: false,
		},
		{
			name:  "InvalidToken",
			token: "invalid-token",
			setupMocks: func(mockRepo *mocks.MockRepository) {
				// No mocks needed - validation fails before repo is accessed
			},
			expectedError: true,
			errorContains: "invalid refresh token",
		},
		{
			name:  "UserNotFound",
			token: refreshToken,
			setupMocks: func(mockRepo *mocks.MockRepository) {
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(nil, database.ErrNotFound).Once()
			},
			expectedError: true,
			errorContains: "invalid token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup service and mocks
			userService, mockRepo, _ := setupUserService(t)

			// Setup specific test case expectations
			tc.setupMocks(mockRepo)

			// Call service
			resp, err := userService.RefreshToken(context.Background(), tc.token)

			// Assert expectations
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
				assert.Equal(t, testUser.Username, resp.User.Username)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}
