// cmd/version.go
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/valpere/trytrago/domain"
	"github.com/valpere/trytrago/domain/logging"
)

// Flags specific to the version command
var (
	versionOutputFormat string // Controls the version output format (text, json)
	short               bool   // Display just the version number
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long: `Display detailed version information about the TryTraGo Dictionary Server.
This includes the version number, build details, and runtime information.

Examples:
  # Display full version information in text format
  trytrago version

  # Display only the version number
  trytrago version --short

  # Display version information in JSON format
  trytrago version --format json`,
	RunE: runVersion,
}

func init() {
	// Add version command to root command
	rootCmd.AddCommand(versionCmd)

	// Define version-specific flags
	versionCmd.Flags().StringVar(&versionOutputFormat, "format", "text", "output format (text or json)")
	versionCmd.Flags().BoolVar(&short, "short", false, "print only the version number")
}

// runVersion implements the version command logic
func runVersion(cmd *cobra.Command, args []string) error {
	// Get version information
	versionInfo, err := domain.GetVersionInfo()
	if err != nil {
		log.Error("failed to get version information", logging.Error(err))
		return fmt.Errorf("failed to get version information: %w", err)
	}

	// Log that we're displaying version information
	log.Debug("displaying version information",
		logging.String("format", versionOutputFormat),
		logging.Bool("short", short),
	)

	// Handle short format request
	if short {
		fmt.Println(versionInfo.Version)
		return nil
	}

	// Handle different output formats
	switch versionOutputFormat {
	case "json":
		// Create a map for JSON output to control field names
		output := map[string]interface{}{
			"version":   versionInfo.Version,
			"commit":    versionInfo.CommitSHA,
			"buildTime": versionInfo.BuildTime,
			"goVersion": versionInfo.GoVersion,
			"os":        versionInfo.GOOS,
			"arch":      versionInfo.GOARCH,
		}

		// Marshal to JSON with indentation for readability
		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			log.Error("failed to marshal version information to JSON",
				logging.Error(err),
			)
			return fmt.Errorf("failed to generate JSON output: %w", err)
		}

		fmt.Println(string(jsonBytes))

	case "text":
		// Use the String() method for text output
		fmt.Println(versionInfo.String())

	default:
		log.Error("invalid output format",
			logging.String("format", versionOutputFormat),
		)
		return fmt.Errorf("invalid output format: %s", versionOutputFormat)
	}

	return nil
}
