package monitoring

import (
	"testing"

	"go.uber.org/zap"
)

func TestInitLogger_JSON(t *testing.T) {
	err := InitLogger("info", "json", "stdout")
	if err != nil {
		t.Errorf("InitLogger() error = %v", err)
	}
	if Log == nil {
		t.Error("Log is nil after InitLogger")
	}
}

func TestInitLogger_Text(t *testing.T) {
	err := InitLogger("debug", "text", "stderr")
	if err != nil {
		t.Errorf("InitLogger() error = %v", err)
	}
	if Log == nil {
		t.Error("Log is nil after InitLogger")
	}
}

func TestInitLogger_InvalidLevel(t *testing.T) {
	err := InitLogger("invalid", "json", "stdout")
	if err != nil {
		t.Errorf("InitLogger() error = %v", err)
	}
	if Log == nil {
		t.Error("Log is nil after InitLogger")
	}
}

func TestGetLogger(t *testing.T) {
	Log = nil
	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger() returned nil")
	}
}

func TestGetLogger_Initialized(t *testing.T) {
	InitLogger("info", "json", "stdout")
	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger() returned nil")
	}
	if logger != Log {
		t.Error("GetLogger() did not return the global Log instance")
	}
}

func TestSync(t *testing.T) {
	InitLogger("info", "json", "stdout")
	Sync()
}

func TestSync_NilLogger(t *testing.T) {
	Log = nil
	Sync()
}

func TestLogger_Usage(t *testing.T) {
	InitLogger("info", "json", "stdout")
	Log.Info("test message", zap.String("key", "value"))
	Sync()
}
