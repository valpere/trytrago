package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/infrastructure/migration"
)

// RunMigrations executes all pending database migrations
func RunMigrations(ctx context.Context, migrator *migration.Migrator, logger logging.Logger) error {
	logger.Info("running database migrations")

	// Get the migration directory path
	migrationsDir := getMigrationsDirectory()
	logger.Debug("migrations directory", logging.String("path", migrationsDir))

	// Ensure the migration table exists
	if err := migrator.EnsureMigrationTable(); err != nil {
		logger.Error("failed to ensure migration table exists", logging.Error(err))
		return fmt.Errorf("failed to ensure migration table exists: %w", err)
	}

	// Run the migrations
	if err := migrator.Migrate(ctx, migrationsDir); err != nil {
		logger.Error("failed to run migrations", logging.Error(err))
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Print migration status
	status, err := migrator.Status(ctx, migrationsDir)
	if err != nil {
		logger.Error("failed to get migration status", logging.Error(err))
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	logger.Info("migration status", logging.Int("applied", countAppliedMigrations(status)))
	for _, m := range status {
		if m["Applied"].(bool) {
			logger.Info("migration applied",
				logging.Int64("version", m["Version"].(int64)),
				logging.String("description", m["Description"].(string)),
			)
		} else {
			logger.Warn("migration pending",
				logging.Int64("version", m["Version"].(int64)),
				logging.String("description", m["Description"].(string)),
			)
		}
	}

	logger.Info("migrations completed successfully")
	return nil
}

// getMigrationsDirectory returns the path to the migrations directory
func getMigrationsDirectory() string {
	// Check if running in development mode (current directory)
	if _, err := os.Stat("migrations"); err == nil {
		return "migrations"
	}

	// Try relative to executable
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		migrationPath := filepath.Join(exeDir, "migrations")
		if _, err := os.Stat(migrationPath); err == nil {
			return migrationPath
		}
	}

	// Default to /app/migrations (Docker environment)
	return "/app/migrations"
}

// countAppliedMigrations counts the number of applied migrations
func countAppliedMigrations(status []map[string]interface{}) int {
	count := 0
	for _, m := range status {
		if m["Applied"].(bool) {
			count++
		}
	}
	return count
}

// GenerateRollbackScript generates SQL to roll back a specific migration
func GenerateRollbackScript(version int64) (string, error) {
	migrationsDir := getMigrationsDirectory()
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Find the migration file for the given version
	var migrationFilename string
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".sql") {
			continue
		}

		if strings.HasPrefix(file.Name(), fmt.Sprintf("V%d__", version)) {
			migrationFilename = file.Name()
			break
		}
	}

	if migrationFilename == "" {
		return "", fmt.Errorf("migration file not found for version %d", version)
	}

	// Generate a rollback script based on the migration
	rollbackSQL, err := generateRollbackSQL(filepath.Join(migrationsDir, migrationFilename))
	if err != nil {
		return "", fmt.Errorf("failed to generate rollback SQL: %w", err)
	}

	return rollbackSQL, nil
}

// generateRollbackSQL analyzes a migration file and generates rollback SQL
func generateRollbackSQL(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read migration file: %w", err)
	}

	// Parse the SQL file to identify created tables, indices, etc.
	lines := strings.Split(string(content), "\n")
	var createdTables, createdIndices, createdViews, createdTriggers []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Extract table names from CREATE TABLE statements
		if strings.HasPrefix(strings.ToUpper(trimmedLine), "CREATE TABLE") {
			parts := strings.Split(trimmedLine, " ")
			for i, part := range parts {
				if strings.ToUpper(part) == "TABLE" && i+1 < len(parts) {
					tableName := strings.TrimSuffix(strings.TrimPrefix(parts[i+1], "IF NOT EXISTS "), "(")
					tableName = strings.TrimSpace(tableName)
					createdTables = append(createdTables, tableName)
					break
				}
			}
		}

		// Extract index names
		if strings.HasPrefix(strings.ToUpper(trimmedLine), "CREATE INDEX") {
			parts := strings.Split(trimmedLine, " ")
			for i, part := range parts {
				if strings.ToUpper(part) == "INDEX" && i+1 < len(parts) {
					indexName := parts[i+1]
					createdIndices = append(createdIndices, indexName)
					break
				}
			}
		}

		// Extract view names
		if strings.HasPrefix(strings.ToUpper(trimmedLine), "CREATE OR REPLACE VIEW") {
			parts := strings.Split(trimmedLine, " ")
			for i, part := range parts {
				if strings.ToUpper(part) == "VIEW" && i+1 < len(parts) {
					viewName := parts[i+1]
					createdViews = append(createdViews, viewName)
					break
				}
			}
		}

		// Extract trigger names
		if strings.HasPrefix(strings.ToUpper(trimmedLine), "CREATE TRIGGER") {
			parts := strings.Split(trimmedLine, " ")
			if len(parts) > 2 {
				triggerName := parts[2]
				createdTriggers = append(createdTriggers, triggerName)
			}
		}
	}

	// Generate rollback SQL
	var rollbackSQL strings.Builder

	rollbackSQL.WriteString("-- Auto-generated rollback script\n")
	rollbackSQL.WriteString("-- WARNING: This is a generated script and may not be complete\n")
	rollbackSQL.WriteString("-- Review and edit as necessary before execution\n\n")

	// Drop triggers first to avoid dependency issues
	for _, trigger := range createdTriggers {
		rollbackSQL.WriteString(fmt.Sprintf("DROP TRIGGER IF EXISTS %s;\n", trigger))
	}

	// Drop views
	for _, view := range createdViews {
		rollbackSQL.WriteString(fmt.Sprintf("DROP VIEW IF EXISTS %s;\n", view))
	}

	// Drop indices
	for _, index := range createdIndices {
		rollbackSQL.WriteString(fmt.Sprintf("DROP INDEX IF EXISTS %s;\n", index))
	}

	// Drop tables in reverse order to handle dependencies
	for i := len(createdTables) - 1; i >= 0; i-- {
		rollbackSQL.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;\n", createdTables[i]))
	}

	return rollbackSQL.String(), nil
}
