package sqlite

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/database/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type dbrepo struct {
	db *gorm.DB
}

// NewRepository creates a new SQLite repository instance
func NewRepository(ctx context.Context, opts repository.Options) (repository.Repository, error) {
	// Construct database path
	dbPath := opts.Database
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join("data", dbPath)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	if opts.Debug {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	// Connect to database
	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// SQLite-specific settings:
	// - Max one writer at a time (SQLite limitation)
	// - WAL mode for better concurrency
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)

	// Enable WAL mode for better concurrency
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign key constraints
	if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create repository instance
	repo := &dbrepo{db: db}

	// Verify connection
	if err := repo.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite: %w", err)
	}

	return repo, nil
}

func (r *dbrepo) Ping(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (r *dbrepo) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (r *dbrepo) CreateEntry(ctx context.Context, entry *database.Entry) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	// Set creation timestamps
	now := time.Now().UTC()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	// Handle meanings and their related items
	for i := range entry.Meanings {
		if entry.Meanings[i].ID == uuid.Nil {
			entry.Meanings[i].ID = uuid.New()
		}
		entry.Meanings[i].EntryID = entry.ID
		entry.Meanings[i].CreatedAt = now
		entry.Meanings[i].UpdatedAt = now

		// Handle examples
		for j := range entry.Meanings[i].Examples {
			if entry.Meanings[i].Examples[j].ID == uuid.Nil {
				entry.Meanings[i].Examples[j].ID = uuid.New()
			}
			entry.Meanings[i].Examples[j].MeaningID = entry.Meanings[i].ID
			entry.Meanings[i].Examples[j].CreatedAt = now
			entry.Meanings[i].Examples[j].UpdatedAt = now
		}

		// Handle translations
		for j := range entry.Meanings[i].Translations {
			if entry.Meanings[i].Translations[j].ID == uuid.Nil {
				entry.Meanings[i].Translations[j].ID = uuid.New()
			}
			entry.Meanings[i].Translations[j].MeaningID = entry.Meanings[i].ID
			entry.Meanings[i].Translations[j].CreatedAt = now
			entry.Meanings[i].Translations[j].UpdatedAt = now
		}
	}

	// Use transaction to ensure all data is created atomically
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entry).Error; err != nil {
			if database.IsDuplicateError(err) {
				return database.ErrDuplicateEntry
			}
			return err
		}
		return nil
	})

	if err != nil {
		return database.NewDatabaseError(err, "create", "entries")
	}

	return nil
}

func (r *dbrepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*database.Entry, error) {
	var entry database.Entry

	// SQLite-optimized query - simpler preloading to avoid complex joins
	result := r.db.WithContext(ctx).
		Preload("Meanings").
		Preload("Meanings.Examples").
		Preload("Meanings.Translations").
		First(&entry, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, database.ErrEntryNotFound
		}
		return nil, database.NewDatabaseError(result.Error, "query", "entries")
	}

	return &entry, nil
}

