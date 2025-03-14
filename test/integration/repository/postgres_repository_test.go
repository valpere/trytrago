package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/database/repository"
	"github.com/valpere/trytrago/domain/database/repository/postgres"
)

// PostgresRepositoryTestSuite contains tests for the PostgreSQL repository implementation
type PostgresRepositoryTestSuite struct {
	suite.Suite
	repo repository.Repository
	ctx  context.Context
}

// SetupSuite initializes the test suite
func (s *PostgresRepositoryTestSuite) SetupSuite() {
	// Use environment variables or default values for database connection
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := 5432
	dbUser := os.Getenv("TEST_DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("TEST_DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}

	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = "trytrago_test"
	}

	// Create a new context with timeout
	s.ctx = context.Background()

	// Create connection options
	opts := repository.Options{
		Driver:          "postgres",
		Host:            dbHost,
		Port:            dbPort,
		Database:        dbName,
		Username:        dbUser,
		Password:        dbPassword,
		SSLMode:         "disable",
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		Debug:           false,
	}

	// Create repository
	var err error
	s.repo, err = postgres.NewRepository(s.ctx, opts)
	require.NoError(s.T(), err, "Failed to create repository")

	// Verify connection
	err = s.repo.Ping(s.ctx)
	require.NoError(s.T(), err, "Failed to ping database")

	// Initialize test database schema
	s.setupTestSchema()
}

// setupTestSchema initializes the test database schema
func (s *PostgresRepositoryTestSuite) setupTestSchema() {
	// Get DB connection and create test tables
	db, err := s.repo.GetDB()
	require.NoError(s.T(), err, "Failed to get database connection")

	// Drop existing tables if they exist
	err = db.Exec(`DROP TABLE IF EXISTS translations CASCADE`).Error
	require.NoError(s.T(), err, "Failed to drop translations table")

	err = db.Exec(`DROP TABLE IF EXISTS examples CASCADE`).Error
	require.NoError(s.T(), err, "Failed to drop examples table")

	err = db.Exec(`DROP TABLE IF EXISTS meanings CASCADE`).Error
	require.NoError(s.T(), err, "Failed to drop meanings table")

	err = db.Exec(`DROP TABLE IF EXISTS entries CASCADE`).Error
	require.NoError(s.T(), err, "Failed to drop entries table")

	err = db.Exec(`DROP TABLE IF EXISTS change_histories CASCADE`).Error
	require.NoError(s.T(), err, "Failed to drop change_histories table")

	// Create tables
	err = db.AutoMigrate(&database.Entry{}, &database.Meaning{}, &database.Example{}, &database.Translation{}, &database.ChangeHistory{})
	require.NoError(s.T(), err, "Failed to create database schema")
}

// TearDownSuite cleans up after the test suite
func (s *PostgresRepositoryTestSuite) TearDownSuite() {
	if s.repo != nil {
		err := s.repo.Close()
		assert.NoError(s.T(), err, "Failed to close repository")
	}
}

