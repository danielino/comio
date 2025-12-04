package config

import "time"

// Config holds the global configuration
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Storage     StorageConfig     `mapstructure:"storage"`
	Replication ReplicationConfig `mapstructure:"replication"`
	Auth        AuthConfig        `mapstructure:"auth"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Metrics     MetricsConfig     `mapstructure:"metrics"`
	Lifecycle   LifecycleConfig   `mapstructure:"lifecycle"`
}

// ServerConfig holds server settings
type ServerConfig struct {
	Host            string    `mapstructure:"host"`
	Port            int       `mapstructure:"port"`
	ReadTimeout     string    `mapstructure:"read_timeout"`
	WriteTimeout    string    `mapstructure:"write_timeout"`
	ShutdownTimeoutStr string `mapstructure:"shutdown_timeout"`
	TLS             TLSConfig `mapstructure:"tls"`
}

// ShutdownTimeout returns the shutdown timeout duration
func (s *ServerConfig) ShutdownTimeout() time.Duration {
	if s.ShutdownTimeoutStr == "" {
		return 30 * time.Second // Default 30 seconds
	}
	d, err := time.ParseDuration(s.ShutdownTimeoutStr)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

// TLSConfig holds TLS settings
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// StorageConfig holds storage settings
type StorageConfig struct {
	Devices           []DeviceConfig `mapstructure:"devices"`
	BlockSize         int            `mapstructure:"block_size"`
	ReplicationFactor int            `mapstructure:"replication_factor"`
}

// DeviceConfig holds device settings
type DeviceConfig struct {
	Path string `mapstructure:"path"`
	Type string `mapstructure:"type"`
}

// ReplicationConfig holds replication settings
type ReplicationConfig struct {
	Nodes        []NodeConfig `mapstructure:"nodes"`
	WriteQuorum  int          `mapstructure:"write_quorum"`
	ReadQuorum   int          `mapstructure:"read_quorum"`
	SyncInterval string       `mapstructure:"sync_interval"`
}

// NodeConfig holds node settings
type NodeConfig struct {
	Address string `mapstructure:"address"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	AdminAccessKey string `mapstructure:"admin_access_key"`
	AdminSecretKey string `mapstructure:"admin_secret_key"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// MetricsConfig holds metrics settings
type MetricsConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

// LifecycleConfig holds lifecycle settings
type LifecycleConfig struct {
	EvaluationInterval string `mapstructure:"evaluation_interval"`
}
