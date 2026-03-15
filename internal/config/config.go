// Package config provides application configuration loading and validation.
// It reads configuration from YAML files, environment variables, and defaults,
// and exposes a fully populated Config struct for use across the application.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ErrMissingEnvVar is returned when a required environment variable is not set.
var ErrMissingEnvVar = errors.New("missing required environment variable")

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
	Secret          string        `mapstructure:"-"`
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
	v := initViper()
	if err := readConfigFile(v); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config: %w", err)
	}

	loadSensitiveValues(v, &cfg)

	if err := validateRequired(&cfg); err != nil {
		return nil, err
	}

	if err := validatePostgres(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func initViper() *viper.Viper {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// Defaults
	v.SetDefault("env", "development")
	v.SetDefault("http.port", ":8080")
	v.SetDefault("http.read_timeout", "10s")
	v.SetDefault("http.write_timeout", "10s")
	v.SetDefault("http.idle_timeout", "60s")
	v.SetDefault("http.shutdown_timeout", "8s")

	v.SetDefault("telegram.debug", false)
	v.SetDefault("auth.access_token_ttl", "30m")
	v.SetDefault("auth.refresh_token_ttl", "168h")
	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.ssl_mode", "disable")
	v.SetDefault("postgres.pool.max_conns", 15)
	v.SetDefault("postgres.pool.min_conns", 2)
	v.SetDefault("postgres.pool.max_conn_lifetime", "30m")
	v.SetDefault("logging.level", "info")

	// Environment
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return v
}

func readConfigFile(v *viper.Viper) error {
	if err := v.ReadInConfig(); err != nil {
		var cfgNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &cfgNotFound) {
			return fmt.Errorf("cannot read config file: %w", err)
		}
		// not found → continue
	}
	return nil
}

func loadSensitiveValues(v *viper.Viper, cfg *Config) {
	cfg.Telegram.Token = v.GetString("telegram_bot_token")
	cfg.Telegram.Debug = v.GetBool("telegram.debug")
	cfg.Auth.Secret = v.GetString("jwt_secret")
	cfg.Postgres.User = v.GetString("postgres_user")
	cfg.Postgres.Password = v.GetString("postgres_password")
	cfg.Postgres.Database = v.GetString("postgres_db")
	cfg.Client.URL = v.GetString("client.url")

	if dbURL := v.GetString("database_url"); dbURL != "" {
		cfg.Postgres.ConnectionURL = dbURL
	}

	cfg.Logging.Level = v.GetString("logging.level")
}

func validateRequired(cfg *Config) error {
	required := map[string]string{
		"Telegram Token":   cfg.Telegram.Token,
		"Secret":           cfg.Auth.Secret,
		"Postgres DB name": cfg.Postgres.Database,
		"Postgres User":    cfg.Postgres.User,
		"Client URL":       cfg.Client.URL,
	}

	for name, val := range required {
		if val == "" {
			return fmt.Errorf("%w: %s (APP_TELEGRAM_BOT_TOKEN / APP_JWT_SECRET / ...)", ErrMissingEnvVar, name)
		}
	}

	return nil
}

func validatePostgres(cfg *Config) error {
	if cfg.Postgres.ConnectionURL == "" &&
		(cfg.Postgres.Host == "" || cfg.Postgres.Database == "") {
		return errors.New("postgres: need either connection_url or host+database+user+password")
	}
	return nil
}
