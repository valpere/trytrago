package repository_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/database/repository"
	"github.com/valpere/trytrago/domain/database/repository/sqlite"
)

// SQLiteRepositoryTestSuite contains tests for the SQLite repository implementation
type SQLiteRepositoryTestSuite struct {
	suite.Suite
	repo   repository.Repository
	ctx    context.Context
	dbPath string
}

// SetupSuite initializes the test suite
func (s *SQLiteRepositoryTestSuite) SetupSuite() {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "trytrago-test-*")
	require.NoError(s.T(), err, "Failed to create temp directory")

	// Create database path
	s.dbPath = filepath.Join(tempDir, "trytrago_test.db")

	// Create a new context with timeout
	s.ctx = context.Background()

	// Create connection options
	opts := repository.Options{
		Driver:          "sqlite",
		Database:        s.dbPath,
		MaxIdleConns:    1,
		MaxOpenConns:    1,
		ConnMaxLifetime: 5 * time.Minute,
		Debug:           false,
	}

	// Create repository
	s.repo, err = sqlite.NewRepository(s.ctx, opts)
	require.NoError(s.T(), err, "Failed to create repository")

	// Verify connection
	err = s.repo.Ping(s.ctx)
	require.NoError(s.T(), err, "Failed to ping database")

	// Initialize test database schema
	s.setupTestSchema()
}

// setupTestSchema initializes the test database schema
func (s *SQLiteRepositoryTestSuite) setupTestSchema() {
	// Get DB connection and create test tables
	db, err := s.repo.GetDB()
	require.NoError(s.T(), err, "Failed to get database connection")

	// Create tables using auto-migrate
	err = db.AutoMigrate(&database.Entry{}, &database.Meaning{}, &database.Example{}, &database.Translation{}, &database.ChangeHistory{})
	require.NoError(s.T(), err, "Failed to create database schema")
}

// TearDownSuite cleans up after the test suite
func (s *SQLiteRepositoryTestSuite) TearDownSuite() {
	// Close the repo connection
	if s.repo != nil {
		err := s.repo.Close()
		assert.NoError(s.T(), err, "Failed to close repository")
	}

	// Delete the test database file
	if s.dbPath != "" {
		err := os.Remove(s.dbPath)
		if err != nil && !os.IsNotExist(err) {
			s.T().Logf("Failed to remove test database file: %v", err)
		}
	}
}

