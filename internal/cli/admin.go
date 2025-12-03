package cli

import (
	"github.com/spf13/cobra"
)

// adminCmd represents the admin command
var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Administrative commands",
}

func init() {
	rootCmd.AddCommand(adminCmd)
}
