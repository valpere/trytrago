package main

import (
	"fmt"
	"os"

	"github.com/valpere/trytrago/cmd"
	"github.com/valpere/trytrago/domain/logging"
)

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

	// Execute root command
	cmd.Execute()
}
