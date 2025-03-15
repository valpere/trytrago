package repository_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/database/repository"
	"github.com/valpere/trytrago/domain/logging"
)

// BaseRepositoryTestSuite defines common testing functionality for different repository implementations
type BaseRepositoryTestSuite struct {
	suite.Suite
	repo   repository.Repository
	ctx    context.Context
	logger logging.Logger
	dbPath string
}

// SetupSuite initializes test database connection
func (s *BaseRepositoryTestSuite) SetupSuite() {
	// Create logger
	opts := logging.NewDefaultOptions()
	opts.Level = logging.DebugLevel
	logger, err := logging.NewLogger(opts)
	require.NoError(s.T(), err, "Failed to create logger")
	s.logger = logger

	// Set test context
	s.ctx = context.Background()
}

// TearDownSuite closes database connection
func (s *BaseRepositoryTestSuite) TearDownSuite() {
	if s.repo != nil {
		err := s.repo.Close()
		if err != nil {
			s.T().Logf("Error closing repository: %v", err)
		}
	}

	// Clean up SQLite file if used
	if s.dbPath != "" {
		err := os.Remove(s.dbPath)
		if err != nil && !os.IsNotExist(err) {
			s.T().Logf("Failed to remove test database file: %v", err)
		}
	}
}

// InitSchema sets up the database schema using auto-migration
func (s *BaseRepositoryTestSuite) InitSchema() {
	db, err := s.repo.GetDB()
	require.NoError(s.T(), err, "Failed to get database connection")

	// Auto-migrate tables
	err = db.AutoMigrate(
		&database.Entry{},
		&database.Meaning{},
		&database.Example{}, 
		&database.Translation{},
		&database.ChangeHistory{},
	)
	require.NoError(s.T(), err, "Failed to migrate tables")
}

// CreateTestEntry creates a test entry with the given parameters
func (s *BaseRepositoryTestSuite) CreateTestEntry(word string, entryType database.EntryType) *database.Entry {
	entry := &database.Entry{
		ID:            s.NewUUID(),
		Word:          word,
		Type:          entryType,
		Pronunciation: fmt.Sprintf("%s-pronunciation", word),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	err := s.repo.CreateEntry(s.ctx, entry)
	require.NoError(s.T(), err, "Failed to create test entry")

	return entry
}

// CreateTestEntryWithMeanings creates a test entry with meanings
func (s *BaseRepositoryTestSuite) CreateTestEntryWithMeanings(word string, entryType database.EntryType, meaningCount int) *database.Entry {
	entry := &database.Entry{
		ID:            s.NewUUID(),
		Word:          word,
		Type:          entryType,
		Pronunciation: fmt.Sprintf("%s-pronunciation", word),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		Meanings:      make([]database.Meaning, meaningCount),
	}

	// Create meanings
	for i := 0; i < meaningCount; i++ {
		meaningID := s.NewUUID()
		entry.Meanings[i] = database.Meaning{
			ID:             meaningID,
			EntryID:        entry.ID,
			Description:    fmt.Sprintf("Meaning %d for %s", i+1, word),
			PartOfSpeechId: s.NewUUID(),
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
	}

	err := s.repo.CreateEntry(s.ctx, entry)
	require.NoError(s.T(), err, "Failed to create test entry with meanings")

	return entry
}

// AddTranslationToMeaning adds a translation to the specified meaning
func (s *BaseRepositoryTestSuite) AddTranslationToMeaning(entry *database.Entry, meaningIndex int, languageID, text string) *database.Translation {
	if meaningIndex >= len(entry.Meanings) {
		s.T().Fatalf("Invalid meaning index: %d", meaningIndex)
	}

	// Create translation
	translation := database.Translation{
		ID:         s.NewUUID(),
		MeaningID:  entry.Meanings[meaningIndex].ID,
		LanguageID: languageID,
		Text:       text,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// Add translation to the meaning
	entry.Meanings[meaningIndex].Translations = append(entry.Meanings[meaningIndex].Translations, translation)

	// Update entry
	err := s.repo.UpdateEntry(s.ctx, entry)
	require.NoError(s.T(), err, "Failed to add translation to meaning")

	return &translation
}

// NewUUID is a helper that creates a new UUID
func (s *BaseRepositoryTestSuite) NewUUID() uuid.UUID {
	return uuid.New()
}

// CreateSQLiteTestDB creates a temporary SQLite database for testing
func CreateSQLiteTestDB() (string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "trytrago-test-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create database path
	dbPath := filepath.Join(tempDir, "trytrago_test.db")
	return dbPath, nil
}

// SkipIfNotIntegrationTest skips the test if not running in integration mode
func SkipIfNotIntegrationTest(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration tests. Set INTEGRATION_TEST=true to run")
	}
}
