package migration

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/valpere/trytrago/domain/database/repository"
	"github.com/valpere/trytrago/domain/logging"
	"gorm.io/gorm"
)

// Helper provides utility functions for database migrations
type Helper struct {
	repo   repository.Repository
	db     *gorm.DB
	logger logging.Logger
}

// NewHelper creates a new migration helper
func NewHelper(repo repository.Repository, logger logging.Logger) (*Helper, error) {
	db, err := repo.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	return &Helper{
		repo:   repo,
		db:     db,
		logger: logger.With(logging.String("component", "migration_helper")),
	}, nil
}

// EnsureMigrationsRun checks if migrations have been applied and runs them if necessary
func (h *Helper) EnsureMigrationsRun(ctx context.Context, migrationsDir string, autoApply bool) error {
	h.logger.Info("checking migration status")

	// Create migrator
	migrator := NewMigrator(h.repo, h.logger)

	// Ensure the migration table exists
	if err := migrator.EnsureMigrationTable(); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// Get migration status
	status, err := migrator.Status(ctx, migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	// Check if any migrations are pending
	pendingCount := 0
	for _, m := range status {
		if !m["Applied"].(bool) {
			pendingCount++
		}
	}

	// Report migration status
	h.logger.Info("migration status",
		logging.Int("total", len(status)),
		logging.Int("applied", len(status)-pendingCount),
		logging.Int("pending", pendingCount),
	)

	// Run migrations if needed and auto-apply is enabled
	if pendingCount > 0 {
		if autoApply {
			h.logger.Info("applying pending migrations", logging.Int("count", pendingCount))
			if err := migrator.Migrate(ctx, migrationsDir); err != nil {
				return fmt.Errorf("failed to apply migrations: %w", err)
			}
			h.logger.Info("migrations applied successfully")
		} else {
			h.logger.Warn("pending migrations detected but auto-apply is disabled",
				logging.Int("pending", pendingCount),
			)
		}
	}

	return nil
}

// PerformDatabaseOptimizations applies optimizations for better performance with large datasets
func (h *Helper) PerformDatabaseOptimizations(ctx context.Context) error {
	h.logger.Info("applying database optimizations")

	// Get raw SQL connection for operations that might not be supported by GORM
	sqlDB, err := h.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Set session-level parameters (these won't persist beyond the connection)
	optimizations := []struct {
		name  string
		value string
	}{
		{"work_mem", "20MB"},
		{"maintenance_work_mem", "256MB"},
		{"random_page_cost", "1.1"},
		{"effective_io_concurrency", "200"},
	}

	for _, opt := range optimizations {
		query := fmt.Sprintf("SET %s = '%s'", opt.name, opt.value)
		_, err := sqlDB.ExecContext(ctx, query)
		if err != nil {
			h.logger.Warn("failed to set optimization parameter",
				logging.String("parameter", opt.name),
				logging.String("value", opt.value),
				logging.Error(err),
			)
			// Continue with other optimizations
		}
	}

	// Check if materialized view exists and refresh it
	exists, err := h.materializedViewExists(ctx, sqlDB, "common_translations")
	if err != nil {
		h.logger.Warn("failed to check if materialized view exists", logging.Error(err))
	} else if exists {
		h.logger.Info("refreshing common_translations materialized view")
		_, err := sqlDB.ExecContext(ctx, "SELECT refresh_common_translations()")
		if err != nil {
			h.logger.Warn("failed to refresh materialized view", logging.Error(err))
		}
	}

	// Analyze tables for query planning
	h.logger.Info("analyzing database tables")
	_, err = sqlDB.ExecContext(ctx, "SELECT analyze_dictionary_tables()")
	if err != nil {
		// This might fail if the function doesn't exist yet
		h.logger.Warn("failed to analyze tables", logging.Error(err))

		// Fall back to manual analyze
		tables := []string{"entries", "meanings", "translations", "users", "comments", "likes"}
		for _, table := range tables {
			_, err := sqlDB.ExecContext(ctx, fmt.Sprintf("ANALYZE %s", table))
			if err != nil {
				h.logger.Warn("failed to analyze table",
					logging.String("table", table),
					logging.Error(err),
				)
			}
		}
	}

	h.logger.Info("database optimizations completed")
	return nil
}

// materializedViewExists checks if a materialized view exists in the database
func (h *Helper) materializedViewExists(ctx context.Context, db *sql.DB, viewName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM pg_catalog.pg_class c 
		JOIN pg_namespace n ON n.oid = c.relnamespace 
		WHERE c.relkind = 'm' 
		AND n.nspname = 'public' 
		AND c.relname = $1
	`

	var count int
	err := db.QueryRowContext(ctx, query, viewName).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetMigrationStatus returns a summary of the migration status
func (h *Helper) GetMigrationStatus(ctx context.Context, migrationsDir string) (map[string]interface{}, error) {
	// Create migrator
	migrator := NewMigrator(h.repo, h.logger)

	// Ensure the migration table exists
	if err := migrator.EnsureMigrationTable(); err != nil {
		return nil, fmt.Errorf("failed to create migration table: %w", err)
	}

	// Get migration status
	status, err := migrator.Status(ctx, migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}

	// Prepare summary
	appliedCount := 0
	pendingCount := 0
	lastApplied := ""
	lastAppliedTime := time.Time{}
	pendingMigrations := []map[string]interface{}{}

	for _, m := range status {
		if m["Applied"].(bool) {
			appliedCount++
			if lastAppliedTime.Before(m["AppliedAt"].(time.Time)) {
				lastAppliedTime = m["AppliedAt"].(time.Time)
				lastApplied = fmt.Sprintf("V%d - %s", m["Version"].(int64), m["Description"].(string))
			}
		} else {
			pendingCount++
			pendingMigrations = append(pendingMigrations, map[string]interface{}{
				"version":     m["Version"].(int64),
				"description": m["Description"].(string),
			})
		}
	}

	return map[string]interface{}{
		"total":              len(status),
		"applied":            appliedCount,
		"pending":            pendingCount,
		"last_applied":       lastApplied,
		"last_applied_at":    lastAppliedTime,
		"pending_migrations": pendingMigrations,
	}, nil
}
