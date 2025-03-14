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

// TestCreateTranslation tests the CreateTranslation function
func TestCreateTranslation(t *testing.T) {
	// Setup fixtures
	meaningID := uuid.New()
	entryID := uuid.New()
	languageID := "fr"
	translationText := "bonjour"

	// Create a meaning with proper structure
	meaning := database.Meaning{
		ID:          meaningID,
		EntryID:     entryID,
		Description: "hello or greeting",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	entry := database.Entry{
		ID:        entryID,
		Word:      "hello",
		Type:      database.WordType,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Meanings:  []database.Meaning{meaning},
	}

	// Create request
	createTranslationReq := &request.CreateTranslationRequest{
		LanguageID: languageID,
		Text:       translationText,
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
				// Setup expectations for ListEntries to find meaning
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{entry}, nil).Once()

				// Setup expectations for UpdateEntry to save translation
				mockRepo.On("UpdateEntry", mock.Anything, mock.MatchedBy(func(e *database.Entry) bool {
					// Verify the translation was added to the meaning
					if len(e.Meanings) != 1 || len(e.Meanings[0].Translations) != 1 {
						return false
					}
					translation := e.Meanings[0].Translations[0]
					return translation.MeaningID == meaningID &&
						translation.LanguageID == languageID &&
						translation.Text == translationText
				})).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name: "MeaningNotFound",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				// Return empty entry list to simulate meaning not found
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{}, nil).Once()
				mockLogger.On("Error", "failed to find meaning", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "ListEntriesError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				expectedError := errors.New("database error")
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return(nil, expectedError).Once()
				mockLogger.On("Error", "failed to find meaning", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to find meaning",
		},
		{
			name: "UpdateEntryError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				// Setup expectations for ListEntries to find meaning
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{entry}, nil).Once()

				// Setup expectations for UpdateEntry to fail
				expectedError := errors.New("database error")
				mockRepo.On("UpdateEntry", mock.Anything, mock.Anything).Return(expectedError).Once()
				mockLogger.On("Error", "failed to save translation", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to save translation",
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
			translationService := service.NewTranslationService(mockRepo, mockLogger)

			// Call service
			resp, err := translationService.CreateTranslation(context.Background(), meaningID, createTranslationReq)

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
				assert.Equal(t, meaningID, resp.MeaningID)
				assert.Equal(t, languageID, resp.LanguageID)
				assert.Equal(t, translationText, resp.Text)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mocks.VerifyLoggerMock(mockLogger, t)
		})
	}
}

// TestUpdateTranslation tests the UpdateTranslation function
func TestUpdateTranslation(t *testing.T) {
	// Setup fixtures
	translationID := uuid.New()
	meaningID := uuid.New()
	entryID := uuid.New()
	oldText := "bonjour"
	newText := "salut"

	// Create translation
	translation := database.Translation{
		ID:         translationID,
		MeaningID:  meaningID,
		LanguageID: "fr",
		Text:       oldText,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	meaning := database.Meaning{
		ID:           meaningID,
		EntryID:      entryID,
		Description:  "hello or greeting",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Translations: []database.Translation{translation},
	}

	entry := database.Entry{
		ID:        entryID,
		Word:      "hello",
		Type:      database.WordType,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Meanings:  []database.Meaning{meaning},
	}

	// Create update request
	updateReq := &request.UpdateTranslationRequest{
		Text: newText,
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
				// Setup expectations for ListEntries to find translation
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{entry}, nil).Once()

				// Setup expectations for UpdateEntry
				mockRepo.On("UpdateEntry", mock.Anything, mock.MatchedBy(func(e *database.Entry) bool {
					// Verify translation was updated
					if len(e.Meanings) != 1 || len(e.Meanings[0].Translations) != 1 {
						return false
					}
					updatedTranslation := e.Meanings[0].Translations[0]
					return updatedTranslation.ID == translationID &&
						updatedTranslation.Text == newText
				})).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name: "TranslationNotFound",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				// Return empty entry list to simulate translation not found
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{}, nil).Once()
				mockLogger.On("Error", "failed to find translation", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "ListEntriesError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				expectedError := errors.New("database error")
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return(nil, expectedError).Once()
				mockLogger.On("Error", "failed to find translation", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to find translation",
		},
		{
			name: "UpdateEntryError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				// Setup expectations for ListEntries to find translation
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{entry}, nil).Once()

				// Setup expectations for UpdateEntry to fail
				expectedError := errors.New("database error")
				mockRepo.On("UpdateEntry", mock.Anything, mock.Anything).Return(expectedError).Once()
				mockLogger.On("Error", "failed to update translation", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to update translation",
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
			translationService := service.NewTranslationService(mockRepo, mockLogger)

			// Call service
			resp, err := translationService.UpdateTranslation(context.Background(), translationID, updateReq)

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
				assert.Equal(t, translationID, resp.ID)
				assert.Equal(t, newText, resp.Text)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mocks.VerifyLoggerMock(mockLogger, t)
		})
	}
}

