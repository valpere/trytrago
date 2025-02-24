package sqlite

import (
	"context"
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

// NewRepository creates a new SQLite dbrepo instance
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

	sqlDB.SetMaxOpenConns(1) // SQLite supports only one writer at a time
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

	// Create dbrepo instance
	repo := &dbrepo{db: db}

	// Verify connection
	if err := repo.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite: %w", err)
	}

	return repo, nil
}

// Repository interface methods implementation is the same as MySQL
// with minor adjustments for SQLite specifics

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

// Abstract methods
func (r *dbrepo) CreateEntry(ctx context.Context, entry *database.Entry) error {
	return nil
}

func (r *dbrepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*database.Entry, error) {
	return nil, nil
}

func (r *dbrepo) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (r *dbrepo) FindTranslations(ctx context.Context, word string, langID string) ([]database.Translation, error) {
	return nil, nil
}

func (r *dbrepo) GetEntryHistory(ctx context.Context, entryID uuid.UUID) ([]database.ChangeHistory, error) {
	return nil, nil
}

func (r *dbrepo) ListEntries(ctx context.Context, params repository.ListParams) ([]database.Entry, error) {
	return nil, nil
}

func (r *dbrepo) RecordChange(ctx context.Context, change *database.ChangeHistory) error {
	return nil
}

func (r *dbrepo) UpdateEntry(ctx context.Context, entry *database.Entry) error {
	return nil
}
