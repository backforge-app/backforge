package httputil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestParseQueryInt(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		key      string
		def      int
		expected int
	}{
		{"valid int", "/?page=2", "page", 1, 2},
		{"missing key", "/?other=5", "page", 1, 1},
		{"invalid int", "/?page=abc", "page", 1, 1},
		{"empty value", "/?page=", "page", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			result := ParseQueryInt(req, tt.key, tt.def)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetQueryPtrString(t *testing.T) {
	t.Run("present value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?name=backforge", nil)
		res := GetQueryPtrString(req, "name")
		assert.NotNil(t, res)
		assert.Equal(t, "backforge", *res)
	})

	t.Run("missing value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := GetQueryPtrString(req, "name")
		assert.Nil(t, res)
	})
}

func TestURLParamUUID(t *testing.T) {
	id := uuid.New()

	t.Run("valid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		// Setup chi context for URL parameters
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		result, ok := URLParamUUID(req, "id")
		assert.True(t, ok)
		assert.Equal(t, id, result)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		result, ok := URLParamUUID(req, "id")
		assert.False(t, ok)
		assert.Equal(t, uuid.Nil, result)
	})

	t.Run("missing parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		result, ok := URLParamUUID(req, "nonexistent")
		assert.False(t, ok)
		assert.Equal(t, uuid.Nil, result)
	})
}
