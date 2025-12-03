package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/danielino/comio/internal/api"
	"github.com/danielino/comio/internal/config"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the ComIO server",
	Long:  `Start the ComIO server with the configured settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config again or use global?
		// For now, assume config is loaded in root
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}
		
		server := api.NewServer(cfg)
		server.SetupRoutes()
		
		if err := server.Start(); err != nil {
			fmt.Println("Server error:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		// Same as serverCmd logic or serverCmd is the parent?
		// The prompt says "comio server start", so server is parent, start is child.
		// But serverCmd.Run above seems to start it.
		// Let's make serverCmd just a container and startCmd do the work.
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}
		
		server := api.NewServer(cfg)
		server.SetupRoutes()
		
		if err := server.Start(); err != nil {
			fmt.Println("Server error:", err)
		}
	},
}
