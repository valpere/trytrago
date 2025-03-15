package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/model"
	"gorm.io/gorm"
)

// Repository defines the interface for database operations
type Repository interface {
	// Entry operations
	CreateEntry(ctx context.Context, entry *database.Entry) error
	GetEntryByID(ctx context.Context, id uuid.UUID) (*database.Entry, error)
	UpdateEntry(ctx context.Context, entry *database.Entry) error
	DeleteEntry(ctx context.Context, id uuid.UUID) error
	ListEntries(ctx context.Context, params ListParams) ([]database.Entry, error)

	// Translation operations
	FindTranslations(ctx context.Context, word string, langID string) ([]database.Translation, error)

	// History operations
	RecordChange(ctx context.Context, change *database.ChangeHistory) error
	GetEntryHistory(ctx context.Context, entryID uuid.UUID) ([]database.ChangeHistory, error)

	// User operations
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUserEntries(ctx context.Context, userID uuid.UUID, params ListParams) ([]database.Entry, error)

	// Social operations
	CreateComment(ctx context.Context, comment *model.Comment) error
	GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	ListComments(ctx context.Context, targetType string, targetID uuid.UUID) ([]model.Comment, error)
	DeleteComment(ctx context.Context, id uuid.UUID) error
	CreateLike(ctx context.Context, like *model.Like) error
	DeleteLike(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) error
	GetLike(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) (*model.Like, error)
	CountLikes(ctx context.Context, targetType string, targetID uuid.UUID) (int64, error)

	// Maintenance operations
	Ping(ctx context.Context) error
	Close() error

	// Access to the underlying database
	GetDB() (*gorm.DB, error)
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

// ListParams defines parameters for listing entries
type ListParams struct {
	Offset   int
	Limit    int
	SortBy   string
	SortDesc bool
	Filters  map[string]interface{}
}

// Options defines database connection options
type Options struct {
	Driver          string
	Host            string
	Port            int
	Database        string
	Username        string
	Password        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	SSLMode         string
	Debug           bool
}
