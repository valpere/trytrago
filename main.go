package main

import (
	"fmt"
	"os"

	"github.com/valpere/trytrago/cmd"
	"github.com/valpere/trytrago/domain/logging"
)

// main is the entry point for the application
func main() {
	// Initialize default logger options
	opts := logging.NewDefaultOptions()

	// Create logger and set as default
	logger, err := logging.NewLogger(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	// Set as default logger for the application
	logging.SetDefaultLogger(logger)

	// Log startup message
	logger.Info("Starting TryTraGo multilanguage dictionary server")

	// Execute root command
	cmd.Execute()
}