func (r *dbrepo) UpdateEntry(ctx context.Context, entry *database.Entry) error {
	// Set update timestamp
	entry.UpdatedAt = time.Now().UTC()

	// SQLite has limited transaction capabilities compared to other databases
	// We need to use simpler approach to avoid "database is locked" errors
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First check if the entry exists
		var count int64
		if err := tx.Model(&database.Entry{}).Where("id = ?", entry.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return database.ErrEntryNotFound
		}

		// Update entry
		if err := tx.Save(entry).Error; err != nil {
			return err
		}

		// Handle meanings - more complex as we need to handle relationships
		for i := range entry.Meanings {
			// Ensure IDs are set
			if entry.Meanings[i].ID == uuid.Nil {
				entry.Meanings[i].ID = uuid.New()
				entry.Meanings[i].CreatedAt = entry.UpdatedAt
			}
			entry.Meanings[i].EntryID = entry.ID
			entry.Meanings[i].UpdatedAt = entry.UpdatedAt

			// Update the meaning
			if err := tx.Save(&entry.Meanings[i]).Error; err != nil {
				return err
			}

			// Update examples
			for j := range entry.Meanings[i].Examples {
				if entry.Meanings[i].Examples[j].ID == uuid.Nil {
					entry.Meanings[i].Examples[j].ID = uuid.New()
					entry.Meanings[i].Examples[j].CreatedAt = entry.UpdatedAt
				}
				entry.Meanings[i].Examples[j].MeaningID = entry.Meanings[i].ID
				entry.Meanings[i].Examples[j].UpdatedAt = entry.UpdatedAt

				if err := tx.Save(&entry.Meanings[i].Examples[j]).Error; err != nil {
					return err
				}
			}

			// Update translations
			for j := range entry.Meanings[i].Translations {
				if entry.Meanings[i].Translations[j].ID == uuid.Nil {
					entry.Meanings[i].Translations[j].ID = uuid.New()
					entry.Meanings[i].Translations[j].CreatedAt = entry.UpdatedAt
				}
				entry.Meanings[i].Translations[j].MeaningID = entry.Meanings[i].ID
				entry.Meanings[i].Translations[j].UpdatedAt = entry.UpdatedAt

				if err := tx.Save(&entry.Meanings[i].Translations[j]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return database.NewDatabaseError(err, "update", "entries")
	}

	return nil
}

func (r *dbrepo) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	// Use a transaction to delete the entry and all related records
	// SQLite requires a specific approach to avoid locking issues
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First check if the entry exists
		var count int64
		if err := tx.Model(&database.Entry{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return database.ErrEntryNotFound
		}

		// Find meanings to get their IDs for deleting examples and translations
		var meanings []database.Meaning
		if err := tx.Where("entry_id = ?", id).Find(&meanings).Error; err != nil {
			return err
		}

		// Delete translations for each meaning
		for _, meaning := range meanings {
			if err := tx.Where("meaning_id = ?", meaning.ID).Delete(&database.Translation{}).Error; err != nil {
				return err
			}

			// Delete examples for each meaning
			if err := tx.Where("meaning_id = ?", meaning.ID).Delete(&database.Example{}).Error; err != nil {
				return err
			}
		}

		// Delete all meanings
		if err := tx.Where("entry_id = ?", id).Delete(&database.Meaning{}).Error; err != nil {
			return err
		}

		// Finally delete the entry
		if err := tx.Delete(&database.Entry{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, database.ErrEntryNotFound) {
			return err
		}
		return database.NewDatabaseError(err, "delete", "entries")
	}

	return nil
}

func (r *dbrepo) ListEntries(ctx context.Context, params repository.ListParams) ([]database.Entry, error) {
	var entries []database.Entry
	query := r.db.WithContext(ctx)

	// Apply filters
	for key, value := range params.Filters {
		query = query.Where(key, value)
	}

	// Apply sorting
	if params.SortBy != "" {
		direction := "ASC"
		if params.SortDesc {
			direction = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", params.SortBy, direction))
	} else {
		// Default sorting by updated_at
		query = query.Order("updated_at DESC")
	}

	// Apply pagination with reasonable defaults
	limit := params.Limit
	if limit <= 0 {
		limit = 20 // Default limit
	}

	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	// SQLite performs better with smaller page sizes
	if limit > 100 {
		limit = 100
	}

	query = query.Limit(limit).Offset(offset)

	// Execute the query
	result := query.Find(&entries)
	if result.Error != nil {
		return nil, database.NewDatabaseError(result.Error, "list", "entries")
	}

	// If we need full entry data, preload related data
	// Due to SQLite's simpler query planner, we use separate queries for better performance
	if len(entries) > 0 {
		// For performance with large datasets, only preload for specific entries
		entryIDs := make([]uuid.UUID, len(entries))
		for i, entry := range entries {
			entryIDs[i] = entry.ID
		}

		// Load full data in a separate step
		// This is more efficient for SQLite than complex joins
		if err := r.db.WithContext(ctx).
			Preload("Meanings").
			Preload("Meanings.Examples").
			Preload("Meanings.Translations").
			Where("id IN ?", entryIDs).
			Find(&entries).Error; err != nil {
			return nil, database.NewDatabaseError(err, "list", "entries")
		}
	}

	return entries, nil
}

func (r *dbrepo) FindTranslations(ctx context.Context, word string, langID string) ([]database.Translation, error) {
	var translations []database.Translation

	// SQLite uses different case-insensitive function
	// Here we use the built-in SQLite case-insensitive comparison
	result := r.db.WithContext(ctx).
		Joins("JOIN meanings ON meanings.id = translations.meaning_id").
		Joins("JOIN entries ON entries.id = meanings.entry_id").
		Where("entries.word LIKE ? COLLATE NOCASE AND translations.language_id = ?", word, langID).
		Find(&translations)

	if result.Error != nil {
		return nil, database.NewDatabaseError(result.Error, "query", "translations")
	}

	return translations, nil
}

func (r *dbrepo) RecordChange(ctx context.Context, change *database.ChangeHistory) error {
	if change.ID == uuid.Nil {
		change.ID = uuid.New()
	}

	change.CreatedAt = time.Now().UTC()

	// SQLite has limitations with binary data - ensure it's handled properly
	result := r.db.WithContext(ctx).Create(change)
	if result.Error != nil {
		return database.NewDatabaseError(result.Error, "create", "change_history")
	}

	return nil
}

func (r *dbrepo) GetEntryHistory(ctx context.Context, entryID uuid.UUID) ([]database.ChangeHistory, error) {
	var history []database.ChangeHistory

	result := r.db.WithContext(ctx).
		Where("entry_id = ?", entryID).
		Order("created_at DESC").
		Find(&history)

	if result.Error != nil {
		return nil, database.NewDatabaseError(result.Error, "query", "change_history")
	}

	return history, nil
}

// WithTransaction is a helper for handling nested transactions
// With SQLite we need to be extra careful to avoid database locks
func (r *dbrepo) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return database.NewDatabaseError(tx.Error, "begin_transaction", "")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	// In SQLite, commit can sometimes fail due to locks
	// We add a retry mechanism with exponential backoff
	var commitErr error
	maxRetries := 3
	retryDelay := 10 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		commitErr = tx.Commit().Error
		if commitErr == nil {
			break
		}

		// If this isn't the last attempt, wait and retry
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	if commitErr != nil {
		return database.NewDatabaseError(commitErr, "commit_transaction", "")
	}

	return nil
}

// GetDB returns the underlying gorm.DB instance
func (r *dbrepo) GetDB() (*gorm.DB, error) {
	return r.db, nil
}
