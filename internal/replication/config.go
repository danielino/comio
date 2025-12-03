package replication

import "time"

type Config struct {
	Enabled       bool          `yaml:"enabled"`
	Mode          Mode          `yaml:"mode"` // async, sync
	RemoteURL     string        `yaml:"remote_url"`
	RemoteToken   string        `yaml:"remote_token"`
	BatchSize     int           `yaml:"batch_size"`
	BatchInterval time.Duration `yaml:"batch_interval"`
	RetryAttempts int           `yaml:"retry_attempts"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
}

type Mode string

const (
	ModeAsync Mode = "async"
	ModeSync  Mode = "sync"
)

func DefaultConfig() Config {
	return Config{
		Enabled:       false,
		Mode:          ModeAsync,
		BatchSize:     100,
		BatchInterval: 1 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    5 * time.Second,
	}
}
