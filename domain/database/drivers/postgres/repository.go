package postgres

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/domain/database"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(ctx context.Context, opts database.Options) (database.Repository, error) {
	// Construct DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		opts.Host, opts.Port, opts.Username, opts.Password, opts.Database, opts.SSLMode,
	)

	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	if opts.Debug {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
	sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)

	// Create repository instance
	repo := &repository{db: db}

	// Verify connection
	if err := repo.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return repo, nil
}

// Implement Repository interface methods
func (r *repository) CreateEntry(ctx context.Context, entry *database.Entry) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *repository) GetEntryByID(ctx context.Context, id uuid.UUID) (*database.Entry, error) {
	var entry database.Entry
	err := r.db.WithContext(ctx).
		Preload("Meanings.Examples").
		Preload("Meanings.Translations").
		First(&entry, "id = ?", id).Error

	if err == gorm.ErrRecordNotFound {
		return nil, database.ErrEntryNotFound
	}
	return &entry, err
}

// ... implement other Repository interface methods ...
