package migration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/valpere/trytrago/domain/database"
	"github.com/valpere/trytrago/domain/database/repository"
	"github.com/valpere/trytrago/domain/logging"
	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	Version     int64
	Description string
	SQL         string
	Timestamp   time.Time
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	Version     int64 `gorm:"primaryKey"`
	Description string
	AppliedAt   time.Time
}

// Migrator handles database migrations
type Migrator struct {
	db     *gorm.DB
	repo   repository.Repository
	logger logging.Logger
}

// NewMigrator creates a new Migrator instance
func NewMigrator(repo repository.Repository, logger logging.Logger) *Migrator {
	db, _ := repo.GetDB()
	return &Migrator{
		db:     db,
		repo:   repo,
		logger: logger.With(logging.String("component", "migrator")),
	}
}

// EnsureMigrationTable creates the migrations table if it doesn't exist
func (m *Migrator) EnsureMigrationTable() error {
	if !m.db.Migrator().HasTable(&MigrationRecord{}) {
		err := m.db.Migrator().CreateTable(&MigrationRecord{})
		if err != nil {
			return fmt.Errorf("failed to create migrations table: %w", err)
		}
	}
	return nil
}

// LoadMigrationsFromDir loads migration files from a directory
func (m *Migrator) LoadMigrationsFromDir(dir string) ([]Migration, error) {
	// Ensure directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s", dir)
	}

	// Read migration files
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".sql") {
			continue
		}

		// Parse migration filename: V{version}_{description}.sql
		filename := file.Name()
		if !strings.HasPrefix(filename, "V") {
			continue
		}

		parts := strings.SplitN(strings.TrimSuffix(filename[1:], ".sql"), "_", 2)
		if len(parts) != 2 {
			m.logger.Warn("ignoring malformed migration filename", logging.String("filename", filename))
			continue
		}

		version, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			m.logger.Warn("ignoring migration with invalid version", 
				logging.String("filename", filename),
				logging.Error(err),
			)
			continue
		}

		description := strings.ReplaceAll(parts[1], "_", " ")

		// Read SQL content
		path := filepath.Join(dir, filename)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Create migration
		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			SQL:         string(content),
			Timestamp:   time.Now(),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// GetAppliedMigrations gets all migrations that have been applied
func (m *Migrator) GetAppliedMigrations() ([]MigrationRecord, error) {
	var records []MigrationRecord
	err := m.db.Order("version ASC").Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	return records, nil
}

// ApplyMigration applies a single migration
func (m *Migrator) ApplyMigration(ctx context.Context, migration Migration) error {
	m.logger.Info("applying migration", 
		logging.Int64("version", migration.Version),
		logging.String("description", migration.Description),
	)

	// Start a transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Execute migration SQL
	if err := tx.Exec(migration.SQL).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration as applied
	record := MigrationRecord{
		Version:     migration.Version,
		Description: migration.Description,
		AppliedAt:   time.Now().UTC(),
	}
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Migrate runs all pending migrations
func (m *Migrator) Migrate(ctx context.Context, dir string) error {
	// Ensure migrations table exists
	if err := m.EnsureMigrationTable(); err != nil {
		return err
	}

	// Load migrations
	migrations, err := m.LoadMigrationsFromDir(dir)
	if err != nil {
		return err
	}

	// Get applied migrations
	appliedMigrations, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Create a map of applied migrations for quick lookup
	appliedMap := make(map[int64]bool)
	for _, record := range appliedMigrations {
		appliedMap[record.Version] = true
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if appliedMap[migration.Version] {
			m.logger.Debug("skipping already applied migration",
				logging.Int64("version", migration.Version),
				logging.String("description", migration.Description),
			)
			continue
		}

		if err := m.ApplyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		m.logger.Info("migration applied successfully",
			logging.Int64("version", migration.Version),
			logging.String("description", migration.Description),
		)
	}

	return nil
}

// CreateInitialSchema creates the initial database schema
func (m *Migrator) CreateInitialSchema(ctx context.Context) error {
	m.logger.Info("creating initial schema")

	// Define tables to auto-migrate
	tables := []interface{}{
		&database.Entry{},
		&database.Meaning{},
		&database.Example{},
		&database.Translation{},
		&database.ChangeHistory{},
		&MigrationRecord{},
	}

	// Auto-migrate each table
	for _, table := range tables {
		if err := m.db.AutoMigrate(table); err != nil {
			return fmt.Errorf("failed to migrate table %T: %w", table, err)
		}
	}

	return nil
}

// Rollback rolls back the last migration
func (m *Migrator) Rollback(ctx context.Context) error {
	// Get the last applied migration
	var lastMigration MigrationRecord
	if err := m.db.Order("version DESC").First(&lastMigration).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("no migrations to roll back")
		}
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	m.logger.Info("rolling back migration",
		logging.Int64("version", lastMigration.Version),
		logging.String("description", lastMigration.Description),
	)

	// In a real implementation, we would need to store rollback SQL in migrations
	// For now, we'll just delete the migration record
	if err := m.db.Delete(&lastMigration).Error; err != nil {
		return fmt.Errorf("failed to delete migration record: %w", err)
	}

	m.logger.Info("migration rolled back successfully",
		logging.Int64("version", lastMigration.Version),
	)

	return nil
}

// Status returns the current migration status
func (m *Migrator) Status(ctx context.Context, dir string) ([]map[string]interface{}, error) {
	// Ensure migrations table exists
	if err := m.EnsureMigrationTable(); err != nil {
		return nil, err
	}

	// Load migrations
	migrations, err := m.LoadMigrationsFromDir(dir)
	if err != nil {
		return nil, err
	}

	// Get applied migrations
	appliedMigrations, err := m.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	// Create a map of applied migrations for quick lookup
	appliedMap := make(map[int64]MigrationRecord)
	for _, record := range appliedMigrations {
		appliedMap[record.Version] = record
	}

	// Create status report
	var status []map[string]interface{}

	for _, migration := range migrations {
		record, applied := appliedMap[migration.Version]

		info := map[string]interface{}{
			"Version":     migration.Version,
			"Description": migration.Description,
			"Applied":     applied,
		}

		if applied {
			info["AppliedAt"] = record.AppliedAt
		}

		status = append(status, info)
	}

	return status, nil
}
