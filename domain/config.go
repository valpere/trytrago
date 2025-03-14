package domain

import (
	"fmt"
	"time"
)

// Config represents the application configuration
type Config struct {
	// Server configuration
	Server struct {
		Port         int           `mapstructure:"port" yaml:"port"`
		Timeout      time.Duration `mapstructure:"timeout" yaml:"timeout"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
		TLS          struct {
			Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
			CertFile string `mapstructure:"cert_file" yaml:"cert_file"`
			KeyFile  string `mapstructure:"key_file" yaml:"key_file"`
		} `mapstructure:"tls" yaml:"tls"`
	} `mapstructure:"server" yaml:"server"`

	// Database configuration
	Database struct {
		Type          string        `mapstructure:"type" yaml:"type"`
		Host          string        `mapstructure:"host" yaml:"host"`
		Port          int           `mapstructure:"port" yaml:"port"`
		User          string        `mapstructure:"user" yaml:"user"`
		Password      string        `mapstructure:"password" yaml:"password"`
		Name          string        `mapstructure:"name" yaml:"name"`
		SSLMode       string        `mapstructure:"sslmode" yaml:"sslmode"`
		MaxOpenConns  int           `mapstructure:"max_open_conns" yaml:"max_open_conns"`
		MaxIdleConns  int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
		ConnLifetime  time.Duration `mapstructure:"conn_lifetime" yaml:"conn_lifetime"`
	} `mapstructure:"database" yaml:"database"`

	// Logging configuration
	Logging struct {
		Level      string `mapstructure:"level" yaml:"level"`
		Format     string `mapstructure:"format" yaml:"format"`
		Output     string `mapstructure:"output" yaml:"output"`
		FilePath   string `mapstructure:"file_path" yaml:"file_path"`
		EnableFile bool   `mapstructure:"enable_file" yaml:"enable_file"`
	} `mapstructure:"logging" yaml:"logging"`

	// Authentication configuration
	Auth struct {
		JWTSecret            string        `mapstructure:"jwt_secret" yaml:"jwt_secret"`
		AccessTokenDuration  time.Duration `mapstructure:"access_token_duration" yaml:"access_token_duration"`
		RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration" yaml:"refresh_token_duration"`
	} `mapstructure:"auth" yaml:"auth"`

	// Cache configuration
	Cache struct {
		Type     string        `mapstructure:"type" yaml:"type"`
		Address  string        `mapstructure:"address" yaml:"address"`
		Password string        `mapstructure:"password" yaml:"password"`
		DB       int           `mapstructure:"db" yaml:"db"`
		TTL      time.Duration `mapstructure:"ttl" yaml:"ttl"`
	} `mapstructure:"cache" yaml:"cache"`

	// Environment and version information
	Environment string `mapstructure:"environment" yaml:"environment"`
	Version     string `mapstructure:"version" yaml:"version"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server port must be positive")
	}

	if c.Database.Type == "" {
		return fmt.Errorf("database type must be specified")
	}

	// Validate logging configuration
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}

	if c.Logging.EnableFile && c.Logging.FilePath == "" {
		return fmt.Errorf("log file path must be specified when file logging is enabled")
	}

	// Validate authentication configuration
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret must be specified")
	}

	if c.Auth.AccessTokenDuration <= 0 {
		return fmt.Errorf("access token duration must be positive")
	}

	if c.Auth.RefreshTokenDuration <= 0 {
		return fmt.Errorf("refresh token duration must be positive")
	}

	return nil
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	switch c.Database.Type {
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Database.Host,
			c.Database.Port,
			c.Database.User,
			c.Database.Password,
			c.Database.Name,
			c.Database.SSLMode,
		)
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true",
			c.Database.User,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.Name,
		)
	case "sqlite":
		return c.Database.Name
	default:
		return ""
	}
}
