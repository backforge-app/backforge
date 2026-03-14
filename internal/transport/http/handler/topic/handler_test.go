package topic

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
	servicetopic "github.com/backforge-app/backforge/internal/service/topic"
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
	topicID := uuid.New()
	userID := uuid.New()
	desc := "A topic about Go"

	validReq := createRequest{
		Title:       "Golang",
		Slug:        "golang",
		Description: &desc,
		CreatedBy:   &userID,
	}

	tests := []struct {
		name           string
		reqBody        interface{}
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			reqBody: validReq,
			mockSetup: func() {
				mockService.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(topicID, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Service Returns Invalid Data",
			reqBody: validReq,
			mockSetup: func() {
				mockService.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, servicetopic.ErrTopicInvalidData)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Conflict - Already Exists",
			reqBody: validReq,
			mockSetup: func() {
				mockService.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, servicetopic.ErrTopicAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:    "Internal Server Error",
			reqBody: validReq,
			mockSetup: func() {
				mockService.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, errors.New("database failure"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, err := json.Marshal(tt.reqBody)
			require.NoError(t, err, "failed to marshal request body")

			req := httptest.NewRequest(http.MethodPost, "/topics", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			handler.CreateHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_UpdateHandler(t *testing.T) {
	mockService, handler := setupTest(t)
	topicID := uuid.New()
	title := "New Title"

	validReq := updateRequest{
		Title: &title,
	}

	tests := []struct {
		name           string
		idParam        string
		reqBody        interface{}
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			idParam: topicID.String(),
			reqBody: validReq,
			mockSetup: func() {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid UUID",
			idParam:        "invalid-uuid",
			reqBody:        validReq,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Topic Not Found",
			idParam: topicID.String(),
			reqBody: validReq,
			mockSetup: func() {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(servicetopic.ErrTopicNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, err := json.Marshal(tt.reqBody)
			require.NoError(t, err, "failed to marshal request body")

			req := httptest.NewRequest(http.MethodPut, "/topics/"+tt.idParam, bytes.NewReader(body))
			req = addChiContext(req, "id", tt.idParam)
			rr := httptest.NewRecorder()

			handler.UpdateHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_GetByIDHandler(t *testing.T) {
	mockService, handler := setupTest(t)
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
					GetByID(gomock.Any(), topicID).
					Return(&domain.Topic{ID: topicID, Title: "Test"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Not Found",
			idParam: topicID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					GetByID(gomock.Any(), topicID).
					Return(nil, servicetopic.ErrTopicNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/topics/"+tt.idParam, nil)
			req = addChiContext(req, "id", tt.idParam)
			rr := httptest.NewRecorder()

			handler.GetByIDHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_GetBySlugHandler(t *testing.T) {
	mockService, handler := setupTest(t)
	slug := "test-topic"

	tests := []struct {
		name           string
		slugParam      string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:      "Success",
			slugParam: slug,
			mockSetup: func() {
				mockService.EXPECT().
					GetBySlug(gomock.Any(), slug).
					Return(&domain.Topic{Slug: slug, Title: "Test"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty Slug",
			slugParam:      "",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/topics/slug/"+tt.slugParam, nil)
			req = addChiContext(req, "slug", tt.slugParam)
			rr := httptest.NewRecorder()

			handler.GetBySlugHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_ListRowsHandler(t *testing.T) {
	mockService, handler := setupTest(t)

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockService.EXPECT().
					ListRows(gomock.Any()).
					Return([]*domain.TopicRow{
						{ID: uuid.New(), Title: "T1", Slug: "s1", QuestionCount: 5},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Service Error",
			mockSetup: func() {
				mockService.EXPECT().
					ListRows(gomock.Any()).
					Return(nil, errors.New("error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/topics", nil)
			rr := httptest.NewRecorder()

			handler.ListRowsHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
