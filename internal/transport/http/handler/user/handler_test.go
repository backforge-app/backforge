package user

import (
	"bytes"
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

// withAuthContext injects a user ID into the request context to simulate an authenticated user.
func withAuthContext(r *http.Request, userID uuid.UUID) *http.Request {
	return r.WithContext(middleware.WithUserID(r.Context(), userID))
}

// ptr is a generic helper to quickly create pointers for inline test struct definitions.
func ptr[T any](v T) *T {
	return &v
}

func TestHandler_GetProfile(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

	tests := []struct {
		name           string
		auth           bool
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "Success",
			auth: true,
			mockSetup: func() {
				mockService.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(&domain.User{
						ID:              userID,
						Email:           "test@example.com",
						IsEmailVerified: true,
						FirstName:       "John",
						LastName:        ptr("Doe"),
						Username:        ptr("johndoe"),
						Role:            domain.UserRoleUser,
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

func TestHandler_UpdateProfile(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

	tests := []struct {
		name           string
		auth           bool
		payload        string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			auth:    true,
			payload: `{"first_name":"Jane","username":"jane_doe"}`,
			mockSetup: func() {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, input serviceuser.UpdateInput) error {
						assert.Equal(t, userID, input.ID)
						assert.Equal(t, "Jane", *input.FirstName)
						assert.Equal(t, "jane_doe", *input.Username)
						return nil
					})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized - Missing Context",
			auth:           false,
			payload:        `{"first_name":"Jane"}`,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Bad Request - Invalid JSON",
			auth:           true,
			payload:        `{"first_name": "Jane"`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Bad Request - Validation Failed (Username too short)",
			auth:           true,
			payload:        `{"username":"a"}`,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Conflict - Username Already Taken",
			auth:    true,
			payload: `{"username":"taken_username"}`,
			mockSetup: func() {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(serviceuser.ErrUserUsernameTaken)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:    "Not Found - User Deleted Mid-Session",
			auth:    true,
			payload: `{"first_name":"Jane"}`,
			mockSetup: func() {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(serviceuser.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "Internal Server Error",
			auth:    true,
			payload: `{"first_name":"Jane"}`,
			mockSetup: func() {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(errors.New("database timeout"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			if tt.auth {
				req = withAuthContext(req, userID)
			}

			rr := httptest.NewRecorder()
			handler.UpdateProfile(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
