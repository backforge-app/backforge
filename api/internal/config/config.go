// Package config provides application configuration loading using viper.
// It supports YAML configuration file, environment variables and defaults.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var (
	ErrMissingEnvVar = errors.New("missing required environment variable")
)

// Config holds all application configuration parameters.
type Config struct {
	Env string `mapstructure:"env"` // the current environment (development, staging, production, etc.)

	HTTP     HTTP     `mapstructure:"http"`
	Telegram Telegram `mapstructure:"telegram"`
	Auth     Auth     `mapstructure:"auth"`
	Postgres Postgres `mapstructure:"postgres"`
	Client   Client   `mapstructure:"client"`
	Logging  Logging  `mapstructure:"logging"`
}

// HTTP contains settings for the HTTP server.
type HTTP struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// Telegram contains Telegram Bot related configuration.
type Telegram struct {
	Token string `mapstructure:"-"`
	Debug bool   `mapstructure:"debug"`
}

// Auth contains JWT authentication settings.
type Auth struct {
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
	JWTSecret       string        `mapstructure:"-"`
}

// Postgres contains PostgreSQL connection and pool settings.
type Postgres struct {
	ConnectionURL string     `mapstructure:"connection_url"`
	Host          string     `mapstructure:"host"`
	Port          int        `mapstructure:"port"`
	SSLMode       string     `mapstructure:"ssl_mode"`
	Pool          PoolConfig `mapstructure:"pool"`
	User          string     `mapstructure:"-"`
	Password      string     `mapstructure:"-"`
	Database      string     `mapstructure:"-"`
}

// PoolConfig contains PostgreSQL connection pool settings.
type PoolConfig struct {
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
}

// Client contains frontend application URLs.
type Client struct {
	URL string `mapstructure:"url"`
}

// Logging contains logging configuration.
type Logging struct {
	Level string `mapstructure:"level"`
}

// Load reads configuration from file, environment variables and defaults.
// It returns fully populated Config or an error.
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// Default values
	v.SetDefault("env", "development")

	// HTTP defaults
	v.SetDefault("http.port", ":8080")
	v.SetDefault("http.read_timeout", "10s")
	v.SetDefault("http.write_timeout", "10s")
	v.SetDefault("http.idle_timeout", "60s")
	v.SetDefault("http.shutdown_timeout", "8s")

	// Telegram defaults
	v.SetDefault("telegram.debug", false)

	// Auth defaults
	v.SetDefault("auth.access_token_ttl", "30m")
	v.SetDefault("auth.refresh_token_ttl", "168h")

	// Postgres defaults
	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.ssl_mode", "disable")
	v.SetDefault("postgres.pool.max_conns", 15)
	v.SetDefault("postgres.pool.min_conns", 2)
	v.SetDefault("postgres.pool.max_conn_lifetime", "30m")

	// Logging defaults
	v.SetDefault("logging.level", "info")

	// Enable environment variables support
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Try to read config file (not required)
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("cannot read config file: %w", err)
		}
		// file not found → continue with defaults + env
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config: %w", err)
	}

	// Load sensitive values only from environment
	cfg.Telegram.Token = v.GetString("telegram_bot_token")
	cfg.Telegram.Debug = v.GetBool("telegram.debug")

	cfg.Auth.JWTSecret = v.GetString("jwt_secret")

	cfg.Postgres.User = v.GetString("postgres_user")
	cfg.Postgres.Password = v.GetString("postgres_password")
	cfg.Postgres.Database = v.GetString("postgres_db")

	// Prefer DATABASE_URL if provided
	if dbURL := v.GetString("database_url"); dbURL != "" {
		cfg.Postgres.ConnectionURL = dbURL
	}

	// Override logging level if set via env
	cfg.Logging.Level = v.GetString("logging.level")

	// Required fields validation
	required := map[string]string{
		"Telegram Token":   cfg.Telegram.Token,
		"JWT Secret":       cfg.Auth.JWTSecret,
		"Postgres DB name": cfg.Postgres.Database,
		"Postgres User":    cfg.Postgres.User,
		"Client URL":       cfg.Client.URL,
	}

	for name, val := range required {
		if val == "" {
			return nil, fmt.Errorf("%w: %s (APP_TELEGRAM_BOT_TOKEN / APP_JWT_SECRET / ...)", ErrMissingEnvVar, name)
		}
	}

	// Postgres connection validation
	if cfg.Postgres.ConnectionURL == "" && (cfg.Postgres.Host == "" || cfg.Postgres.Database == "") {
		return nil, errors.New("postgres: need either connection_url or host+database+user+password")
	}

	return &cfg, nil
}
