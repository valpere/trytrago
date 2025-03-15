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
	"github.com/valpere/trytrago/test/mocks"
)

// setupEntryService sets up a mock repository and logger for entry service tests
func setupEntryService(t *testing.T) (service.EntryService, *mocks.MockRepository, *mocks.MockLogger) {
	mockRepo := new(mocks.MockRepository)
	mockLogger := mocks.SetupLoggerMock()

	// Create the service
	entryService := service.NewEntryService(mockRepo, mockLogger)

	return entryService, mockRepo, mockLogger
}

// TestCreateEntry tests the CreateEntry function
func TestCreateEntry(t *testing.T) {
	// Setup fixtures
	testWord := "test"
	testType := "WORD"
	testPronunciation := "test"

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
				mockRepo.On("CreateEntry", mock.Anything, mock.MatchedBy(func(e *database.Entry) bool {
					return e.Word == testWord &&
						string(e.Type) == testType &&
						e.Pronunciation == testPronunciation
				})).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name: "RepositoryError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				dbErr := errors.New("database error")
				mockRepo.On("CreateEntry", mock.Anything, mock.Anything).Return(dbErr).Once()
				mockLogger.On("Error", "failed to create entry", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to create entry",
		},
		{
			name: "DuplicateEntry",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("CreateEntry", mock.Anything, mock.Anything).Return(database.ErrDuplicateEntry).Once()
				mockLogger.On("Error", "failed to create entry", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "duplicate entry",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup service and mocks
			entryService, mockRepo, mockLogger := setupEntryService(t)

			// Setup specific test case expectations
			tc.setupMocks(mockRepo, mockLogger)

			// Create request DTO
			req := &request.CreateEntryRequest{
				Word:          testWord,
				Type:          testType,
				Pronunciation: testPronunciation,
			}

			// Call service
			resp, err := entryService.CreateEntry(context.Background(), req)

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
				assert.Equal(t, testWord, resp.Word)
				assert.Equal(t, testType, resp.Type)
				assert.Equal(t, testPronunciation, resp.Pronunciation)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetEntryByID tests the GetEntryByID function
func TestGetEntryByID(t *testing.T) {
	// Setup fixtures
	testID := uuid.New()
	testWord := "test"
	testType := database.WordType
	testEntry := &database.Entry{
		ID:            testID,
		Word:          testWord,
		Type:          testType,
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
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
				mockRepo.On("GetEntryByID", mock.Anything, testID).Return(testEntry, nil).Once()
			},
			expectedError: false,
		},
		{
			name: "EntryNotFound",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("GetEntryByID", mock.Anything, testID).Return(nil, database.ErrEntryNotFound).Once()
				mockLogger.On("Error", "failed to get entry", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "DatabaseError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				dbErr := errors.New("database error")
				mockRepo.On("GetEntryByID", mock.Anything, testID).Return(nil, dbErr).Once()
				mockLogger.On("Error", "failed to get entry", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to get entry",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup service and mocks
			entryService, mockRepo, mockLogger := setupEntryService(t)

			// Setup specific test case expectations
			tc.setupMocks(mockRepo, mockLogger)

			// Call service
			resp, err := entryService.GetEntryByID(context.Background(), testID)

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
				assert.Equal(t, testID, resp.ID)
				assert.Equal(t, testWord, resp.Word)
				assert.Equal(t, string(testType), resp.Type)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUpdateEntry tests the UpdateEntry function
func TestUpdateEntry(t *testing.T) {
	// Setup fixtures
	testID := uuid.New()
	originalWord := "original"
	updatedWord := "updated"
	testType := database.WordType
	updatedType := database.PhraseType
	testEntry := &database.Entry{
		ID:            testID,
		Word:          originalWord,
		Type:          testType,
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
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
				mockRepo.On("GetEntryByID", mock.Anything, testID).Return(testEntry, nil).Once()
				mockRepo.On("UpdateEntry", mock.Anything, mock.MatchedBy(func(entry *database.Entry) bool {
					return entry.ID == testID &&
						entry.Word == updatedWord &&
						entry.Type == updatedType
				})).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name: "EntryNotFound",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("GetEntryByID", mock.Anything, testID).Return(nil, database.ErrEntryNotFound).Once()
				mockLogger.On("Error", "failed to get entry for update", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "UpdateError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("GetEntryByID", mock.Anything, testID).Return(testEntry, nil).Once()
				dbErr := errors.New("database error")
				mockRepo.On("UpdateEntry", mock.Anything, mock.Anything).Return(dbErr).Once()
				mockLogger.On("Error", "failed to update entry", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to update entry",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup service and mocks
			entryService, mockRepo, mockLogger := setupEntryService(t)

			// Setup specific test case expectations
			tc.setupMocks(mockRepo, mockLogger)

			// Create update request
			updateReq := &request.UpdateEntryRequest{
				Word:          updatedWord,
				Type:          string(updatedType),
				Pronunciation: "updated",
			}

			// Call service
			resp, err := entryService.UpdateEntry(context.Background(), testID, updateReq)

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
				assert.Equal(t, testID, resp.ID)
				assert.Equal(t, updatedWord, resp.Word)
				assert.Equal(t, string(updatedType), resp.Type)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}