// TestDeleteTranslation tests the DeleteTranslation function
func TestDeleteTranslation(t *testing.T) {
	// Setup fixtures
	translationID := uuid.New()
	meaningID := uuid.New()
	entryID := uuid.New()

	// Create translation
	translation := database.Translation{
		ID:         translationID,
		MeaningID:  meaningID,
		LanguageID: "fr",
		Text:       "bonjour",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	meaning := database.Meaning{
		ID:           meaningID,
		EntryID:      entryID,
		Description:  "hello or greeting",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Translations: []database.Translation{translation},
	}

	entry := database.Entry{
		ID:        entryID,
		Word:      "hello",
		Type:      database.WordType,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Meanings:  []database.Meaning{meaning},
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
				// Setup expectations for ListEntries to find translation
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{entry}, nil).Once()

				// Setup expectations for UpdateEntry
				mockRepo.On("UpdateEntry", mock.Anything, mock.MatchedBy(func(e *database.Entry) bool {
					// Verify translation was removed
					return len(e.Meanings) == 1 && len(e.Meanings[0].Translations) == 0
				})).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name: "TranslationNotFound",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				// Return empty entry list to simulate translation not found
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{}, nil).Once()
				mockLogger.On("Error", "failed to find translation for deletion", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "ListEntriesError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				expectedError := errors.New("database error")
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return(nil, expectedError).Once()
				mockLogger.On("Error", "failed to find translation for deletion", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to find translation",
		},
		{
			name: "UpdateEntryError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				// Setup expectations for ListEntries to find translation
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{entry}, nil).Once()

				// Setup expectations for UpdateEntry to fail
				expectedError := errors.New("database error")
				mockRepo.On("UpdateEntry", mock.Anything, mock.Anything).Return(expectedError).Once()
				mockLogger.On("Error", "failed to delete translation", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to delete translation",
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
			translationService := service.NewTranslationService(mockRepo, mockLogger)

			// Call service
			err := translationService.DeleteTranslation(context.Background(), translationID)

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

// TestListTranslations tests the ListTranslations function
func TestListTranslations(t *testing.T) {
	// Setup fixtures
	meaningID := uuid.New()
	entryID := uuid.New()

	// Create translations for different languages
	translations := []database.Translation{
		{
			ID:         uuid.New(),
			MeaningID:  meaningID,
			LanguageID: "fr",
			Text:       "bonjour",
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		},
		{
			ID:         uuid.New(),
			MeaningID:  meaningID,
			LanguageID: "es",
			Text:       "hola",
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		},
	}

	meaning := database.Meaning{
		ID:           meaningID,
		EntryID:      entryID,
		Description:  "hello or greeting",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Translations: translations,
	}

	entry := database.Entry{
		ID:        entryID,
		Word:      "hello",
		Type:      database.WordType,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Meanings:  []database.Meaning{meaning},
	}

	// Test cases
	testCases := []struct {
		name          string
		languageID    string
		setupMocks    func(*mocks.MockRepository, *mocks.MockLogger)
		expectedCount int
		expectedError bool
		errorContains string
	}{
		{
			name:       "ListAll",
			languageID: "",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{entry}, nil).Once()
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:       "FilterByLanguage",
			languageID: "fr",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{entry}, nil).Once()
			},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name:       "MeaningNotFound",
			languageID: "",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{}, nil).Once()
				mockLogger.On("Error", "failed to find meaning", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name:       "DatabaseError",
			languageID: "",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				expectedError := errors.New("database error")
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return(nil, expectedError).Once()
				mockLogger.On("Error", "failed to find meaning", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to find meaning",
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
			translationService := service.NewTranslationService(mockRepo, mockLogger)

			// Call service
			resp, err := translationService.ListTranslations(context.Background(), meaningID, tc.languageID)

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
				assert.Len(t, resp.Translations, tc.expectedCount)

				// If filtering by language, verify language ID
				if tc.languageID != "" {
					for _, trans := range resp.Translations {
						assert.Equal(t, tc.languageID, trans.LanguageID)
					}
				}
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mocks.VerifyLoggerMock(mockLogger, t)
		})
	}
}

