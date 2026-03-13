package render

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSON_OK_Created_NoContent_Msg(t *testing.T) {
	t.Run("JSON function writes correct status and body", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"hello": "world"}

		err := JSON(w, http.StatusTeapot, data)
		require.NoError(t, err)
		require.Equal(t, http.StatusTeapot, w.Code)
		require.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Equal(t, data, resp)
	})

	t.Run("OK writes 200 with data", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := map[string]string{"ok": "true"}

		err := OK(w, payload)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Equal(t, payload, resp)
	})

	t.Run("Created writes 201 with data", func(t *testing.T) {
		w := httptest.NewRecorder()
		payload := map[string]string{"created": "yes"}

		err := Created(w, payload)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Equal(t, payload, resp)
	})

	t.Run("NoContent writes 204 without body", func(t *testing.T) {
		w := httptest.NewRecorder()
		NoContent(w)
		require.Equal(t, http.StatusNoContent, w.Code)
		require.Equal(t, 0, w.Body.Len())
	})

	t.Run("Msg writes standard message response", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := Msg(w, "success")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code)

		var resp Message
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Equal(t, "success", resp.Message)
	})
}
