package question

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	servicequestion "github.com/backforge-app/backforge/internal/service/question"
)

func setupTest(t *testing.T) (*MockService, *Handler) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	logger := zap.NewNop().Sugar()
	handler := NewHandler(mockService, logger)
	return mockService, handler
}

func addChiContext(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestHandler_CreateHandler(t *testing.T) {
	mockService, handler := setupTest(t)

	validTopicID := uuid.New()
	validCreatedBy := uuid.New()
	validReturnedID := uuid.New()

	validReqBody := createRequest{
		Title:     "What is a Goroutine?",
		Slug:      "what-is-goroutine",
		Content:   map[string]interface{}{"text": "Explain goroutines."},
		Level:     "beginner",
		TopicID:   &validTopicID,
		IsFree:    true,
		CreatedBy: &validCreatedBy,
	}

	tests := []struct {
		name           string
		reqBody        interface{}
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			reqBody: validReqBody,
			mockSetup: func() {
				mockService.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(validReturnedID, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Level",
			reqBody: createRequest{
				Title:   "Test",
				Slug:    "test",
				Content: map[string]interface{}{"text": "test"},
				Level:   "super-hard",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Conflict - Already Exists",
			reqBody: validReqBody,
			mockSetup: func() {
				mockService.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, servicequestion.ErrQuestionAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:    "Internal Server Error",
			reqBody: validReqBody,
			mockSetup: func() {
				mockService.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, errors.New("db down"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, err := json.Marshal(tt.reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/questions", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_UpdateHandler(t *testing.T) {
	mockService, handler := setupTest(t)

	questionID := uuid.New()
	newTitle := "Updated Title"
	newLevel := "medium"

	validReqBody := updateRequest{
		Title: &newTitle,
		Level: &newLevel,
	}

	tests := []struct {
		name           string
		questionIDStr  string
		reqBody        interface{}
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:          "Success",
			questionIDStr: questionID.String(),
			reqBody:       validReqBody,
			mockSetup: func() {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid UUID in URL",
			questionIDStr:  "not-a-uuid",
			reqBody:        validReqBody,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "Not Found",
			questionIDStr: questionID.String(),
			reqBody:       validReqBody,
			mockSetup: func() {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(servicequestion.ErrQuestionNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, err := json.Marshal(tt.reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/questions/"+tt.questionIDStr, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = addChiContext(req, "id", tt.questionIDStr)

			rr := httptest.NewRecorder()

			handler.Update(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_GetByIDHandler(t *testing.T) {
	mockService, handler := setupTest(t)

	questionID := uuid.New()
	mockQuestion := &domain.Question{
		ID:    questionID,
		Title: "Test Question",
		Slug:  "test-question",
		Level: domain.QuestionLevelBeginner,
		Tags:  []*domain.Tag{},
	}

	tests := []struct {
		name           string
		questionIDStr  string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:          "Success",
			questionIDStr: questionID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					GetByID(gomock.Any(), questionID).
					Return(mockQuestion, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID format",
			questionIDStr:  "123-abc",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "Not Found",
			questionIDStr: questionID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					GetByID(gomock.Any(), questionID).
					Return(nil, servicequestion.ErrQuestionNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/admin/questions/"+tt.questionIDStr, nil)
			req = addChiContext(req, "id", tt.questionIDStr)
			rr := httptest.NewRecorder()

			handler.GetByID(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_ListCardsHandler(t *testing.T) {
	mockService, handler := setupTest(t)

	mockCards := []*domain.QuestionCard{
		{
			ID:    uuid.New(),
			Title: "Card 1",
			Level: domain.QuestionLevelBeginner,
			Tags:  []string{"go", "concurrency"},
		},
	}

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:        "Success with defaults",
			queryParams: "",
			mockSetup: func() {
				mockService.EXPECT().
					ListCards(gomock.Any(), gomock.Any()).
					Return(mockCards, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Level Query Param",
			queryParams:    "?level=impossible",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Internal Error",
			queryParams: "?limit=10",
			mockSetup: func() {
				mockService.EXPECT().
					ListCards(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/questions"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			handler.ListCards(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_ListByTopicHandler(t *testing.T) {
	mockService, handler := setupTest(t)

	topicID := uuid.New()

	mockQuestions := []*domain.Question{
		{
			ID:      uuid.New(),
			Title:   "Question 1",
			Slug:    "question-1",
			Content: map[string]interface{}{"text": "content"},
			Level:   domain.QuestionLevelBeginner,
			TopicID: &topicID,
			IsFree:  true,
			Tags: []*domain.Tag{
				{ID: uuid.New()},
			},
		},
	}

	tests := []struct {
		name           string
		topicIDStr     string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:       "Success",
			topicIDStr: topicID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					ListByTopic(gomock.Any(), topicID).
					Return(mockQuestions, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Topic ID",
			topicIDStr:     "not-a-uuid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Internal Server Error",
			topicIDStr: topicID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					ListByTopic(gomock.Any(), topicID).
					Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(
				http.MethodGet,
				"/topics/"+tt.topicIDStr+"/questions",
				nil,
			)

			req = addChiContext(req, "id", tt.topicIDStr)

			rr := httptest.NewRecorder()

			handler.ListByTopic(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
