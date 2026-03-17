package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockUserRoleChecker struct {
	mock.Mock
}

func (m *mockUserRoleChecker) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func TestAdminOnly(t *testing.T) {
	log := zap.NewNop().Sugar()
	userID := uuid.New()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("success admin access", func(t *testing.T) {
		mockSvc := new(mockUserRoleChecker)
		mockSvc.On("IsAdmin", mock.Anything, userID).Return(true, nil)

		mw := AdminOnly(log, mockSvc)
		h := mw(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID))
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("forbidden for non-admin", func(t *testing.T) {
		mockSvc := new(mockUserRoleChecker)
		mockSvc.On("IsAdmin", mock.Anything, userID).Return(false, nil)

		mw := AdminOnly(log, mockSvc)
		h := mw(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID))
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("error from service", func(t *testing.T) {
		mockSvc := new(mockUserRoleChecker)
		mockSvc.On("IsAdmin", mock.Anything, userID).Return(false, errors.New("db error"))

		mw := AdminOnly(log, mockSvc)
		h := mw(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID))
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
