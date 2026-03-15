package analytics

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

func TestHandler_GetOverallProgress(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

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
					GetOverallProgress(gomock.Any(), userID).
					Return(&domain.OverallProgress{
						Total:   100,
						Known:   20,
						Learned: 30,
						Skipped: 5,
						New:     45,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized",
			auth:           false,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Service Error",
			auth: true,
			mockSetup: func() {
				mockService.EXPECT().
					GetOverallProgress(gomock.Any(), userID).
					Return(nil, errors.New("database failure"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/analytics/overall", nil)
			if tt.auth {
				req = withAuthContext(req, userID)
			}
			rr := httptest.NewRecorder()

			handler.GetOverallProgress(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_GetProgressByTopicPercent(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

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
					GetProgressByTopicPercent(gomock.Any(), userID).
					Return([]*domain.TopicProgressPercent{
						{TopicID: uuid.New(), Completed: 5, Total: 10, Percent: 50.0},
						{TopicID: uuid.New(), Completed: 10, Total: 10, Percent: 100.0},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized",
			auth:           false,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Service Error",
			auth: true,
			mockSetup: func() {
				mockService.EXPECT().
					GetProgressByTopicPercent(gomock.Any(), userID).
					Return(nil, errors.New("failed to fetch list"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/analytics/topics", nil)
			if tt.auth {
				req = withAuthContext(req, userID)
			}
			rr := httptest.NewRecorder()

			handler.GetProgressByTopicPercent(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_ResetAllProgress(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()

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
					ResetAllProgress(gomock.Any(), userID).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized",
			auth:           false,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Service Error",
			auth: true,
			mockSetup: func() {
				mockService.EXPECT().
					ResetAllProgress(gomock.Any(), userID).
					Return(errors.New("reset failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodDelete, "/analytics/reset", nil)
			if tt.auth {
				req = withAuthContext(req, userID)
			}
			rr := httptest.NewRecorder()

			handler.ResetAllProgress(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
