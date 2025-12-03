package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// bucketCmd represents the bucket command
var bucketCmd = &cobra.Command{
	Use:   "bucket",
	Short: "Bucket management commands",
}

var bucketCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new bucket",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Creating bucket %s\n", args[0])
		// Call service
	},
}

func init() {
	rootCmd.AddCommand(bucketCmd)
	bucketCmd.AddCommand(bucketCreateCmd)
}
