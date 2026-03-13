package render

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type decodeTestStruct struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func TestDecode(t *testing.T) {
	t.Run("empty body returns ErrEmptyBody", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.ContentLength = 0
		var dst decodeTestStruct

		err := Decode(req, &dst)
		require.ErrorIs(t, err, ErrEmptyBody)
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name": "Alice", "email": "alice@example.com"`) // missing }
		req := httptest.NewRequest(http.MethodPost, "/", body)
		var dst decodeTestStruct

		err := Decode(req, &dst)
		require.Error(t, err)
	})

	t.Run("unknown field returns error", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name": "Alice", "email": "alice@example.com", "extra": 1}`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		var dst decodeTestStruct

		err := Decode(req, &dst)
		require.Error(t, err)
	})

	t.Run("multiple JSON objects returns ErrMultipleJSON", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"Alice","email":"alice@example.com"}{"name":"Bob","email":"bob@example.com"}`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		var dst decodeTestStruct

		err := Decode(req, &dst)
		require.ErrorIs(t, err, ErrMultipleJSON)
	})

	t.Run("valid JSON passes", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"Alice","email":"alice@example.com"}`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		var dst decodeTestStruct

		err := Decode(req, &dst)
		require.NoError(t, err)
		require.Equal(t, "Alice", dst.Name)
		require.Equal(t, "alice@example.com", dst.Email)
	})

	t.Run("fails validation", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"","email":"invalid-email"}`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		var dst decodeTestStruct

		err := Decode(req, &dst)
		require.Error(t, err)

		validationErrs := ValidationErrors(err)
		require.Equal(t, map[string]string{
			"name":  "field is required",
			"email": "must be a valid email",
		}, validationErrs)
	})
}
