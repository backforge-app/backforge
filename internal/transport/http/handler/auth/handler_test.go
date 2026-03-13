package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	serviceauth "github.com/backforge-app/backforge/internal/service/auth"
)

func newTestLogger(t *testing.T) *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.TimeKey = "" // suppress timestamp
	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		t.Fatalf("failed to build zap logger: %v", err)
	}
	return logger.Sugar()
}

// newHandlerMocks creates a Handler with a mocked Service.
func newHandlerMocks(t *testing.T) (*Handler, *MockService) {
	ctrl := gomock.NewController(t)
	svc := NewMockService(ctrl)
	log := newTestLogger(t)
	return NewHandler(svc, log), svc
}

func performRequest(handlerFunc http.HandlerFunc, body any, t *testing.T) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	return rr
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder, dst any) {
	err := json.NewDecoder(rr.Body).Decode(dst)
	assert.NoError(t, err)
}

func TestLoginHandler_Success(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := loginRequest{
		ID:        1,
		FirstName: "John",
		AuthDate:  123456,
		Hash:      "hash",
	}

	expectedInput := serviceauth.TelegramLoginInput{
		ID:        req.ID,
		FirstName: req.FirstName,
		AuthDate:  req.AuthDate,
		Hash:      req.Hash,
	}

	svc.EXPECT().LoginWithTelegram(gomock.Any(), expectedInput).Return("access", "refresh", nil)

	rr := performRequest(handler.LoginHandler, req, t)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp loginResponse
	decodeBody(t, rr, &resp)
	assert.Equal(t, "access", resp.AccessToken)
	assert.Equal(t, "refresh", resp.RefreshToken)
}

func TestLoginHandler_InvalidTelegramAuth(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := loginRequest{
		ID:        1,
		FirstName: "John",
		AuthDate:  123456,
		Hash:      "hash",
	}

	svc.EXPECT().
		LoginWithTelegram(gomock.Any(), gomock.Any()).
		Return("", "", serviceauth.ErrInvalidTelegramAuth)

	rr := performRequest(handler.LoginHandler, req, t)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInvalidCredentials.Error(), resp["error"])
}

func TestLoginHandler_TelegramAuthExpired(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := loginRequest{
		ID:        1,
		FirstName: "John",
		AuthDate:  123456,
		Hash:      "hash",
	}

	svc.EXPECT().
		LoginWithTelegram(gomock.Any(), gomock.Any()).
		Return("", "", serviceauth.ErrTelegramAuthExpired)

	rr := performRequest(handler.LoginHandler, req, t)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInvalidCredentials.Error(), resp["error"])
}

func TestLoginHandler_InternalError(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := loginRequest{
		ID:        1,
		FirstName: "John",
		AuthDate:  123456,
		Hash:      "hash",
	}

	svc.EXPECT().
		LoginWithTelegram(gomock.Any(), gomock.Any()).
		Return("", "", errors.New("db failure"))

	rr := performRequest(handler.LoginHandler, req, t)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInternalServer.Error(), resp["error"])
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	handler, _ := newHandlerMocks(t)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.LoginHandler(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, "invalid request payload", resp["error"])
}

func TestRefreshHandler_Success(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := refreshRequest{RefreshToken: "token"}

	svc.EXPECT().
		RefreshTokens(gomock.Any(), "token").
		Return("access-new", "refresh-new", nil)

	rr := performRequest(handler.RefreshHandler, req, t)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp refreshResponse
	decodeBody(t, rr, &resp)

	assert.Equal(t, "access-new", resp.AccessToken)
	assert.Equal(t, "refresh-new", resp.RefreshToken)
}

func TestRefreshHandler_InvalidToken(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := refreshRequest{RefreshToken: "bad"}

	svc.EXPECT().
		RefreshTokens(gomock.Any(), "bad").
		Return("", "", serviceauth.ErrRefreshTokenInvalid)

	rr := performRequest(handler.RefreshHandler, req, t)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInvalidCredentials.Error(), resp["error"])
}

func TestRefreshHandler_RevokedToken(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := refreshRequest{RefreshToken: "revoked"}

	svc.EXPECT().
		RefreshTokens(gomock.Any(), "revoked").
		Return("", "", serviceauth.ErrRefreshTokenRevoked)

	rr := performRequest(handler.RefreshHandler, req, t)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInvalidCredentials.Error(), resp["error"])
}

func TestRefreshHandler_TokenCollision(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := refreshRequest{RefreshToken: "token"}

	svc.EXPECT().
		RefreshTokens(gomock.Any(), "token").
		Return("", "", serviceauth.ErrRefreshTokenAlreadyExists)

	rr := performRequest(handler.RefreshHandler, req, t)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInternalServer.Error(), resp["error"])
}

func TestRefreshHandler_InternalError(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := refreshRequest{RefreshToken: "token"}

	svc.EXPECT().
		RefreshTokens(gomock.Any(), "token").
		Return("", "", errors.New("db failure"))

	rr := performRequest(handler.RefreshHandler, req, t)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInternalServer.Error(), resp["error"])
}

func TestRefreshHandler_ValidationError(t *testing.T) {
	handler, _ := newHandlerMocks(t)

	req := refreshRequest{}

	rr := performRequest(handler.RefreshHandler, req, t)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]any
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)

	errMsg, ok := resp["error"].(string)
	assert.True(t, ok, "error key should be string")
	assert.Equal(t, "validation failed", errMsg)

	details, ok := resp["details"].(map[string]any)
	if !ok {
		raw, ok := resp["details"].(map[string]string)
		assert.True(t, ok, "details key should exist")
		details = make(map[string]any, len(raw))
		for k, v := range raw {
			details[k] = v
		}
	}
	assert.NotEmpty(t, details)
	assert.Contains(t, details, "refresh_token")
	assert.Equal(t, "field is required", details["refresh_token"])
}