// TestAddTranslationComment tests the AddTranslationComment function
func TestAddTranslationComment(t *testing.T) {
	// Setup fixtures
	translationID := uuid.New()
	meaningID := uuid.New()
	entryID := uuid.New()
	userID := uuid.New()
	commentContent := "Great translation!"

	// Create structure with a translation
	translation := database.Translation{
		ID:         translationID,
		MeaningID:  meaningID,
		LanguageID: "fr",
		Text:       "bonjour",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	meaning := database.Meaning{
		ID:           meaningID,
		EntryID:      entryID,
		Description:  "hello or greeting",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Translations: []database.Translation{translation},
	}

	entry := database.Entry{
		ID:        entryID,
		Word:      "hello",
		Type:      database.WordType,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Meanings:  []database.Meaning{meaning},
	}

	// Create comment request
	commentReq := &request.CreateCommentRequest{
		UserID:  userID,
		Content: commentContent,
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
				// Find the translation
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{entry}, nil).Once()
			},
			expectedError: false,
		},
		{
			name: "TranslationNotFound",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{}, nil).Once()
				mockLogger.On("Error", "failed to find translation", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "DatabaseError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				expectedError := errors.New("database error")
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return(nil, expectedError).Once()
				mockLogger.On("Error", "failed to find translation", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to find translation",
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
			translationService := service.NewTranslationService(mockRepo, mockLogger)

			// Call service
			resp, err := translationService.AddTranslationComment(context.Background(), translationID, commentReq)

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
				assert.Equal(t, commentContent, resp.Content)
				assert.Equal(t, userID, resp.User.ID)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mocks.VerifyLoggerMock(mockLogger, t)
		})
	}
}

// TestToggleTranslationLike tests the ToggleTranslationLike function
func TestToggleTranslationLike(t *testing.T) {
	// Setup fixtures
	translationID := uuid.New()
	meaningID := uuid.New()
	entryID := uuid.New()
	userID := uuid.New()

	// Create structure with a translation
	translation := database.Translation{
		ID:         translationID,
		MeaningID:  meaningID,
		LanguageID: "fr",
		Text:       "bonjour",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	meaning := database.Meaning{
		ID:           meaningID,
		EntryID:      entryID,
		Description:  "hello or greeting",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Translations: []database.Translation{translation},
	}

	entry := database.Entry{
		ID:        entryID,
		Word:      "hello",
		Type:      database.WordType,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Meanings:  []database.Meaning{meaning},
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
				// Find the translation
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{entry}, nil).Once()
				mockLogger.On("Info", "like processed", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: false,
		},
		{
			name: "TranslationNotFound",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return([]database.Entry{}, nil).Once()
				mockLogger.On("Error", "failed to find translation", mock.Anything, mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "not found",
		},
		{
			name: "DatabaseError",
			setupMocks: func(mockRepo *mocks.MockRepository, mockLogger *mocks.MockLogger) {
				expectedError := errors.New("database error")
				mockRepo.On("ListEntries", mock.Anything, mock.Anything).Return(nil, expectedError).Once()
				mockLogger.On("Error", "failed to find translation", mock.Anything).Return().Once()
			},
			expectedError: true,
			errorContains: "failed to find translation",
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
			translationService := service.NewTranslationService(mockRepo, mockLogger)

			// Call service
			err := translationService.ToggleTranslationLike(context.Background(), translationID, userID)

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
