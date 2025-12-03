package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// objectCmd represents the object command
var objectCmd = &cobra.Command{
	Use:   "object",
	Short: "Object management commands",
}

var objectPutCmd = &cobra.Command{
	Use:   "put <bucket> <key> <file>",
	Short: "Put an object",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Putting object %s/%s from %s\n", args[0], args[1], args[2])
		// Call service
	},
}

func init() {
	rootCmd.AddCommand(objectCmd)
	objectCmd.AddCommand(objectPutCmd)
}
