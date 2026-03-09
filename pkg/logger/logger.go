// Package logger provides a production-ready, reusable Zap logger.
//
// It supports structured logging, environment-specific configs, and optional
// contextual fields for request tracing.
package logger

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global sugared logger instance.
var Logger *zap.SugaredLogger
var once sync.Once

// Config defines the logger configuration.
type Config struct {
	Environment string // "production" or "development"
	Level       string // debug, info, warn, error
	OutputPaths []string
	ErrorOutput []string
}

// Init initializes the global logger based on the environment.
//
// Use "production" for JSON-structured logs, otherwise console human-readable logs.
// If called multiple times, initialization occurs only once.
func Init(cfg Config) error {
	var initErr error
	once.Do(func() {
		var zapCfg zap.Config

		switch cfg.Environment {
		case "production":
			zapCfg = zap.NewProductionConfig()
			zapCfg.Encoding = "json"
			if len(cfg.OutputPaths) > 0 {
				zapCfg.OutputPaths = cfg.OutputPaths
			} else {
				zapCfg.OutputPaths = []string{"stdout"}
			}
			if len(cfg.ErrorOutput) > 0 {
				zapCfg.ErrorOutputPaths = cfg.ErrorOutput
			} else {
				zapCfg.ErrorOutputPaths = []string{"stderr"}
			}
			zapCfg.EncoderConfig.TimeKey = "timestamp"
			zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		default:
			zapCfg = zap.NewDevelopmentConfig()
			zapCfg.Encoding = "console"
			zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			zapCfg.OutputPaths = []string{"stdout"}
			zapCfg.ErrorOutputPaths = []string{"stderr"}
		}

		if cfg.Level != "" {
			if lvl, err := zapcore.ParseLevel(cfg.Level); err == nil {
				zapCfg.Level = zap.NewAtomicLevelAt(lvl)
			} else {
				initErr = fmt.Errorf("invalid log level: %w", err)
				return
			}
		}

		zapLogger, err := zapCfg.Build()
		if err != nil {
			initErr = fmt.Errorf("build logger: %w", err)
			return
		}

		Logger = zapLogger.Sugar()
	})
	return initErr
}

// Sync flushes any buffered logs. Call before application exit.
func Sync() {
	if Logger != nil {
		_ = Logger.Sync() //nolint:errcheck
	}
}

// WithFields adds structured fields to the logger for context, returning a new SugaredLogger.
func WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	if Logger == nil {
		return zap.NewNop().Sugar()
	}
	return Logger.With(fields)
}

// WithField adds a single structured field.
func WithField(key string, value interface{}) *zap.SugaredLogger {
	return WithFields(map[string]interface{}{key: value})
}

// Example usage:
//  pkg/logger.Init(pkg/logger.Config{Environment: "production"})
//  pkg/logger.Logger.Info("service started")
//  pkg/logger.WithField("request_id", reqID).Info("processing request")
