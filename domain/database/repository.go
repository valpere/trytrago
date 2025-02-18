package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/domain/database/drivers/postgres"
)

// Repository defines the interface for database operations
type Repository interface {
	// Entry operations
	CreateEntry(ctx context.Context, entry *Entry) error
	GetEntryByID(ctx context.Context, id uuid.UUID) (*Entry, error)
	UpdateEntry(ctx context.Context, entry *Entry) error
	DeleteEntry(ctx context.Context, id uuid.UUID) error
	ListEntries(ctx context.Context, params ListParams) ([]Entry, error)

	// Translation operations
	FindTranslations(ctx context.Context, word string, langID string) ([]Translation, error)

	// History operations
	RecordChange(ctx context.Context, change *ChangeHistory) error
	GetEntryHistory(ctx context.Context, entryID uuid.UUID) ([]ChangeHistory, error)

	// Maintenance operations
	Ping(ctx context.Context) error
	Close() error
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

// NewRepository creates a new database repository based on the driver type
func NewRepository(ctx context.Context, opts Options) (Repository, error) {
	switch opts.Driver {
	case "postgres":
		return postgres.NewRepository(ctx, opts)
	// case "mysql":
	// 	return mysql.NewRepository(ctx, opts)
	// case "sqlite":
	// 	return sqlite.NewRepository(ctx, opts)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", opts.Driver)
	}
}
