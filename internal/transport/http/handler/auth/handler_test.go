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

	"github.com/backforge-app/backforge/internal/config"
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

	cfg := &config.Config{
		Env: "development",
	}

	return NewHandler(svc, log, cfg), svc
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

func TestLogin_Success(t *testing.T) {
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

	rr := performRequest(handler.Login, req, t)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp loginResponse
	decodeBody(t, rr, &resp)
	assert.Equal(t, "access", resp.AccessToken)
	assert.Equal(t, "refresh", resp.RefreshToken)
}

func TestLogin_InvalidTelegramAuth(t *testing.T) {
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

	rr := performRequest(handler.Login, req, t)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInvalidCredentials.Error(), resp["error"])
}

func TestLogin_TelegramAuthExpired(t *testing.T) {
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

	rr := performRequest(handler.Login, req, t)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInvalidCredentials.Error(), resp["error"])
}

func TestLogin_InternalError(t *testing.T) {
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

	rr := performRequest(handler.Login, req, t)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInternalServer.Error(), resp["error"])
}

func TestLogin_InvalidJSON(t *testing.T) {
	handler, _ := newHandlerMocks(t)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, "invalid request payload", resp["error"])
}

func TestRefresh_Success(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := refreshRequest{RefreshToken: "token"}

	svc.EXPECT().
		Refresh(gomock.Any(), "token").
		Return("access-new", "refresh-new", nil)

	rr := performRequest(handler.Refresh, req, t)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp refreshResponse
	decodeBody(t, rr, &resp)

	assert.Equal(t, "access-new", resp.AccessToken)
	assert.Equal(t, "refresh-new", resp.RefreshToken)
}

func TestRefresh_InvalidToken(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := refreshRequest{RefreshToken: "bad"}

	svc.EXPECT().
		Refresh(gomock.Any(), "bad").
		Return("", "", serviceauth.ErrRefreshTokenInvalid)

	rr := performRequest(handler.Refresh, req, t)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInvalidCredentials.Error(), resp["error"])
}

func TestRefresh_RevokedToken(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := refreshRequest{RefreshToken: "revoked"}

	svc.EXPECT().
		Refresh(gomock.Any(), "revoked").
		Return("", "", serviceauth.ErrRefreshTokenRevoked)

	rr := performRequest(handler.Refresh, req, t)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInvalidCredentials.Error(), resp["error"])
}

func TestRefresh_TokenCollision(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := refreshRequest{RefreshToken: "token"}

	svc.EXPECT().
		Refresh(gomock.Any(), "token").
		Return("", "", serviceauth.ErrRefreshTokenAlreadyExists)

	rr := performRequest(handler.Refresh, req, t)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInternalServer.Error(), resp["error"])
}

func TestRefresh_InternalError(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	req := refreshRequest{RefreshToken: "token"}

	svc.EXPECT().
		Refresh(gomock.Any(), "token").
		Return("", "", errors.New("db failure"))

	rr := performRequest(handler.Refresh, req, t)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var resp map[string]any
	decodeBody(t, rr, &resp)
	assert.Equal(t, ErrInternalServer.Error(), resp["error"])
}

func TestRefresh_ValidationError(t *testing.T) {
	handler, _ := newHandlerMocks(t)

	req := refreshRequest{}

	rr := performRequest(handler.Refresh, req, t)

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

func TestHandler_DevLogin(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name           string
		env            string
		requestBody    any
		setupMock      func(svc *MockService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "Success",
			env:         "development",
			requestBody: devLoginRequest{UserID: validUUID},
			setupMock: func(svc *MockService) {
				svc.EXPECT().
					DevLogin(gomock.Any(), gomock.Any()).
					Return("access", "refresh", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Blocked in Production",
			env:            "production",
			requestBody:    devLoginRequest{UserID: validUUID},
			setupMock:      func(svc *MockService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid UUID Format",
			env:            "development",
			requestBody:    devLoginRequest{UserID: "not-a-uuid"},
			setupMock:      func(svc *MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
		{
			name:        "Service Error",
			env:         "development",
			requestBody: devLoginRequest{UserID: validUUID},
			setupMock: func(svc *MockService) {
				svc.EXPECT().
					DevLogin(gomock.Any(), gomock.Any()).
					Return("", "", errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  ErrInternalServer.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, svc := newHandlerMocks(t)
			handler.cfg.Env = tt.env
			tt.setupMock(svc)

			rr := performRequest(handler.DevLogin, tt.requestBody, t)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var resp map[string]any
				decodeBody(t, rr, &resp)
				assert.Equal(t, tt.expectedError, resp["error"])
			} else if tt.expectedStatus == http.StatusOK {
				var resp loginResponse
				decodeBody(t, rr, &resp)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			}
		})
	}
}
