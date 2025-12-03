package cli

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration commands",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
