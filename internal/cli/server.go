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
		// Load configuration
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		// Wire up all dependencies using dependency injection
		container, err := api.NewServiceContainer(cfg)
		if err != nil {
			fmt.Println("Error initializing services:", err)
			return
		}

		// Create server with injected dependencies
		server := api.NewServer(cfg, container)
		server.SetupRoutes()

		// Start the server
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
		// Load configuration
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		// Wire up all dependencies using dependency injection
		container, err := api.NewServiceContainer(cfg)
		if err != nil {
			fmt.Println("Error initializing services:", err)
			return
		}

		// Create server with injected dependencies
		server := api.NewServer(cfg, container)
		server.SetupRoutes()

		// Start the server
		if err := server.Start(); err != nil {
			fmt.Println("Server error:", err)
		}
	},
}
