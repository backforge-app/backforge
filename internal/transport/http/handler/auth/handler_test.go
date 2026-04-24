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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	serviceauth "github.com/backforge-app/backforge/internal/service/auth"
	serviceuser "github.com/backforge-app/backforge/internal/service/user"
)

// newTestLogger creates a silent zap logger for testing to keep test output clean.
func newTestLogger(t *testing.T) *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.TimeKey = ""
	cfg.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	logger, err := cfg.Build(zap.AddCallerSkip(1))
	require.NoError(t, err)
	return logger.Sugar()
}

// newHandlerMocks creates a new Handler with injected mocked Service.
func newHandlerMocks(t *testing.T) (*Handler, *MockService) {
	ctrl := gomock.NewController(t)
	svc := NewMockService(ctrl)
	log := newTestLogger(t)

	return NewHandler(svc, log), svc
}

// performJSONRequest is a helper that converts a payload to JSON, executes the handler,
// and returns the ResponseRecorder. It accepts both strings (for testing invalid JSON)
// and struct payloads.
func performJSONRequest(t *testing.T, handler http.HandlerFunc, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		switch v := body.(type) {
		case string:
			buf.WriteString(v)
		default:
			err := json.NewEncoder(&buf).Encode(v)
			require.NoError(t, err, "failed to encode json payload")
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// ptr is a generic helper to quickly create pointers for inline test struct definitions.
func ptr[T any](v T) *T {
	return &v
}

func TestHandler_Register(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	tests := []struct {
		name           string
		payload        any
		setupMock      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Success",
			payload: registerRequest{
				Email:     "test@example.com",
				Password:  "securepass123",
				FirstName: "John",
			},
			setupMock: func() {
				svc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Bad Request - Invalid JSON",
			payload:        `{"email":"test@example.com", "password":}`,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request payload",
		},
		{
			name: "Bad Request - Validation Failed",
			payload: registerRequest{
				Email:     "invalid-email",
				Password:  "short",
				FirstName: "J",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
		{
			name: "Conflict - Email Taken",
			payload: registerRequest{
				Email:     "taken@example.com",
				Password:  "securepass123",
				FirstName: "John",
			},
			setupMock: func() {
				svc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(serviceuser.ErrUserEmailTaken)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  ErrEmailTaken.Error(),
		},
		{
			name: "Conflict - Username Taken",
			payload: registerRequest{
				Email:     "test@example.com",
				Password:  "securepass123",
				FirstName: "John",
				Username:  ptr("taken_user"),
			},
			setupMock: func() {
				svc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(serviceuser.ErrUserUsernameTaken)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  ErrUsernameTaken.Error(),
		},
		{
			name: "Internal Server Error",
			payload: registerRequest{
				Email:     "test@example.com",
				Password:  "securepass123",
				FirstName: "John",
			},
			setupMock: func() {
				svc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(errors.New("db down"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  ErrInternalServer.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			rr := performJSONRequest(t, handler.Register, tt.payload)
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
				assert.Equal(t, tt.expectedError, resp["error"])
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	tests := []struct {
		name           string
		payload        any
		setupMock      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Success",
			payload: loginRequest{
				Email:    "test@example.com",
				Password: "correctpassword",
			},
			setupMock: func() {
				svc.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return("access_token", "refresh_token", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Unauthorized - Invalid Credentials",
			payload: loginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMock: func() {
				svc.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return("", "", serviceauth.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrInvalidCredentials.Error(),
		},
		{
			name: "Unauthorized - Email Not Verified",
			payload: loginRequest{
				Email:    "test@example.com",
				Password: "correctpassword",
			},
			setupMock: func() {
				svc.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return("", "", serviceauth.ErrEmailNotVerified)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrEmailNotVerified.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			rr := performJSONRequest(t, handler.Login, tt.payload)
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
				assert.Equal(t, tt.expectedError, resp["error"])
			} else {
				var resp tokenResponse
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			}
		})
	}
}

func TestHandler_VerifyEmail(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	tests := []struct {
		name           string
		payload        verifyEmailRequest
		setupMock      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			payload: verifyEmailRequest{Token: "valid-token"},
			setupMock: func() {
				svc.EXPECT().VerifyEmail(gomock.Any(), "valid-token").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Bad Request - Invalid Token",
			payload: verifyEmailRequest{Token: "bad-token"},
			setupMock: func() {
				svc.EXPECT().VerifyEmail(gomock.Any(), "bad-token").Return(serviceauth.ErrInvalidVerificationToken)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			rr := performJSONRequest(t, handler.VerifyEmail, tt.payload)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_RequestPasswordReset(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	tests := []struct {
		name           string
		payload        requestResetRequest
		setupMock      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			payload: requestResetRequest{Email: "test@example.com"},
			setupMock: func() {
				svc.EXPECT().RequestPasswordReset(gomock.Any(), "test@example.com").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Internal Server Error",
			payload: requestResetRequest{Email: "test@example.com"},
			setupMock: func() {
				svc.EXPECT().RequestPasswordReset(gomock.Any(), "test@example.com").Return(errors.New("mail fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			rr := performJSONRequest(t, handler.RequestPasswordReset, tt.payload)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_ResetPassword(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	tests := []struct {
		name           string
		payload        resetPasswordRequest
		setupMock      func()
		expectedStatus int
	}{
		{
			name: "Success",
			payload: resetPasswordRequest{
				Token:       "valid-token",
				NewPassword: "new_secure_password",
			},
			setupMock: func() {
				svc.EXPECT().ResetPassword(gomock.Any(), "valid-token", "new_secure_password").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Bad Request - Invalid Token",
			payload: resetPasswordRequest{
				Token:       "bad-token",
				NewPassword: "new_secure_password",
			},
			setupMock: func() {
				svc.EXPECT().ResetPassword(gomock.Any(), "bad-token", "new_secure_password").Return(serviceauth.ErrInvalidVerificationToken)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			rr := performJSONRequest(t, handler.ResetPassword, tt.payload)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

// Заменяем GitHubCallback на YandexCallback
func TestHandler_YandexCallback(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	tests := []struct {
		name           string
		payload        yandexCallbackRequest
		setupMock      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			payload: yandexCallbackRequest{Code: "auth_code"},
			setupMock: func() {
				svc.EXPECT().LoginWithYandex(gomock.Any(), "auth_code").Return("access", "refresh", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Bad Request - OAuth Failed",
			payload: yandexCallbackRequest{Code: "bad_code"},
			setupMock: func() {
				svc.EXPECT().LoginWithYandex(gomock.Any(), "bad_code").Return("", "", serviceauth.ErrOAuthExchangeFailed)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			rr := performJSONRequest(t, handler.YandexCallback, tt.payload)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_Refresh(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	tests := []struct {
		name           string
		payload        refreshRequest
		setupMock      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			payload: refreshRequest{RefreshToken: "valid-token"},
			setupMock: func() {
				svc.EXPECT().Refresh(gomock.Any(), "valid-token").Return("access", "refresh", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Unauthorized - Revoked Token",
			payload: refreshRequest{RefreshToken: "revoked-token"},
			setupMock: func() {
				svc.EXPECT().Refresh(gomock.Any(), "revoked-token").Return("", "", serviceauth.ErrRefreshTokenRevoked)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:    "Internal Server Error",
			payload: refreshRequest{RefreshToken: "valid-token"},
			setupMock: func() {
				svc.EXPECT().Refresh(gomock.Any(), "valid-token").Return("", "", errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			rr := performJSONRequest(t, handler.Refresh, tt.payload)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_ResendVerification(t *testing.T) {
	handler, svc := newHandlerMocks(t)

	tests := []struct {
		name           string
		payload        any
		setupMock      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "Success",
			payload: resendVerificationRequest{Email: "test@example.com"},
			setupMock: func() {
				svc.EXPECT().ResendVerificationEmail(gomock.Any(), "test@example.com").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Bad Request - Invalid Email Format",
			payload:        resendVerificationRequest{Email: "not-an-email"},
			setupMock:      func() {}, // Decode fails, service is not called
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation failed",
		},
		{
			name:    "Bad Request - Already Verified",
			payload: resendVerificationRequest{Email: "verified@example.com"},
			setupMock: func() {
				svc.EXPECT().ResendVerificationEmail(gomock.Any(), "verified@example.com").Return(serviceauth.ErrEmailAlreadyVerified)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrAlreadyVerified.Error(),
		},
		{
			name:    "Internal Server Error",
			payload: resendVerificationRequest{Email: "error@example.com"},
			setupMock: func() {
				svc.EXPECT().ResendVerificationEmail(gomock.Any(), "error@example.com").Return(errors.New("db timeout"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  ErrInternalServer.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			rr := performJSONRequest(t, handler.ResendVerification, tt.payload)
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
				assert.Equal(t, tt.expectedError, resp["error"])
			}
		})
	}
}