// TestCreateEntry tests the CreateEntry method
func (s *PostgresRepositoryTestSuite) TestCreateEntry() {
	// Create a test entry
	entry := &database.Entry{
		ID:            uuid.New(),
		Word:          "test",
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

// TestCreateEntryWithMeanings tests creating an entry with meanings
func (s *PostgresRepositoryTestSuite) TestCreateEntryWithMeanings() {
	// Create a test entry with meanings
	entry := &database.Entry{
		ID:            uuid.New(),
		Word:          "test_with_meanings",
		Type:          database.WordType,
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		Meanings: []database.Meaning{
			{
				ID:             uuid.New(),
				Description:    "Test meaning 1",
				PartOfSpeechId: uuid.New(),
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
			},
			{
				ID:             uuid.New(),
				Description:    "Test meaning 2",
				PartOfSpeechId: uuid.New(),
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
			},
		},
	}

	// Ensure EntryID is set correctly
	for i := range entry.Meanings {
		entry.Meanings[i].EntryID = entry.ID
	}

	// Create the entry
	err := s.repo.CreateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to create entry with meanings")

	// Verify the entry was created with meanings
	retrievedEntry, err := s.repo.GetEntryByID(s.ctx, entry.ID)
	assert.NoError(s.T(), err, "Failed to retrieve entry")
	assert.NotNil(s.T(), retrievedEntry, "Retrieved entry should not be nil")
	assert.Equal(s.T(), entry.ID, retrievedEntry.ID, "Entry ID should match")
	assert.Equal(s.T(), entry.Word, retrievedEntry.Word, "Entry word should match")

	// Verify meanings
	assert.Len(s.T(), retrievedEntry.Meanings, 2, "Entry should have 2 meanings")

	// Map meanings by ID for easier comparison
	meaningMap := make(map[uuid.UUID]database.Meaning)
	for _, m := range retrievedEntry.Meanings {
		meaningMap[m.ID] = m
	}

	// Verify each meaning
	for _, originalMeaning := range entry.Meanings {
		retrievedMeaning, ok := meaningMap[originalMeaning.ID]
		assert.True(s.T(), ok, "Meaning should exist")
		assert.Equal(s.T(), originalMeaning.Description, retrievedMeaning.Description, "Meaning description should match")
		assert.Equal(s.T(), originalMeaning.PartOfSpeechId, retrievedMeaning.PartOfSpeechId, "Part of speech ID should match")
	}
}

// TestUpdateEntry tests the UpdateEntry method
func (s *PostgresRepositoryTestSuite) TestUpdateEntry() {
	// Create a test entry
	entry := &database.Entry{
		ID:            uuid.New(),
		Word:          "update_test",
		Type:          database.WordType,
		Pronunciation: "test",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Create the entry
	err := s.repo.CreateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to create entry")

	// Update the entry
	entry.Word = "updated_test"
	entry.Type = database.PhraseType
	entry.Pronunciation = "updated"
	entry.UpdatedAt = time.Now().UTC()

	err = s.repo.UpdateEntry(s.ctx, entry)
	assert.NoError(s.T(), err, "Failed to update entry")

	// Verify the entry was updated
	retrievedEntry, err := s.repo.GetEntryByID(s.ctx, entry.ID)
	assert.NoError(s.T(), err, "Failed to retrieve entry")
	assert.NotNil(s.T(), retrievedEntry, "Retrieved entry should not be nil")
	assert.Equal(s.T(), entry.Word, retrievedEntry.Word, "Entry word should be updated")
	assert.Equal(s.T(), entry.Type, retrievedEntry.Type, "Entry type should be updated")
	assert.Equal(s.T(), entry.Pronunciation, retrievedEntry.Pronunciation, "Entry pronunciation should be updated")
}

// TestDeleteEntry tests the DeleteEntry method
func (s *PostgresRepositoryTestSuite) TestDeleteEntry() {
	// Create a test entry
	entry := &database.Entry{
		ID:            uuid.New(),
		Word:          "delete_test",
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
	assert.True(s.T(), database.IsNotFoundError(err), "Error should be 'not found'")
}

// TestListEntries tests the ListEntries method
func (s *PostgresRepositoryTestSuite) TestListEntries(t *testing.T) {
	// Create test entries
	entries := []database.Entry{
		{
			ID:            uuid.New(),
			Word:          "list_test_1",
			Type:          database.WordType,
			Pronunciation: "test1",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            uuid.New(),
			Word:          "list_test_2",
			Type:          database.WordType,
			Pronunciation: "test2",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            uuid.New(),
			Word:          "list_test_3",
			Type:          database.PhraseType,
			Pronunciation: "test3",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
	}

	// Create entries
	for i := range entries {
		err := s.repo.CreateEntry(s.ctx, &entries[i])
		assert.NoError(s.T(), err, "Failed to create test entry")
	}

	// Test listing entries with no filters
	t.Run("NoFilters", func(t *testing.T) {
		params := repository.ListParams{
			Limit:  10,
			Offset: 0,
		}

		results, err := s.repo.ListEntries(s.ctx, params)
		assert.NoError(t, err, "Failed to list entries")
		assert.GreaterOrEqual(t, len(results), 3, "Should have at least 3 entries")
	})

	// Test listing entries with word filter
	t.Run("WordFilter", func(t *testing.T) {
		params := repository.ListParams{
			Limit:   10,
			Offset:  0,
			Filters: map[string]interface{}{"word LIKE ?": "%list_test_%"},
		}

		results, err := s.repo.ListEntries(s.ctx, params)
		assert.NoError(t, err, "Failed to list entries")
		assert.Len(t, results, 3, "Should have 3 entries matching filter")
	})

	// Test listing entries with type filter
	t.Run("TypeFilter", func(t *testing.T) {
		params := repository.ListParams{
			Limit:   10,
			Offset:  0,
			Filters: map[string]interface{}{"type = ?": string(database.PhraseType)},
		}

		results, err := s.repo.ListEntries(s.ctx, params)
		assert.NoError(t, err, "Failed to list entries")
		assert.GreaterOrEqual(t, len(results), 1, "Should have at least 1 entry of type PHRASE")

		for _, entry := range results {
			assert.Equal(t, database.PhraseType, entry.Type, "Entry should be of type PHRASE")
		}
	})

	// Test pagination
	t.Run("Pagination", func(t *testing.T) {
		params := repository.ListParams{
			Limit:  2,
			Offset: 0,
		}

		results, err := s.repo.ListEntries(s.ctx, params)
		assert.NoError(t, err, "Failed to list entries")
		assert.Len(t, results, 2, "Should have 2 entries (limit)")

		params.Offset = 2
		results, err = s.repo.ListEntries(s.ctx, params)
		assert.NoError(t, err, "Failed to list entries")
		assert.GreaterOrEqual(t, len(results), 1, "Should have at least 1 entry (next page)")
	})
}

// Add more tests for the repository methods...

// TestPostgresRepository runs the test suite
func TestPostgresRepository(t *testing.T) {
	// Skip tests if we're not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration tests. Set INTEGRATION_TEST=true to run")
	}

	suite.Run(t, new(PostgresRepositoryTestSuite))
}
