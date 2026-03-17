package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load(t *testing.T) {
	err := os.Setenv("APP_TELEGRAM_BOT_TOKEN", "test_token")
	require.NoError(t, err)
	err = os.Setenv("APP_JWT_SECRET", "test_secret")
	require.NoError(t, err)
	err = os.Setenv("APP_POSTGRES_DB", "test_db")
	require.NoError(t, err)
	err = os.Setenv("APP_POSTGRES_USER", "test_user")
	require.NoError(t, err)
	err = os.Setenv("APP_CLIENT_URL", "http://test.com")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Unsetenv("APP_TELEGRAM_BOT_TOKEN"))
		require.NoError(t, os.Unsetenv("APP_JWT_SECRET"))
		require.NoError(t, os.Unsetenv("APP_POSTGRES_DB"))
		require.NoError(t, os.Unsetenv("APP_POSTGRES_USER"))
		require.NoError(t, os.Unsetenv("APP_CLIENT_URL"))
		require.NoError(t, os.Unsetenv("APP_HTTP_PORT"))
	}()

	t.Run("success load with defaults and env", func(t *testing.T) {
		err := os.Setenv("APP_HTTP_PORT", ":9999")
		require.NoError(t, err)

		cfg, err := Load()
		require.NoError(t, err)

		assert.Equal(t, ":9999", cfg.HTTP.Port)
		assert.Equal(t, "test_token", cfg.Telegram.Token)
		assert.Equal(t, 10.0, cfg.RateLimit.Global.Limit)
		assert.Equal(t, 3*time.Minute, cfg.RateLimit.CleanupInterval)
	})

	t.Run("fail when required value is missing", func(t *testing.T) {
		err := os.Unsetenv("APP_JWT_SECRET")
		require.NoError(t, err)

		cfg, err := Load()
		assert.ErrorIs(t, err, ErrMissingEnvVar)
		assert.Nil(t, cfg)

		err = os.Setenv("APP_JWT_SECRET", "test_secret")
		require.NoError(t, err)
	})

	t.Run("postgres validation logic", func(t *testing.T) {
		err := os.Setenv("APP_DATABASE_URL", "postgres://localhost:5432/db")
		require.NoError(t, err)

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "postgres://localhost:5432/db", cfg.Postgres.ConnectionURL)

		err = os.Unsetenv("APP_DATABASE_URL")
		require.NoError(t, err)
	})
}
