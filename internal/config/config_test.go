package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupEnvVars is a helper to securely set required environment variables for testing
// and returns a cleanup function to unset them after the test completes.
func setupEnvVars(t *testing.T) func() {
	vars := map[string]string{
		"APP_JWT_SECRET":           "test_jwt_secret",
		"APP_POSTGRES_DB":          "test_db",
		"APP_POSTGRES_USER":        "test_user",
		"APP_CLIENT_URL":           "https://test.com",
		"APP_YANDEX_CLIENT_SECRET": "test_yandex_secret",
		"APP_SMTP_PASSWORD":        "test_smtp_password",
	}

	for k, v := range vars {
		require.NoError(t, os.Setenv(k, v))
	}

	return func() {
		for k := range vars {
			require.NoError(t, os.Unsetenv(k))
		}
		require.NoError(t, os.Unsetenv("APP_HTTP_PORT"))
		require.NoError(t, os.Unsetenv("APP_DATABASE_URL"))
	}
}

func TestConfig_Load(t *testing.T) {
	t.Run("success load with defaults and env", func(t *testing.T) {
		cleanup := setupEnvVars(t)
		defer cleanup()

		err := os.Setenv("APP_HTTP_PORT", ":9999")
		require.NoError(t, err)

		cfg, err := Load()
		require.NoError(t, err)

		assert.Equal(t, "test_jwt_secret", cfg.Auth.Secret)
		assert.Equal(t, "test_yandex_secret", cfg.OAuth.Yandex.ClientSecret)
		assert.Equal(t, "test_smtp_password", cfg.SMTP.Password)
		assert.Equal(t, "https://test.com", cfg.Client.URL)

		assert.Equal(t, ":9999", cfg.HTTP.Port)

		assert.Equal(t, 10.0, cfg.RateLimit.Global.Limit)
		assert.Equal(t, 3*time.Minute, cfg.RateLimit.CleanupInterval)
		assert.Equal(t, 24*time.Hour, cfg.Auth.EmailVerificationTTL)
	})

	t.Run("fail when a required value is missing", func(t *testing.T) {
		cleanup := setupEnvVars(t)
		defer cleanup()

		// Remove one required secret
		err := os.Unsetenv("APP_YANDEX_CLIENT_SECRET")
		require.NoError(t, err)

		cfg, err := Load()
		assert.ErrorIs(t, err, ErrMissingEnvVar)
		assert.ErrorContains(t, err, "Yandex Client Secret")
		assert.Nil(t, cfg)
	})

	t.Run("postgres validation logic - connection url", func(t *testing.T) {
		cleanup := setupEnvVars(t)
		defer cleanup()

		err := os.Setenv("APP_DATABASE_URL", "postgres://user:pass@localhost:5432/db")
		require.NoError(t, err)

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.Postgres.ConnectionURL)
	})

	t.Run("postgres validation logic - fail when no host/db and no url", func(t *testing.T) {
		cleanup := setupEnvVars(t)
		defer cleanup()

		cfg := &Config{
			Postgres: Postgres{
				ConnectionURL: "",
				Host:          "",
				Database:      "",
			},
		}

		err := validatePostgres(cfg)
		assert.ErrorContains(t, err, "postgres: need either connection_url or host+database+user+password")
	})
}
