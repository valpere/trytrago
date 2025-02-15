// domain/config.go
package domain

import (
    "time"
)

// Config represents the complete application configuration
type Config struct {
    // Basic application settings
    Environment string `mapstructure:"environment"`
    Verbose     bool   `mapstructure:"verbose"`

    // Server configuration - handles both REST and gRPC endpoints
    Server struct {
        HTTPPort int `mapstructure:"http_port"`
        GRPCPort int `mapstructure:"grpc_port"`
        
        // Rate limiting settings to protect the server
        RateLimit struct {
            RequestsPerSecond int `mapstructure:"requests_per_second"`
            BurstSize        int `mapstructure:"burst_size"`
        } `mapstructure:"rate_limit"`
    } `mapstructure:"server"`

    // Database configuration - supports multiple database types
    Database struct {
        Type          string        `mapstructure:"type"`
        Host          string        `mapstructure:"host"`
        Port          int           `mapstructure:"port"`
        Name          string        `mapstructure:"name"`
        User          string        `mapstructure:"user"`
        Password      string        `mapstructure:"password"`
        
        // Connection pool settings - crucial for handling 60M entries
        PoolSize      int           `mapstructure:"pool_size"`
        MaxIdleConns  int           `mapstructure:"max_idle_conns"`
        MaxOpenConns  int           `mapstructure:"max_open_conns"`
        ConnTimeout   time.Duration `mapstructure:"conn_timeout"`
    } `mapstructure:"database"`

    // Logging configuration
    Logging struct {
        Level    string `mapstructure:"level"`
        Format   string `mapstructure:"format"`
        FilePath string `mapstructure:"file_path"`
    } `mapstructure:"logging"`

    // Authentication settings
    Auth struct {
        JWTSecret       string        `mapstructure:"jwt_secret"`
        TokenExpiration time.Duration `mapstructure:"token_expiration"`
    } `mapstructure:"auth"`
}

// NewDefaultConfig returns a Config instance with sensible defaults
func NewDefaultConfig() *Config {
    cfg := &Config{}
    
    // Set default values optimized for a dictionary with 60M entries
    cfg.Environment = "development"
    cfg.Verbose = false
    
    cfg.Server.HTTPPort = 8080
    cfg.Server.GRPCPort = 9090
    cfg.Server.RateLimit.RequestsPerSecond = 1000  // Adjusted for high-load
    cfg.Server.RateLimit.BurstSize = 100
    
    cfg.Database.Type = "postgres"  // Default to PostgreSQL for large datasets
    cfg.Database.Host = "localhost"
    cfg.Database.Port = 5432
    cfg.Database.PoolSize = 20      // Higher pool size for better concurrency
    cfg.Database.MaxIdleConns = 10
    cfg.Database.MaxOpenConns = 100 // Higher limit for large-scale operations
    cfg.Database.ConnTimeout = 30 * time.Second
    
    cfg.Logging.Level = "info"
    cfg.Logging.Format = "json"     // JSON format for better log processing
    
    cfg.Auth.TokenExpiration = 24 * time.Hour
    
    return cfg
}
