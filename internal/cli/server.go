package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/danielino/comio/internal/api"
	"github.com/danielino/comio/internal/config"
)

// startServer contains the common server startup logic
func startServer() {
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

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	// Wait for shutdown signal
	<-quit
	fmt.Println("\nShutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout())
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		fmt.Println("Error during shutdown:", err)
	}
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the ComIO server",
	Long:  `Start the ComIO server with the configured settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
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
		startServer()
	},
}
