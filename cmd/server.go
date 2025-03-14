package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain"
	"github.com/valpere/trytrago/domain/database/repository"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/infrastructure/auth"
	"github.com/valpere/trytrago/interface/server"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the TryTraGo server",
	Long:  `Start the TryTraGo multilanguage dictionary server with REST API support.`,
	RunE:  runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Add server-specific flags
	serverCmd.Flags().IntP("port", "p", 8080, "HTTP server port (overrides config)")
	serverCmd.Flags().StringP("db-type", "t", "", "Database type (postgres, mysql, sqlite)")
	serverCmd.Flags().StringP("log-level", "l", "", "Log level (debug, info, warn, error)")
}

func runServer(cmd *cobra.Command, args []string) error {
	// Initialize logger
	logger := logging.GetLogger()
	logger.Info("Starting TryTraGo server")

	// Load configuration
	config := loadConfiguration()

	// Override config with command line flags if provided
	if port, _ := cmd.Flags().GetInt("port"); port != 0 {
		config.Server.Port = port
	}
	if dbType, _ := cmd.Flags().GetString("db-type"); dbType != "" {
		config.Database.Type = dbType
	}
	if logLevel, _ := cmd.Flags().GetString("log-level"); logLevel != "" {
		config.Logging.Level = logLevel
	}

	// Initialize database
	repo, err := initializeRepository(config)
	if err != nil {
		logger.Error("Failed to initialize repository", logging.Error(err))
		return err
	}

	// Initialize JWT
	auth.InitJWT(config.Auth.JWTSecret, config.Auth.AccessTokenDuration)

	// Initialize services
	entryService := service.NewEntryService(repo, logger)
	translationService := service.NewTranslationService(repo, logger)
	userService := service.NewUserService(repo, logger)

	// Start server
	srv := server.NewServer(
		config,
		logger,
		entryService,
		translationService,
		userService,
	)

	// Set up graceful shutdown
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("Server error", logging.Error(err))
			shutdownCh <- os.Interrupt
		}
	}()

	logger.Info("TryTraGo server started successfully")

	// Wait for shutdown signal
	<-shutdownCh
	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", logging.Error(err))
		return err
	}

	logger.Info("Server gracefully stopped")
	return nil
}

// loadConfiguration loads the application configuration
func loadConfiguration() domain.Config {
	var config domain.Config

	// Set defaults
	config.Server.Port = 8080
	config.Server.Timeout = 30 * time.Second
	config.Server.ReadTimeout = 15 * time.Second
	config.Server.WriteTimeout = 15 * time.Second

	config.Database.Type = "postgres"
	config.Database.Host = "localhost"
	config.Database.Port = 5432
	config.Database.Name = "trytrago"
	config.Database.MaxOpenConns = 20
	config.Database.MaxIdleConns = 10
	config.Database.ConnLifetime = 5 * time.Minute

	config.Logging.Level = "info"
	config.Logging.Format = "console"

	config.Auth.JWTSecret = "your-secret-key-change-this-in-production"
	config.Auth.AccessTokenDuration = 1 * time.Hour
	config.Auth.RefreshTokenDuration = 7 * 24 * time.Hour

	// Read from viper if available
	if viper.IsSet("server.port") {
		config.Server.Port = viper.GetInt("server.port")
	}
	if viper.IsSet("server.timeout") {
		config.Server.Timeout = viper.GetDuration("server.timeout")
	}
	if viper.IsSet("server.read_timeout") {
		config.Server.ReadTimeout = viper.GetDuration("server.read_timeout")
	}
	if viper.IsSet("server.write_timeout") {
		config.Server.WriteTimeout = viper.GetDuration("server.write_timeout")
	}

	if viper.IsSet("database.type") {
		config.Database.Type = viper.GetString("database.type")
	}
	if viper.IsSet("database.host") {
		config.Database.Host = viper.GetString("database.host")
	}
	if viper.IsSet("database.port") {
		config.Database.Port = viper.GetInt("database.port")
	}
	if viper.IsSet("database.name") {
		config.Database.Name = viper.GetString("database.name")
	}
	if viper.IsSet("database.user") {
		config.Database.User = viper.GetString("database.user")
	}
	if viper.IsSet("database.password") {
		config.Database.Password = viper.GetString("database.password")
	}

	if viper.IsSet("logging.level") {
		config.Logging.Level = viper.GetString("logging.level")
	}
	if viper.IsSet("logging.format") {
		config.Logging.Format = viper.GetString("logging.format")
	}

	if viper.IsSet("auth.jwt_secret") {
		config.Auth.JWTSecret = viper.GetString("auth.jwt_secret")
	}
	if viper.IsSet("auth.access_token_duration") {
		config.Auth.AccessTokenDuration = viper.GetDuration("auth.access_token_duration")
	}
	if viper.IsSet("auth.refresh_token_duration") {
		config.Auth.RefreshTokenDuration = viper.GetDuration("auth.refresh_token_duration")
	}

	if viper.IsSet("environment") {
		config.Environment = viper.GetString("environment")
	} else {
		config.Environment = "development"
	}

	return config
}

// initializeRepository initializes the repository based on configuration
func initializeRepository(config domain.Config) (repository.Repository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := repository.Options{
		Driver:          config.Database.Type,
		Host:            config.Database.Host,
		Port:            config.Database.Port,
		Database:        config.Database.Name,
		Username:        config.Database.User,
		Password:        config.Database.Password,
		SSLMode:         "disable",
		MaxIdleConns:    config.Database.MaxIdleConns,
		MaxOpenConns:    config.Database.MaxOpenConns,
		ConnMaxLifetime: config.Database.ConnLifetime,
		Debug:           config.Logging.Level == "debug",
	}

	return domain.NewRepository(ctx, opts)
}
