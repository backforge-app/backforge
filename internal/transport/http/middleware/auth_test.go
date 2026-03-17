package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAuth(t *testing.T) {
	secret := "test-secret"
	log := zap.NewNop().Sugar()
	userID := uuid.New()

	createToken := func(id uuid.UUID, exp time.Duration) string {
		claims := jwt.MapClaims{
			"sub": id.String(),
			"exp": time.Now().Add(exp).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		s, err := token.SignedString([]byte(secret))
		require.NoError(t, err)
		return s
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := UserIDFromContext(r.Context())
		if ok && id == userID {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusTeapot)
	})

	t.Run("valid token", func(t *testing.T) {
		token := createToken(userID, time.Hour)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		Auth(secret, log)(nextHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("expired token", func(t *testing.T) {
		token := createToken(userID, -time.Hour)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		Auth(secret, log)(nextHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		Auth(secret, log)(nextHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
