package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielino/comio/internal/config"
	"github.com/danielino/comio/internal/monitoring"
)

const serverAddr = "http://localhost:8080"

var (
	cfgFile   string
	version   string
	buildTime string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "comio",
	Short: "ComIO - Community IO Storage",
	Long: `ComIO is a production-ready S3-compliant storage solution in Golang
featuring RESTful API, CLI management, storage replication, raw device handling,
and authentication.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(v, b string) error {
	version = v
	buildTime = b
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/comio/config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := monitoring.InitLogger(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output); err != nil {
		fmt.Println("Error initializing logger:", err)
		os.Exit(1)
	}
}
