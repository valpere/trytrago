package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/valpere/trytrago/domain"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version",
	Long:  `Display the version information of the dictionary server`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Dictionary Server v%s\n", domain.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
