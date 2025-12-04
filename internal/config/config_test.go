package config

import (
	"testing"
)

func TestConfig_Structs(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
			TLS: TLSConfig{
				Enabled:  true,
				CertFile: "/path/to/cert",
				KeyFile:  "/path/to/key",
			},
		},
		Storage: StorageConfig{
			Devices: []DeviceConfig{
				{Path: "/dev/sda", Type: "block"},
			},
			BlockSize:         4096,
			ReplicationFactor: 3,
		},
		Replication: ReplicationConfig{
			Nodes: []NodeConfig{
				{Address: "node1:8080"},
			},
			WriteQuorum: 2,
			ReadQuorum:  1,
		},
		Auth: AuthConfig{
			Enabled:        true,
			AdminAccessKey: "admin",
			AdminSecretKey: "secret",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Metrics: MetricsConfig{
			Enabled:  true,
			Endpoint: "/metrics",
		},
		Lifecycle: LifecycleConfig{
			EvaluationInterval: "1h",
		},
	}

	if cfg.Server.Host != "localhost" {
		t.Errorf("Server.Host = %s, want localhost", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if !cfg.Server.TLS.Enabled {
		t.Error("TLS.Enabled should be true")
	}
	if cfg.Storage.BlockSize != 4096 {
		t.Errorf("Storage.BlockSize = %d, want 4096", cfg.Storage.BlockSize)
	}
	if cfg.Replication.WriteQuorum != 2 {
		t.Errorf("Replication.WriteQuorum = %d, want 2", cfg.Replication.WriteQuorum)
	}
	if !cfg.Auth.Enabled {
		t.Error("Auth.Enabled should be true")
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %s, want info", cfg.Logging.Level)
	}
	if !cfg.Metrics.Enabled {
		t.Error("Metrics.Enabled should be true")
	}
}

func TestServerConfig(t *testing.T) {
	cfg := ServerConfig{
		Host:         "0.0.0.0",
		Port:         9000,
		ReadTimeout:  "30s",
		WriteTimeout: "30s",
		TLS: TLSConfig{
			Enabled:  false,
			CertFile: "",
			KeyFile:  "",
		},
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host = %s, want 0.0.0.0", cfg.Host)
	}
	if cfg.Port != 9000 {
		t.Errorf("Port = %d, want 9000", cfg.Port)
	}
	if cfg.TLS.Enabled {
		t.Error("TLS should be disabled")
	}
}

func TestStorageConfig(t *testing.T) {
	cfg := StorageConfig{
		Devices: []DeviceConfig{
			{Path: "/data1", Type: "file"},
			{Path: "/data2", Type: "file"},
		},
		BlockSize:         8192,
		ReplicationFactor: 2,
	}

	if len(cfg.Devices) != 2 {
		t.Errorf("Devices count = %d, want 2", len(cfg.Devices))
	}
	if cfg.BlockSize != 8192 {
		t.Errorf("BlockSize = %d, want 8192", cfg.BlockSize)
	}
	if cfg.ReplicationFactor != 2 {
		t.Errorf("ReplicationFactor = %d, want 2", cfg.ReplicationFactor)
	}
}

func TestReplicationConfig(t *testing.T) {
	cfg := ReplicationConfig{
		Nodes: []NodeConfig{
			{Address: "192.168.1.1:8080"},
			{Address: "192.168.1.2:8080"},
			{Address: "192.168.1.3:8080"},
		},
		WriteQuorum:  2,
		ReadQuorum:   1,
		SyncInterval: "5m",
	}

	if len(cfg.Nodes) != 3 {
		t.Errorf("Nodes count = %d, want 3", len(cfg.Nodes))
	}
	if cfg.WriteQuorum != 2 {
		t.Errorf("WriteQuorum = %d, want 2", cfg.WriteQuorum)
	}
	if cfg.SyncInterval != "5m" {
		t.Errorf("SyncInterval = %s, want 5m", cfg.SyncInterval)
	}
}

func TestAuthConfig(t *testing.T) {
	cfg := AuthConfig{
		Enabled:        false,
		AdminAccessKey: "",
		AdminSecretKey: "",
	}

	if cfg.Enabled {
		t.Error("Auth should be disabled")
	}
}

func TestLoggingConfig(t *testing.T) {
	cfg := LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "file",
	}

	if cfg.Level != "debug" {
		t.Errorf("Level = %s, want debug", cfg.Level)
	}
	if cfg.Format != "text" {
		t.Errorf("Format = %s, want text", cfg.Format)
	}
}

func TestMetricsConfig(t *testing.T) {
	cfg := MetricsConfig{
		Enabled:  true,
		Endpoint: "/prometheus",
	}

	if !cfg.Enabled {
		t.Error("Metrics should be enabled")
	}
	if cfg.Endpoint != "/prometheus" {
		t.Errorf("Endpoint = %s, want /prometheus", cfg.Endpoint)
	}
}

func TestLifecycleConfig(t *testing.T) {
	cfg := LifecycleConfig{
		EvaluationInterval: "24h",
	}

	if cfg.EvaluationInterval != "24h" {
		t.Errorf("EvaluationInterval = %s, want 24h", cfg.EvaluationInterval)
	}
}
