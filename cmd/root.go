/*
Copyright (c) 2025 Valentyn Solomko <valentyn.solomko@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valpere/trytrago/domain"
	"github.com/valpere/trytrago/domain/logging"
)

var (
	// Global logger instance
	log logging.Logger
)

// Global flags that apply across all commands
var (
	cfgFile     string // Path to configuration file
	verbose     bool   // Enable verbose output
	logLevel    string // Logging level (debug, info, warn, error)
	logFormat   string // Log format (json or console)
	environment string // Runtime environment
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "trytrago",
	Short: "A multilanguage dictionary server",
	Long: `trytrago is a modern dictionary server providing both REST and gRPC APIs
	for managing and accessing multilanguage dictionary entries. It supports multiple
	database backends and can handle large-scale dictionaries efficiently.
	
	Complete documentation is available at https://github.com/valpere/trytrago`,
	Version: domain.Version,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func initLogging() {
	opts := logging.NewDefaultOptions()

	// Apply configuration from viper/flags
	opts.Level = logging.Level(viper.GetString("logging.level"))
	opts.Format = logging.Format(viper.GetString("logging.format"))
	opts.FilePath = viper.GetString("logging.file_path")
	opts.Environment = viper.GetString("environment")

	// Create the logger
	logger, err := logging.NewLogger(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	log = logger
}

func init() {
	// Initialize cobra and viper integration
	// Add initialization of logger after viper config is loaded
	cobra.OnInitialize(initConfig, initLogging)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.trytrago.yaml)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "logging level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "json", "log format (json or console)")
	rootCmd.PersistentFlags().StringVar(&environment, "env", "development", "environment (development, production)")

	// Bind persistent flags with viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("logging.level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("logging.format", rootCmd.PersistentFlags().Lookup("log-format"))
	viper.BindPFlag("environment", rootCmd.PersistentFlags().Lookup("env"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Look for config in the following locations:
		// 1. Current directory
		// 2. $HOME/.trytrago
		// 3. /etc/trytrago (for system-wide settings)
		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join(home, ".trytrago"))
		viper.AddConfigPath("/etc/trytrago")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Read environment variables prefixed with TRYTRAGO_
	viper.SetEnvPrefix("TRYTRAGO")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
