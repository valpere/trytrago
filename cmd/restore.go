package cmd

import "github.com/spf13/cobra"

var (
	restorePath string
	dryRun      bool
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore dictionary content",
	Long:  `Restore dictionary content from a JSON backup file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRestore()
	},
}

func init() {
	restoreCmd.Flags().StringVar(&restorePath, "input", "", "Input backup file path")
	restoreCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate backup without restoring")
	restoreCmd.MarkFlagRequired("input")

	rootCmd.AddCommand(restoreCmd)
}

func runRestore() error {
	// TODO: Implement restore logic
	return nil
}
