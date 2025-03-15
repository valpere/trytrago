package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/database/repository"
	"github.com/valpere/trytrago/domain/model"
)

type dbrepo struct {
	db *gorm.DB
}

func NewRepository(ctx context.Context, opts repository.Options) (repository.Repository, error) {
	// Construct DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		opts.Host, opts.Port, opts.Username, opts.Password, opts.Database, opts.SSLMode,
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
	repo := &dbrepo{db: db}

	// Verify connection
	if err := repo.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return repo, nil
}

// Implement Repository interface methods
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

	result := r.db.WithContext(ctx).
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

	// Update the entry in a transaction
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

	query = query.Limit(limit).Offset(offset)

	// Execute the query
	result := query.Find(&entries)
	if result.Error != nil {
		return nil, database.NewDatabaseError(result.Error, "list", "entries")
	}

	// If we need full entry data, preload related data
	if len(entries) > 0 {
		// For performance with large datasets, only preload for specific entries
		entryIDs := make([]uuid.UUID, len(entries))
		for i, entry := range entries {
			entryIDs[i] = entry.ID
		}

		// Fetch the complete data - PostgreSQL optimized query
		if err := r.db.WithContext(ctx).
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

	// Optimized query using joins - PostgreSQL specific with LOWER function
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
	if change.ID == uuid.Nil {
		change.ID = uuid.New()
	}

	change.CreatedAt = time.Now().UTC()

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

	if err := tx.Commit().Error; err != nil {
		return database.NewDatabaseError(err, "commit_transaction", "")
	}

	return nil
}

// GetDB returns the underlying gorm.DB instance
func (r *dbrepo) GetDB() (*gorm.DB, error) {
	return r.db, nil
}

// CountLikes counts the number of likes for a specific target
func (r *dbrepo) CountLikes(ctx context.Context, targetType string, targetID uuid.UUID) (int64, error) {
	var count int64

	// Use a SQL query to count likes
	query := `SELECT COUNT(*) FROM likes WHERE target_type = $1 AND target_id = $2 AND deleted_at IS NULL`

	err := r.db.WithContext(ctx).Raw(query, targetType, targetID).Scan(&count).Error
	if err != nil {
		return 0, database.NewDatabaseError(err, "count", "likes")
	}

	return count, nil
}

// CreateComment creates a new comment
func (r *dbrepo) CreateComment(ctx context.Context, comment *model.Comment) error {
	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}

	now := time.Now().UTC()
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = now
	}
	if comment.UpdatedAt.IsZero() {
		comment.UpdatedAt = now
	}

	result := r.db.WithContext(ctx).Create(comment)
	if result.Error != nil {
		return database.NewDatabaseError(result.Error, "create", "comments")
	}

	return nil
}

// User operations
func (r *dbrepo) CreateUser(ctx context.Context, user *model.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Set timestamps if not already set
	now := time.Now().UTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return database.NewDatabaseError(result.Error, "create", "users")
	}

	return nil
}

func (r *dbrepo) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User

	result := r.db.WithContext(ctx).First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, database.ErrNotFound
		}
		return nil, database.NewDatabaseError(result.Error, "query", "users")
	}

	return &user, nil
}

func (r *dbrepo) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User

	result := r.db.WithContext(ctx).First(&user, "username = ?", username)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, database.ErrNotFound
		}
		return nil, database.NewDatabaseError(result.Error, "query", "users")
	}

	return &user, nil
}

func (r *dbrepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User

	result := r.db.WithContext(ctx).First(&user, "email = ?", email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, database.ErrNotFound
		}
		return nil, database.NewDatabaseError(result.Error, "query", "users")
	}

	return &user, nil
}

func (r *dbrepo) UpdateUser(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now().UTC()

	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return database.NewDatabaseError(result.Error, "update", "users")
	}

	return nil
}

func (r *dbrepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id)
	if result.Error != nil {
		return database.NewDatabaseError(result.Error, "delete", "users")
	}

	if result.RowsAffected == 0 {
		return database.ErrNotFound
	}

	return nil
}

func (r *dbrepo) ListUserEntries(ctx context.Context, userID uuid.UUID, params repository.ListParams) ([]database.Entry, error) {
	var entries []database.Entry
	query := r.db.WithContext(ctx).Where("created_by_id = ?", userID)

	// Apply additional filters
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
		query = query.Order("updated_at DESC")
	}

	// Apply pagination
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}
	query = query.Limit(limit).Offset(offset)

	// Execute query
	result := query.Find(&entries)
	if result.Error != nil {
		return nil, database.NewDatabaseError(result.Error, "list", "entries")
	}

	return entries, nil
}

// Comment operations
func (r *dbrepo) GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	var comment model.Comment

	result := r.db.WithContext(ctx).First(&comment, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, database.ErrNotFound
		}
		return nil, database.NewDatabaseError(result.Error, "query", "comments")
	}

	return &comment, nil
}

func (r *dbrepo) ListComments(ctx context.Context, targetType string, targetID uuid.UUID) ([]model.Comment, error) {
	var comments []model.Comment

	result := r.db.WithContext(ctx).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("created_at DESC").
		Find(&comments)

	if result.Error != nil {
		return nil, database.NewDatabaseError(result.Error, "list", "comments")
	}

	return comments, nil
}

func (r *dbrepo) DeleteComment(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Comment{}, "id = ?", id)
	if result.Error != nil {
		return database.NewDatabaseError(result.Error, "delete", "comments")
	}

	if result.RowsAffected == 0 {
		return database.ErrNotFound
	}

	return nil
}

// Like operations
func (r *dbrepo) CreateLike(ctx context.Context, like *model.Like) error {
	if like.ID == uuid.Nil {
		like.ID = uuid.New()
	}

	if like.CreatedAt.IsZero() {
		like.CreatedAt = time.Now().UTC()
	}

	result := r.db.WithContext(ctx).Create(like)
	if result.Error != nil {
		return database.NewDatabaseError(result.Error, "create", "likes")
	}

	return nil
}

func (r *dbrepo) DeleteLike(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Delete(&model.Like{})

	if result.Error != nil {
		return database.NewDatabaseError(result.Error, "delete", "likes")
	}

	if result.RowsAffected == 0 {
		return database.ErrNotFound
	}

	return nil
}

func (r *dbrepo) GetLike(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) (*model.Like, error) {
	var like model.Like

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		First(&like)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, database.ErrNotFound
		}
		return nil, database.NewDatabaseError(result.Error, "query", "likes")
	}

	return &like, nil
}
