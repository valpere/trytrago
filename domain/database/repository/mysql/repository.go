package mysql

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/database/repository"
)

type dbrepo struct {
	db *gorm.DB
}

// NewRepository creates a new MySQL repository instance
func NewRepository(ctx context.Context, opts repository.Options) (repository.Repository, error) {
	// Construct MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		opts.Username,
		opts.Password,
		opts.Host,
		opts.Port,
		opts.Database,
	)

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
	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
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
	repo := &dbrepo{db: db}

	// Verify connection
	if err := repo.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	return repo, nil
}

// Implementation of Repository interface methods
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
	result := r.db.WithContext(ctx).Create(entry)
	if result.Error != nil {
		return database.NewDatabaseError(result.Error, "create", "entries")
	}
	return nil
}

func (r *dbrepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*database.Entry, error) {
	var entry database.Entry
	result := r.db.WithContext(ctx).
		Preload("Meanings.Examples").
		Preload("Meanings.Translations").
		First(&entry, "id = ?", id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, database.ErrEntryNotFound
		}
		return nil, database.NewDatabaseError(result.Error, "query", "entries")
	}

	return &entry, nil
}

func (r *dbrepo) UpdateEntry(ctx context.Context, entry *database.Entry) error {
	result := r.db.WithContext(ctx).Save(entry)
	if result.Error != nil {
		return database.NewDatabaseError(result.Error, "update", "entries")
	}
	return nil
}

func (r *dbrepo) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&database.Entry{}, "id = ?", id)
	if result.Error != nil {
		return database.NewDatabaseError(result.Error, "delete", "entries")
	}
	if result.RowsAffected == 0 {
		return database.ErrEntryNotFound
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
	}

	// Apply pagination
	query = query.Offset(params.Offset).Limit(params.Limit)

	result := query.Find(&entries)
	if result.Error != nil {
		return nil, database.NewDatabaseError(result.Error, "list", "entries")
	}

	return entries, nil
}

func (r *dbrepo) FindTranslations(ctx context.Context, word string, langID string) ([]database.Translation, error) {
	var translations []database.Translation
	result := r.db.WithContext(ctx).
		Joins("JOIN meanings ON meanings.id = translations.meaning_id").
		Joins("JOIN entries ON entries.id = meanings.entry_id").
		Where("LOWER(entries.word) = LOWER(?) AND translations.language_id = ?", word, langID).
		Find(&translations)

	if result.Error != nil {
		return nil, database.NewDatabaseError(result.Error, "query", "translations")
	}

	return translations, nil
}

func (r *dbrepo) RecordChange(ctx context.Context, change *database.ChangeHistory) error {
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
