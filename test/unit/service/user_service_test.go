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

	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/model"
	"github.com/valpere/trytrago/test/mocks"
)

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
		setupMocks    func(*mocks.MockRepository, *mocks.MockLogger)
		expectedError bool
		errorContains string
	}{
		{
			name: "Success",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *model.User) bool {
					// Password should be hashed
					return user.Username == testUsername &&
						user.Email == testEmail &&
						user.Password != testPassword // Password should be hashed
				})).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name: "RepositoryError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				expectedError := errors.New("database error")
				mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(expectedError).Once()
				mockLogger.On("Error", "failed to create user", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to create user",
		},
		{
			name: "DuplicateUsername",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(database.ErrDuplicateEntry).Once()
				mockLogger.On("Error", "failed to create user", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "duplicate entry",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(mocks.MockRepository)
			mockLogger := mocks.SetupLoggerMock()

			// Configure expectations
			tc.setupMocks(mockRepo, mockLogger)

			// Create service
			userService := service.NewUserService(mockRepo, mockLogger)

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
			mocks.VerifyLoggerMock(mockLogger, t)
		})
	}
}

// TestAuthenticate tests the Authenticate function
func TestAuthenticate(t *testing.T) {
	// Setup fixtures
	testID := uuid.New()
	testUsername := "testuser"
	testPassword := "password"
	hashedPassword := "$2a$10$dBR5d8VTLjQvQOPiwbHCzuQUEVLvtvVSbG2pJUT3c4DHmfVCJNpou" // "password" hashed
	testUser := &model.User{
		ID:        testID,
		Username:  testUsername,
		Password:  hashedPassword,
		Email:     "test@example.com",
		Role:      model.RoleUser,
		IsActive:  true,
		LastLogin: nil,
	}

	// Create request for successful auth
	validAuthReq := &request.AuthRequest{
		Username: testUsername,
		Password: testPassword,
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
				mockLogger.On("Warn", "authentication failed: user not found", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "invalid credentials",
		},
		{
			name:    "InvalidPassword",
			request: invalidPasswordReq,
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("GetUserByUsername", mock.Anything, testUsername).Return(testUser, nil).Once()
				mockLogger.On("Warn", "authentication failed: invalid password", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "invalid credentials",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(mocks.MockRepository)
			mockLogger := mocks.SetupLoggerMock()

			// Configure expectations
			tc.setupMocks(mockRepo, mockLogger)

			// Create service
			userService := service.NewUserService(mockRepo, mockLogger)

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
			mocks.VerifyLoggerMock(mockLogger, t)
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
		setupMocks    func(*mocks.MockRepository, *mocks.MockLogger)
		expectedError bool
		errorContains string
	}{
		{
			name: "Success",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
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
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(nil, database.ErrNotFound).Once()
				mockLogger.On("Error", "failed to get user for update", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "DuplicateUsername",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(testUser, nil).Once()
				mockRepo.On("UpdateUser", mock.Anything, mock.Anything).Return(database.ErrDuplicateEntry).Once()
				mockLogger.On("Error", "failed to update user", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "duplicate entry",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(mocks.MockRepository)
			mockLogger := mocks.SetupLoggerMock()

			// Configure expectations
			tc.setupMocks(mockRepo, mockLogger)

			// Create service
			userService := service.NewUserService(mockRepo, mockLogger)

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
			mocks.VerifyLoggerMock(mockLogger, t)
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
		setupMocks    func(*mocks.MockRepository, *mocks.MockLogger)
		expectedError bool
		errorContains string
	}{
		{
			name: "Success",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("DeleteUser", mock.Anything, testID).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name: "UserNotFound",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("DeleteUser", mock.Anything, testID).Return(database.ErrNotFound).Once()
				mockLogger.On("Error", "failed to delete user", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "DatabaseError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				expectedError := errors.New("database error")
				mockRepo.On("DeleteUser", mock.Anything, testID).Return(expectedError).Once()
				mockLogger.On("Error", "failed to delete user", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to delete user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(mocks.MockRepository)
			mockLogger := mocks.SetupLoggerMock()

			// Configure expectations
			tc.setupMocks(mockRepo, mockLogger)

			// Create service
			userService := service.NewUserService(mockRepo, mockLogger)

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
			mocks.VerifyLoggerMock(mockLogger, t)
		})
	}
}

// TestRefreshToken tests the RefreshToken function
func TestRefreshToken(t *testing.T) {
	// Setup fixtures
	testID := uuid.New()
	testUsername := "testuser"
	testRefreshToken := "valid-refresh-token"
	testUser := &model.User{
		ID:       testID,
		Username: testUsername,
		Email:    "test@example.com",
		Role:     model.RoleUser,
		IsActive: true,
	}

	// Test cases
	testCases := []struct {
		name          string
		token         string
		setupMocks    func(*mocks.MockRepository, *mocks.MockLogger)
		expectedError bool
		errorContains string
	}{
		{
			name:  "Success",
			token: testRefreshToken,
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				// We need to mock the jwt validation that would happen in auth.ValidateRefreshToken
				// This would normally return the user ID from the token
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(testUser, nil).Once()
			},
			expectedError: false,
		},
		{
			name:  "InvalidToken",
			token: "invalid-token",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockLogger.On("Warn", "invalid refresh token", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "invalid refresh token",
		},
		{
			name:  "UserNotFound",
			token: testRefreshToken,
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("GetUserByID", mock.Anything, testID).Return(nil, database.ErrNotFound).Once()
				mockLogger.On("Error", "failed to get user for token refresh", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "invalid token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(mocks.MockRepository)
			mockLogger := mocks.SetupLoggerMock()

			// Configure expectations
			tc.setupMocks(mockRepo, mockLogger)

			// Create service
			userService := service.NewUserService(mockRepo, mockLogger)

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
				assert.Equal(t, testUsername, resp.User.Username)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mocks.VerifyLoggerMock(mockLogger, t)
		})
	}
}
