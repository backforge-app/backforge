package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func TestRateLimiter(t *testing.T) {
	log := zap.NewNop().Sugar()
	// Allow only 1 request per second with 1 burst.
	limit := rate.Limit(1)
	burst := 1
	cleanup := time.Minute

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allow single request", func(t *testing.T) {
		mw := RateLimiter(log, limit, burst, cleanup)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.1.1.1"
		rr := httptest.NewRecorder()

		mw(nextHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("block excessive requests", func(t *testing.T) {
		mw := RateLimiter(log, limit, burst, cleanup)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "2.2.2.2"

		// First request - OK.
		mw(nextHandler).ServeHTTP(httptest.NewRecorder(), req)

		// Second request immediate - 429.
		rr := httptest.NewRecorder()
		mw(nextHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	})
}
