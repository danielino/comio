package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Config file settings
	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("/etc/comio/")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// Environment variable settings
	v.SetEnvPrefix("COMIO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is okay if we have defaults/env vars
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.tls.enabled", false)

	v.SetDefault("storage.block_size", 4096)
	v.SetDefault("storage.replication_factor", 3)

	v.SetDefault("replication.write_quorum", 2)
	v.SetDefault("replication.read_quorum", 1)
	v.SetDefault("replication.sync_interval", "5m")

	v.SetDefault("auth.enabled", true)

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")

	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.endpoint", "/admin/metrics")

	v.SetDefault("lifecycle.evaluation_interval", "24h")
}
