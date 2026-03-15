package user

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/backforge-app/backforge/internal/domain"
	serviceuser "github.com/backforge-app/backforge/internal/service/user"
	"github.com/backforge-app/backforge/internal/transport/http/middleware"
)

// setupTest initializes the mock controller, service, and handler.
func setupTest(t *testing.T) (*MockService, *Handler) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	logger := zap.NewNop().Sugar()
	handler := NewHandler(mockService, logger)
	return mockService, handler
}

// withAuthContext injects a user ID into the request context.
func withAuthContext(r *http.Request, userID uuid.UUID) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, userID))
}

func TestHandler_GetProfile(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

	// Helper pointers for optional fields
	lastName := "Doe"
	username := "johndoe"

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
		auth           bool
	}{
		{
			name: "Success",
			auth: true,
			mockSetup: func() {
				mockService.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(&domain.User{
						ID:         userID,
						TelegramID: 12345678,
						FirstName:  "John",
						LastName:   &lastName,
						Username:   &username,
						Role:       domain.UserRoleUser,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized - Missing Context",
			auth:           false,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "User Not Found",
			auth: true,
			mockSetup: func() {
				mockService.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(nil, serviceuser.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Internal Service Error",
			auth: true,
			mockSetup: func() {
				mockService.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(nil, errors.New("unexpected database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
			if tt.auth {
				req = withAuthContext(req, userID)
			}
			rr := httptest.NewRecorder()

			handler.GetProfile(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
