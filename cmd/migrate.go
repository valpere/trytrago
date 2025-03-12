package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valpere/trytrago/domain"
	"github.com/valpere/trytrago/domain/database/repository"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/infrastructure/migration"
	"github.com/valpere/trytrago/migrations"
)

var (
	migrationPath    string
	migrateAutoApply bool
	migrateRollback  bool
	migrateVersion   int64
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database migrations",
	Long:  `Apply, rollback, or check the status of database migrations`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigrate(migrateAutoApply, migrateRollback, migrateVersion)
	},
}

func init() {
	migrateCmd.Flags().StringVar(&migrationPath, "path", "migrations", "Path to migration files")
	migrateCmd.Flags().BoolVar(&migrateAutoApply, "apply", false, "Automatically apply pending migrations")
	migrateCmd.Flags().BoolVar(&migrateRollback, "rollback", false, "Rollback the last migration or a specific version")
	migrateCmd.Flags().Int64Var(&migrateVersion, "version", 0, "Specific migration version to rollback (0 for last applied)")

	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(autoApply, rollback bool, version int64) error {
	log.Info("initializing database migration")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Initialize repository
	opts := repository.Options{
		Driver:   "postgres", // Default to PostgreSQL, could be configured
		Host:     "localhost",
		Port:     5432,
		Database: "trytrago",
		Username: "postgres",
		Password: "postgres",
		SSLMode:  "disable",
		Debug:    verbose,
	}

	// Use viper configuration if available
	if viper.IsSet("database.host") {
		opts.Host = viper.GetString("database.host")
	}
	if viper.IsSet("database.port") {
		opts.Port = viper.GetInt("database.port")
	}
	if viper.IsSet("database.name") {
		opts.Database = viper.GetString("database.name")
	}
	if viper.IsSet("database.user") {
		opts.Username = viper.GetString("database.user")
	}
	if viper.IsSet("database.password") {
		opts.Password = viper.GetString("database.password")
	}
	if viper.IsSet("database.type") {
		opts.Driver = viper.GetString("database.type")
	}

	// Create repository
	repo, err := domain.NewRepository(ctx, opts)
	if err != nil {
		log.Error("failed to create repository", logging.Error(err))
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Create migrator
	migrator := migration.NewMigrator(repo, log)

	// Check status or apply migrations based on flags
	if rollback {
		return handleRollback(ctx, migrator, version)
	}

	if autoApply {
		return migrations.RunMigrations(ctx, migrator, log)
	}

	return showMigrationStatus(ctx, migrator)
}

func showMigrationStatus(ctx context.Context, migrator *migration.Migrator) error {
	// Ensure migration table exists
	if err := migrator.EnsureMigrationTable(); err != nil {
		log.Error("failed to ensure migration table exists", logging.Error(err))
		return fmt.Errorf("failed to ensure migration table exists: %w", err)
	}

	// Get migration status
	status, err := migrator.Status(ctx, migrationPath)
	if err != nil {
		log.Error("failed to get migration status", logging.Error(err))
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	// Display migration status
	appliedCount := 0
	pendingCount := 0

	fmt.Println("Migration Status:")
	fmt.Println("================")
	for _, m := range status {
		if m["Applied"].(bool) {
			appliedCount++
			fmt.Printf("✅ V%d - %s (applied at %s)\n",
				m["Version"].(int64),
				m["Description"].(string),
				m["AppliedAt"].(time.Time).Format(time.RFC3339),
			)
		} else {
			pendingCount++
			fmt.Printf("❌ V%d - %s (pending)\n",
				m["Version"].(int64),
				m["Description"].(string),
			)
		}
	}

	fmt.Printf("\nSummary: %d applied, %d pending\n", appliedCount, pendingCount)

	if pendingCount > 0 {
		fmt.Println("\nTo apply pending migrations, run with --apply flag")
	}

	return nil
}

func handleRollback(ctx context.Context, migrator *migration.Migrator, version int64) error {
	if version > 0 {
		// Generate rollback script for specific version
		log.Info("generating rollback script", logging.Int64("version", version))
		script, err := migrations.GenerateRollbackScript(version)
		if err != nil {
			log.Error("failed to generate rollback script", logging.Error(err))
			return fmt.Errorf("failed to generate rollback script: %w", err)
		}

		// In a real implementation, you might want to execute this script
		// but for safety, we'll just print it
		fmt.Println("Rollback Script for Version", version, ":")
		fmt.Println("=====================================")
		fmt.Println(script)
		fmt.Println("=====================================")
		fmt.Println("To execute this rollback, use the appropriate database tool")

		return nil
	}

	// Rollback the last migration
	log.Info("rolling back the last migration")
	err := migrator.Rollback(ctx)
	if err != nil {
		log.Error("failed to rollback migration", logging.Error(err))
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	log.Info("successfully rolled back the last migration")
	return nil
}
