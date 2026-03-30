package progress

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/backforge-app/backforge/internal/domain"
	serviceprogress "github.com/backforge-app/backforge/internal/service/progress"
	"github.com/backforge-app/backforge/internal/transport/http/middleware"
)

func setupTest(t *testing.T) (*MockService, *Handler) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	logger := zap.NewNop().Sugar()
	handler := NewHandler(mockService, logger)
	return mockService, handler
}

func withAuthContext(r *http.Request, userID uuid.UUID) *http.Request {
	return r.WithContext(middleware.WithUserID(r.Context(), userID))
}

func addChiContext(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestHandler_Mark(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	topicID := uuid.New()
	questionID := uuid.New()

	reqPayload := markRequest{
		TopicID:    topicID,
		QuestionID: questionID,
	}

	input := serviceprogress.MarkQuestionInput{
		UserID:     userID,
		TopicID:    topicID,
		QuestionID: questionID,
	}

	tests := []struct {
		name           string
		method         func(w http.ResponseWriter, r *http.Request)
		mockSetup      func()
		expectedStatus int
		auth           bool
	}{
		{
			name:   "MarkKnown Success",
			method: handler.MarkKnown,
			mockSetup: func() {
				mockService.EXPECT().MarkKnown(gomock.Any(), input).Return(nil)
			},
			expectedStatus: http.StatusOK,
			auth:           true,
		},
		{
			name:           "MarkLearned Unauthorized",
			method:         handler.MarkLearned,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			auth:           false,
		},
		{
			name:   "MarkSkipped Service Error",
			method: handler.MarkSkipped,
			mockSetup: func() {
				mockService.EXPECT().MarkSkipped(gomock.Any(), input).Return(serviceprogress.ErrInvalidProgressStatus)
			},
			expectedStatus: http.StatusBadRequest,
			auth:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			body, err := json.Marshal(reqPayload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			if tt.auth {
				req = withAuthContext(req, userID)
			}
			rr := httptest.NewRecorder()

			tt.method(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_GetTopicProgress(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	topicID := uuid.New()

	tests := []struct {
		name           string
		idParam        string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			idParam: topicID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					GetByTopic(gomock.Any(), userID, topicID).
					Return(&domain.TopicProgressAggregate{Known: 5, New: 10}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Not Found",
			idParam: topicID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					GetByTopic(gomock.Any(), userID, topicID).
					Return(nil, serviceprogress.ErrTopicProgressNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid UUID",
			idParam:        "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = withAuthContext(req, userID)
			req = addChiContext(req, "id", tt.idParam)
			rr := httptest.NewRecorder()

			handler.GetTopicProgress(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_GetQuestionProgress(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	questionID := uuid.New()

	tests := []struct {
		name           string
		idParam        string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			idParam: questionID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					GetByUserAndQuestion(gomock.Any(), userID, questionID).
					Return(&domain.UserQuestionProgress{Status: domain.ProgressStatusLearned}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Nil Result Returns 404",
			idParam: questionID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					GetByUserAndQuestion(gomock.Any(), userID, questionID).
					Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = withAuthContext(req, userID)
			req = addChiContext(req, "id", tt.idParam)
			rr := httptest.NewRecorder()

			handler.GetQuestionProgress(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_ResetTopic(t *testing.T) {
	mockService, handler := setupTest(t)
	userID := uuid.New()
	topicID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockService.EXPECT().
			ResetTopicProgress(gomock.Any(), userID, topicID).
			Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		req = withAuthContext(req, userID)
		req = addChiContext(req, "id", topicID.String())
		rr := httptest.NewRecorder()

		handler.ResetTopic(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
