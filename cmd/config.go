package cmd

import "github.com/spf13/cobra"

var (
	outputFormat string
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
	configCmd.Flags().StringVar(&outputFormat, "format", "yaml", "Output format (yaml or json)")

	rootCmd.AddCommand(configCmd)
}

func showConfig() error {
	// TODO: Implement show config logic
	return nil
}
