// Package logger provides a centralized Zap logger for the application.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global sugared logger instance.
var Logger *zap.SugaredLogger

// Init initializes the logger based on the environment.
// Use "production" for JSON-structured logs; otherwise, development console logs.
// Returns an error if initialization fails.
func Init(env string) error {
	var cfg zap.Config

	if env == "production" {
		// Production config: JSON output, info level+, sampling for high volume.
		cfg = zap.NewProductionConfig()
		cfg.Encoding = "json"
		cfg.OutputPaths = []string{"stdout"}
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Development config: Human-readable console output, debug level+, stack traces.
		cfg = zap.NewDevelopmentConfig()
		cfg.Encoding = "console"
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapLogger, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}

	Logger = zapLogger.Sugar()
	return nil
}

// Sync flushes any buffered logs. Call this before app exit.
func Sync() {
	if Logger != nil {
		_ = Logger.Sync() //nolint:errcheck
	}
}
