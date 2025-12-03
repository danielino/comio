package monitoring

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger initializes the global logger
func InitLogger(level, format, output string) error {
	var config zap.Config

	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Set log level
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zap.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	// Set output
	if output == "stdout" {
		config.OutputPaths = []string{"stdout"}
	} else if output == "stderr" {
		config.OutputPaths = []string{"stderr"}
	} else {
		config.OutputPaths = []string{output}
	}

	// Build logger
	logger, err := config.Build()
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if Log == nil {
		// Fallback to a basic logger if not initialized
		l, _ := zap.NewProduction()
		return l
	}
	return Log
}
