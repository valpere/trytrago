package cmd

import "github.com/spf13/cobra"

var (
	backupPath string
	compress   bool
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup dictionary content",
	Long:  `Create a backup of the dictionary content in JSON format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBackup()
	},
}

func init() {
	backupCmd.Flags().StringVar(&backupPath, "output", "backup.json", "Output file path")
	backupCmd.Flags().BoolVar(&compress, "compress", false, "Compress backup file")

	rootCmd.AddCommand(backupCmd)
}

func runBackup() error {
	// TODO: Implement backup logic
	return nil
}
