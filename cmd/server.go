package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valpere/trytrago/domain/logging"
)

// Server-specific flags
var (
	httpPort       int           // HTTP server port
	grpcPort       int           // gRPC server port
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
	Long: `Start the trytrago dictionary server with both REST and gRPC APIs.
The server can be configured to use different database backends and supports
various configuration options for optimization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func init() {

	// Define server-specific flags
	serverCmd.Flags().IntVar(&httpPort, "http-port", 8080, "HTTP server port")
	serverCmd.Flags().IntVar(&grpcPort, "grpc-port", 9090, "gRPC server port")

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
	viper.BindPFlag("server.grpc_port", serverCmd.Flags().Lookup("grpc-port"))
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
		logging.Int("grpc_port", grpcPort),
		logging.String("environment", environment),
	)

	// // If an error occurs
	// if err := someOperation(); err != nil {
	// 	log.Error("failed to perform operation",
	// 		logging.Error(err),
	// 		logging.String("component", "server"),
	// 	)
	// 	return err
	// }

	// TODO: Implement server startup logic
	log.Info("Starting trytrago server")

	return nil
}
