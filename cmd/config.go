package cmd

import "github.com/spf13/cobra"

var (
	// Global flags that apply across all commands
	// TODO: find out how it interacts with version
	outputFormatCfg string // Output format (yaml or json)
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display configuration",
	Long:  `Display the current configuration settings`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showConfig()
	},
}

func init() {
	configCmd.Flags().StringVar(&outputFormatCfg, "format", "yaml", "Output format (yaml or json)")

	rootCmd.AddCommand(configCmd)
}

func showConfig() error {
	// TODO: Implement show config logic
	return nil
}
