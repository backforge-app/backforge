package render

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFail(t *testing.T) {
	t.Run("with specific error", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := Fail(w, http.StatusBadRequest, errors.New("some error"))

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, w.Code)

		var resp Error
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Equal(t, "some error", resp.Error)
		require.Nil(t, resp.Details)
	})

	t.Run("with nil error", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := Fail(w, http.StatusInternalServerError, nil)

		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, w.Code)

		var resp Error
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Equal(t, "unknown error", resp.Error)
		require.Nil(t, resp.Details)
	})
}

func TestFailMessage(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := FailMessage(w, http.StatusBadRequest, "custom error")

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, w.Code)

		var resp Error
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Equal(t, "custom error", resp.Error)
		require.Nil(t, resp.Details)
	})

	t.Run("empty message falls back", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := FailMessage(w, http.StatusInternalServerError, "")

		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, w.Code)

		var resp Error
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Equal(t, "unknown error", resp.Error)
		require.Nil(t, resp.Details)
	})
}

func TestFailWithDetails(t *testing.T) {
	t.Run("with message and details", func(t *testing.T) {
		w := httptest.NewRecorder()
		details := map[string]string{"field": "required"}
		err := FailWithDetails(w, http.StatusBadRequest, "validation error", details)

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, w.Code)

		var resp Error
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Equal(t, "validation error", resp.Error)

		actual := make(map[string]string)
		if m, ok := resp.Details.(map[string]interface{}); ok {
			for k, v := range m {
				actual[k] = v.(string)
			}
		}
		require.Equal(t, details, actual)
	})

	t.Run("empty message falls back", func(t *testing.T) {
		w := httptest.NewRecorder()
		details := map[string]string{"field": "required"}
		err := FailWithDetails(w, http.StatusInternalServerError, "", details)

		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, w.Code)

		var resp Error
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Equal(t, "unknown error", resp.Error)

		actual := make(map[string]string)
		if m, ok := resp.Details.(map[string]interface{}); ok {
			for k, v := range m {
				actual[k] = v.(string)
			}
		}
		require.Equal(t, details, actual)
	})
}
