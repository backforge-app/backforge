package tag

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
	servicetag "github.com/backforge-app/backforge/internal/service/tag"
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
	tagID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name           string
		reqBody        interface{}
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "Success",
			reqBody: createRequest{
				Name:      "Golang",
				CreatedBy: &userID,
			},
			mockSetup: func() {
				mockService.EXPECT().
					Create(gomock.Any(), "Golang", &userID).
					Return(tagID, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Conflict - Already Exists",
			reqBody: createRequest{
				Name: "Duplicate",
			},
			mockSetup: func() {
				mockService.EXPECT().
					Create(gomock.Any(), "Duplicate", gomock.Any()).
					Return(uuid.Nil, servicetag.ErrTagAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Internal Error",
			reqBody: createRequest{
				Name: "Error",
			},
			mockSetup: func() {
				mockService.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uuid.Nil, errors.New("db failure"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			body, err := json.Marshal(tt.reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			handler.CreateHandler(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_GetByIDHandler(t *testing.T) {
	mockService, handler := setupTest(t)
	tagID := uuid.New()

	tests := []struct {
		name           string
		idParam        string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			idParam: tagID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					GetByID(gomock.Any(), tagID).
					Return(&domain.Tag{ID: tagID, Name: "Found"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Not Found",
			idParam: tagID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					GetByID(gomock.Any(), tagID).
					Return(nil, servicetag.ErrTagNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid UUID",
			idParam:        "not-a-uuid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/tags/"+tt.idParam, nil)
			req = addChiContext(req, "id", tt.idParam)
			rr := httptest.NewRecorder()

			handler.GetByIDHandler(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_ListHandler(t *testing.T) {
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
					List(gomock.Any()).
					Return([]*domain.Tag{{ID: uuid.New(), Name: "Tag1"}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Service Error",
			mockSetup: func() {
				mockService.EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/tags", nil)
			rr := httptest.NewRecorder()

			handler.ListHandler(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_DeleteHandler(t *testing.T) {
	mockService, handler := setupTest(t)
	tagID := uuid.New()

	tests := []struct {
		name           string
		idParam        string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:    "Success",
			idParam: tagID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					Delete(gomock.Any(), tagID).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Not Found",
			idParam: tagID.String(),
			mockSetup: func() {
				mockService.EXPECT().
					Delete(gomock.Any(), tagID).
					Return(servicetag.ErrTagNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodDelete, "/tags/"+tt.idParam, nil)
			req = addChiContext(req, "id", tt.idParam)
			rr := httptest.NewRecorder()

			handler.DeleteHandler(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