// TestCreateAndRetrieveEntry tests both CreateEntry and GetEntryByID methods
func (s *SQLiteRepositoryTestSuite) TestCreateAndRetrieveEntry() {
	// Create a test entry
	entry := &database.Entry{
		ID:            uuid.New(),
		Word:          "sqlite_test",
		Type:          database.WordType,
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Create the entry
	err := s.repo.CreateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to create entry")

	// Verify the entry was created
	retrievedEntry, err := s.repo.GetEntryByID(s.ctx, entry.ID)
	assert.NoError(s.T(), err, "Failed to retrieve entry")
	assert.NotNil(s.T(), retrievedEntry, "Retrieved entry should not be nil")
	assert.Equal(s.T(), entry.ID, retrievedEntry.ID, "Entry ID should match")
	assert.Equal(s.T(), entry.Word, retrievedEntry.Word, "Entry word should match")
	assert.Equal(s.T(), entry.Type, retrievedEntry.Type, "Entry type should match")
	assert.Equal(s.T(), entry.Pronunciation, retrievedEntry.Pronunciation, "Entry pronunciation should match")
}

// TestCreateEntryWithNestedData tests creating an entry with meanings, examples, and translations
func (s *SQLiteRepositoryTestSuite) TestCreateEntryWithNestedData() {
	// Create a test entry with nested data
	entryID := uuid.New()
	meaningID := uuid.New()
	entry := &database.Entry{
		ID:            entryID,
		Word:          "nested_test",
		Type:          database.WordType,
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		Meanings: []database.Meaning{
			{
				ID:             meaningID,
				EntryID:        entryID,
				Description:    "Test meaning",
				PartOfSpeechId: uuid.New(),
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				Examples: []database.Example{
					{
						ID:        uuid.New(),
						MeaningID: meaningID,
						Text:      "Example usage",
						Context:   "In a sentence",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					},
				},
				Translations: []database.Translation{
					{
						ID:         uuid.New(),
						MeaningID:  meaningID,
						LanguageID: "es",
						Text:       "Prueba",
						CreatedAt:  time.Now().UTC(),
						UpdatedAt:  time.Now().UTC(),
					},
				},
			},
		},
	}

	// Create the entry
	err := s.repo.CreateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to create entry with nested data")

	// Verify the entry was created with all nested data
	retrievedEntry, err := s.repo.GetEntryByID(s.ctx, entry.ID)
	assert.NoError(s.T(), err, "Failed to retrieve entry")
	assert.NotNil(s.T(), retrievedEntry, "Retrieved entry should not be nil")

	// Verify meanings
	assert.Len(s.T(), retrievedEntry.Meanings, 1, "Should have 1 meaning")
	meaning := retrievedEntry.Meanings[0]
	assert.Equal(s.T(), meaningID, meaning.ID, "Meaning ID should match")
	assert.Equal(s.T(), "Test meaning", meaning.Description, "Meaning description should match")

	// Verify examples
	assert.Len(s.T(), meaning.Examples, 1, "Should have 1 example")
	example := meaning.Examples[0]
	assert.Equal(s.T(), "Example usage", example.Text, "Example text should match")
	assert.Equal(s.T(), "In a sentence", example.Context, "Example context should match")

	// Verify translations
	assert.Len(s.T(), meaning.Translations, 1, "Should have 1 translation")
	translation := meaning.Translations[0]
	assert.Equal(s.T(), "es", translation.LanguageID, "Translation language ID should match")
	assert.Equal(s.T(), "Prueba", translation.Text, "Translation text should match")
}

// TestUpdateEntry tests the UpdateEntry method
func (s *SQLiteRepositoryTestSuite) TestUpdateEntry() {
	// Create a test entry
	entry := &database.Entry{
		ID:            uuid.New(),
		Word:          "update_test_sqlite",
		Type:          database.WordType,
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Create the entry
	err := s.repo.CreateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to create entry")

	// Update the entry
	updatedWord := "updated_test_sqlite"
	updatedType := database.PhraseType
	updatedPronunciation := "updated_sqlite"

	entry.Word = updatedWord
	entry.Type = updatedType
	entry.Pronunciation = updatedPronunciation
	entry.UpdatedAt = time.Now().UTC()

	err = s.repo.UpdateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to update entry")

	// Verify the entry was updated
	retrievedEntry, err := s.repo.GetEntryByID(s.ctx, entry.ID)
	assert.NoError(s.T(), err, "Failed to retrieve updated entry")
	assert.NotNil(s.T(), retrievedEntry, "Retrieved entry should not be nil")
	assert.Equal(s.T(), updatedWord, retrievedEntry.Word, "Entry word should be updated")
	assert.Equal(s.T(), updatedType, retrievedEntry.Type, "Entry type should be updated")
	assert.Equal(s.T(), updatedPronunciation, retrievedEntry.Pronunciation, "Entry pronunciation should be updated")
}

// TestUpdateEntryWithNewMeaning tests updating an entry by adding a new meaning
func (s *SQLiteRepositoryTestSuite) TestUpdateEntryWithNewMeaning() {
	// Create a test entry
	entryID := uuid.New()
	entry := &database.Entry{
		ID:            entryID,
		Word:          "meaning_update_test",
		Type:          database.WordType,
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		Meanings:      []database.Meaning{},
	}

	// Create the entry
	err := s.repo.CreateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to create entry")

	// Add a meaning to the entry
	meaningID := uuid.New()
	meaning := database.Meaning{
		ID:             meaningID,
		EntryID:        entryID,
		Description:    "New test meaning",
		PartOfSpeechId: uuid.New(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	entry.Meanings = append(entry.Meanings, meaning)
	entry.UpdatedAt = time.Now().UTC()

	// Update the entry
	err = s.repo.UpdateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to update entry with new meaning")

	// Verify the entry was updated with the new meaning
	retrievedEntry, err := s.repo.GetEntryByID(s.ctx, entry.ID)
	assert.NoError(s.T(), err, "Failed to retrieve updated entry")
	assert.NotNil(s.T(), retrievedEntry, "Retrieved entry should not be nil")
	assert.Len(s.T(), retrievedEntry.Meanings, 1, "Entry should have 1 meaning")

	// Verify meaning details
	retrievedMeaning := retrievedEntry.Meanings[0]
	assert.Equal(s.T(), meaningID, retrievedMeaning.ID, "Meaning ID should match")
	assert.Equal(s.T(), "New test meaning", retrievedMeaning.Description, "Meaning description should match")
}

// TestDeleteEntry tests the DeleteEntry method
func (s *SQLiteRepositoryTestSuite) TestDeleteEntry() {
	// Create a test entry
	entry := &database.Entry{
		ID:            uuid.New(),
		Word:          "delete_test_sqlite",
		Type:          database.WordType,
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Create the entry
	err := s.repo.CreateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to create entry")

	// Delete the entry
	err = s.repo.DeleteEntry(s.ctx, entry.ID)
	assert.NoError(s.T(), err, "Failed to delete entry")

	// Verify the entry was deleted
	_, err = s.repo.GetEntryByID(s.ctx, entry.ID)
	assert.Error(s.T(), err, "Entry should be deleted")
	assert.True(s.T(), database.IsNotFoundError(err), "Should return not found error")
}

// TestListEntries tests the ListEntries method
func (s *SQLiteRepositoryTestSuite) TestListEntries() {
	// Create multiple test entries
	entries := []database.Entry{
		{
			ID:            uuid.New(),
			Word:          "sqlite_list_test_1",
			Type:          database.WordType,
			Pronunciation: "test1",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            uuid.New(),
			Word:          "sqlite_list_test_2",
			Type:          database.WordType,
			Pronunciation: "test2",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            uuid.New(),
			Word:          "sqlite_list_test_3",
			Type:          database.PhraseType,
			Pronunciation: "test3",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
	}

	// Create each entry
	for i := range entries {
		err := s.repo.CreateEntry(s.ctx, &entries[i])
		assert.NoError(s.T(), err, "Failed to create test entry")
	}

	// Test listing all entries
	params := repository.ListParams{
		Limit:  100,
		Offset: 0,
	}

	results, err := s.repo.ListEntries(s.ctx, params)
	assert.NoError(s.T(), err, "Failed to list entries")
	assert.GreaterOrEqual(s.T(), len(results), 3, "Should have at least 3 entries")

	// Test filtering by word
	params = repository.ListParams{
		Limit:   100,
		Offset:  0,
		Filters: map[string]interface{}{"word LIKE ?": "%sqlite_list_test_%"},
	}

	results, err = s.repo.ListEntries(s.ctx, params)
	assert.NoError(s.T(), err, "Failed to list entries with word filter")
	assert.Len(s.T(), results, 3, "Should have exactly 3 entries matching the filter")

	// Test filtering by type
	params = repository.ListParams{
		Limit:   100,
		Offset:  0,
		Filters: map[string]interface{}{"type = ?": string(database.PhraseType)},
	}

	results, err = s.repo.ListEntries(s.ctx, params)
	assert.NoError(s.T(), err, "Failed to list entries with type filter")
	assert.GreaterOrEqual(s.T(), len(results), 1, "Should have at least 1 entry of type PHRASE")

	// Verify all results are of type PHRASE
	for _, entry := range results {
		if entry.Word == "sqlite_list_test_3" {
			assert.Equal(s.T(), database.PhraseType, entry.Type, "Entry should be of type PHRASE")
		}
	}

	// Test pagination
	params = repository.ListParams{
		Limit:  2,
		Offset: 0,
	}

	results, err = s.repo.ListEntries(s.ctx, params)
	assert.NoError(s.T(), err, "Failed to list entries with pagination")
	assert.Len(s.T(), results, 2, "Should have 2 entries due to limit")

	// Test next page
	params.Offset = 2
	nextResults, err := s.repo.ListEntries(s.ctx, params)
	assert.NoError(s.T(), err, "Failed to list entries for second page")
	assert.NotEmpty(s.T(), nextResults, "Second page should not be empty")
}

// TestFindTranslations tests the FindTranslations method
func (s *SQLiteRepositoryTestSuite) TestFindTranslations() {
	// Create an entry with translations
	entryID := uuid.New()
	meaningID := uuid.New()

	// Create the entry with a meaning and translations
	entry := &database.Entry{
		ID:            entryID,
		Word:          "hello",
		Type:          database.WordType,
		Pronunciation: "həˈlō",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		Meanings: []database.Meaning{
			{
				ID:             meaningID,
				EntryID:        entryID,
				Description:    "Used as a greeting",
				PartOfSpeechId: uuid.New(),
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
				Translations: []database.Translation{
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
				},
			},
		},
	}

	// Create the entry
	err := s.repo.CreateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to create entry with translations")

	// Test finding translations for "hello" in French
	translations, err := s.repo.FindTranslations(s.ctx, "hello", "fr")
	assert.NoError(s.T(), err, "Failed to find translations")
	assert.Len(s.T(), translations, 1, "Should have 1 French translation")
	assert.Equal(s.T(), "bonjour", translations[0].Text, "Translation text should match")

	// Test finding translations for "hello" in Spanish
	translations, err = s.repo.FindTranslations(s.ctx, "hello", "es")
	assert.NoError(s.T(), err, "Failed to find translations")
	assert.Len(s.T(), translations, 1, "Should have 1 Spanish translation")
	assert.Equal(s.T(), "hola", translations[0].Text, "Translation text should match")

	// Test case insensitivity
	translations, err = s.repo.FindTranslations(s.ctx, "Hello", "fr")
	assert.NoError(s.T(), err, "Failed to find translations with case insensitive search")
	assert.Len(s.T(), translations, 1, "Should have 1 French translation despite case difference")

	// Test non-existent language
	translations, err = s.repo.FindTranslations(s.ctx, "hello", "de")
	assert.NoError(s.T(), err, "Should not error for non-existent language")
	assert.Empty(s.T(), translations, "Should have no German translations")

	// Test non-existent word
	translations, err = s.repo.FindTranslations(s.ctx, "goodbye", "fr")
	assert.NoError(s.T(), err, "Should not error for non-existent word")
	assert.Empty(s.T(), translations, "Should have no translations for non-existent word")
}

// TestRecordAndRetrieveChangeHistory tests the RecordChange and GetEntryHistory methods
func (s *SQLiteRepositoryTestSuite) TestRecordAndRetrieveChangeHistory() {
	// Create an entry to track history for
	entryID := uuid.New()
	entry := &database.Entry{
		ID:            entryID,
		Word:          "history_test",
		Type:          database.WordType,
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Create the entry
	err := s.repo.CreateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to create entry")

	// Record change history items
	changes := []database.ChangeHistory{
		{
			ID:        uuid.New(),
			EntryID:   entryID,
			Action:    "create",
			Data:      []byte(`{"word":"history_test"}`),
			UserID:    uuid.New(),
			CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
		},
		{
			ID:        uuid.New(),
			EntryID:   entryID,
			Action:    "update",
			Data:      []byte(`{"word":"history_test_updated"}`),
			UserID:    uuid.New(),
			CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
		},
	}

	// Record each change
	for _, change := range changes {
		err := s.repo.RecordChange(s.ctx, &change)
		assert.NoError(s.T(), err, "Failed to record change")
	}

	// Get the entry history
	history, err := s.repo.GetEntryHistory(s.ctx, entryID)
	assert.NoError(s.T(), err, "Failed to get entry history")
	assert.Len(s.T(), history, 2, "Should have 2 history items")

	// Check that changes are sorted by created_at in descending order (most recent first)
	assert.Equal(s.T(), "update", history[0].Action, "Most recent change should be first")
	assert.Equal(s.T(), "create", history[1].Action, "Oldest change should be last")
}

// TestSQLiteRepository runs the test suite
func TestSQLiteRepository(t *testing.T) {
	// Skip tests if we're not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration tests. Set INTEGRATION_TEST=true to run")
	}

	suite.Run(t, new(SQLiteRepositoryTestSuite))
}
