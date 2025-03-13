package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain"
	"github.com/valpere/trytrago/domain/database/repository"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/infrastructure/migration"
	serverInterface "github.com/valpere/trytrago/interface/server"
)

// Server-specific flags
var (
	httpPort       int           // HTTP server port
	dbType         string        // Database type (postgres, mysql, sqlite)
	dbHost         string        // Database host
	dbPort         int           // Database port
	dbName         string        // Database name
	dbUser         string        // Database user
	dbPass         string        // Database password
	dbPoolSize     int           // Connection pool size
	dbMaxIdleConns int           // Maximum number of idle connections
	dbMaxOpenConns int           // Maximum number of open connections
	dbConnTimeout  time.Duration // Connection timeout
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the trytrago server",
	Long: `Start the trytrago dictionary server with REST API.
The server can be configured to use different database backends and supports
various configuration options for optimization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func init() {

	// Define server-specific flags
	serverCmd.Flags().IntVar(&httpPort, "http-port", 8080, "HTTP server port")

	// Database connection flags
	serverCmd.Flags().StringVar(&dbType, "db-type", "postgres", "Database type (postgres, mysql, sqlite)")
	serverCmd.Flags().StringVar(&dbHost, "db-host", "localhost", "Database host")
	serverCmd.Flags().IntVar(&dbPort, "db-port", 5432, "Database port")
	serverCmd.Flags().StringVar(&dbName, "db-name", "trytra", "Database name")
	serverCmd.Flags().StringVar(&dbUser, "db-user", "", "Database user")
	serverCmd.Flags().StringVar(&dbPass, "db-pass", "", "Database password")

	// Database optimization flags (important for 60M entries)
	serverCmd.Flags().IntVar(&dbPoolSize, "db-pool-size", 10, "Database connection pool size")
	serverCmd.Flags().IntVar(&dbMaxIdleConns, "db-max-idle-conns", 5, "Maximum idle database connections")
	serverCmd.Flags().IntVar(&dbMaxOpenConns, "db-max-open-conns", 50, "Maximum open database connections")
	serverCmd.Flags().DurationVar(&dbConnTimeout, "db-conn-timeout", 30*time.Second, "Database connection timeout")

	// Bind flags with viper
	viper.BindPFlag("server.http_port", serverCmd.Flags().Lookup("http-port"))
	viper.BindPFlag("database.type", serverCmd.Flags().Lookup("db-type"))
	viper.BindPFlag("database.host", serverCmd.Flags().Lookup("db-host"))
	viper.BindPFlag("database.port", serverCmd.Flags().Lookup("db-port"))
	viper.BindPFlag("database.name", serverCmd.Flags().Lookup("db-name"))
	viper.BindPFlag("database.user", serverCmd.Flags().Lookup("db-user"))
	viper.BindPFlag("database.password", serverCmd.Flags().Lookup("db-pass"))
	viper.BindPFlag("database.pool_size", serverCmd.Flags().Lookup("db-pool-size"))
	viper.BindPFlag("database.max_idle_conns", serverCmd.Flags().Lookup("db-max-idle-conns"))
	viper.BindPFlag("database.max_open_conns", serverCmd.Flags().Lookup("db-max-open-conns"))
	viper.BindPFlag("database.conn_timeout", serverCmd.Flags().Lookup("db-conn-timeout"))

	// Add server command to root command
	rootCmd.AddCommand(serverCmd)
}

// runServer implements the server command
func runServer() error {
	log.Info("starting server",
		logging.Int("http_port", httpPort),
		logging.String("environment", environment),
	)

	// Create context with timeout for initialization operations
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create configuration from loaded settings
	config := &domain.Config{}
	config.Environment = environment
	config.Verbose = verbose

	config.Server.HTTPPort = httpPort
	config.Server.RateLimit.RequestsPerSecond = viper.GetInt("server.rate_limit.requests_per_second")
	config.Server.RateLimit.BurstSize = viper.GetInt("server.rate_limit.burst_size")

	config.Database.Type = dbType
	config.Database.Host = dbHost
	config.Database.Port = dbPort
	config.Database.Name = dbName
	config.Database.User = dbUser
	config.Database.Password = dbPass
	config.Database.PoolSize = dbPoolSize
	config.Database.MaxIdleConns = dbMaxIdleConns
	config.Database.MaxOpenConns = dbMaxOpenConns
	config.Database.ConnTimeout = dbConnTimeout

	config.Logging.Level = logLevel
	config.Logging.Format = logFormat
	config.Logging.FilePath = viper.GetString("logging.file_path")

	config.Auth.JWTSecret = viper.GetString("auth.jwt_secret")
	config.Auth.TokenExpiration = viper.GetDuration("auth.token_expiration")

	// Set up database connection
	dbOpts := repository.Options{
		Driver:          dbType,
		Host:            dbHost,
		Port:            dbPort,
		Database:        dbName,
		Username:        dbUser,
		Password:        dbPass,
		MaxIdleConns:    dbMaxIdleConns,
		MaxOpenConns:    dbMaxOpenConns,
		ConnMaxLifetime: dbConnTimeout,
		Debug:           verbose,
	}

	// Create repository
	repo, err := domain.NewRepository(ctx, dbOpts)
	if err != nil {
		log.Error("failed to create repository",
			logging.Error(err),
			logging.String("component", "server"),
		)
		return err
	}

	// Run database migrations
	log.Info("checking database migrations")
	migrationHelper, err := migration.NewHelper(repo, log)
	if err != nil {
		log.Error("failed to create migration helper",
			logging.Error(err),
			logging.String("component", "server"),
		)
		return err
	}

	// Run migrations automatically in development, manually in production
	autoApply := environment != "production"
	if err := migrationHelper.EnsureMigrationsRun(ctx, "migrations", autoApply); err != nil {
		log.Error("failed to ensure migrations are applied",
			logging.Error(err),
			logging.String("component", "server"),
		)

		// In production, fail if migrations are not applied
		if environment == "production" {
			return err
		}

		// In development, log warning but continue
		log.Warn("continuing without migrations applied")
	}

	// Apply database optimizations
	if err := migrationHelper.PerformDatabaseOptimizations(ctx); err != nil {
		log.Warn("failed to apply database optimizations",
			logging.Error(err),
			logging.String("component", "server"),
		)
		// Continue despite optimization failures
	}

	// Initialize services
	entryService := service.NewEntryService(repo, log)
	translationService := service.NewTranslationService(repo, log)
	userService := service.NewUserService(repo, log)

	// Start server
	server := serverInterface.NewServer(
		config,
		log,
		entryService,
		translationService,
		userService,
	)

	log.Info("TryTraGo server started successfully")
	return server.Start()
}
