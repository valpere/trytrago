package repository

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/valpere/trytrago/domain/database"
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
