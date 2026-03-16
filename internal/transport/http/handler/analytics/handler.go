// Package analytics implements HTTP handlers for analytics management.
package analytics

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/backforge-app/backforge/internal/transport/http/middleware"
	"github.com/backforge-app/backforge/internal/transport/http/render"
)

// Handler handles analytics-related HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates a new analytics Handler.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// GetOverallProgress godoc
// @Summary Get overall progress
// @Description Returns aggregated statistics for the authenticated user's dashboard
// @Tags Analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} overallProgressResponse
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /analytics/overall [get]
//
// GetOverallProgress handles GET /analytics/overall requests.
// It retrieves aggregated statistics for the authenticated user dashboard.
func (h *Handler) GetOverallProgress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Assuming user ID is passed as a URL param. Alternatively, extract it from auth middleware context.
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		h.log.Warn("unauthorized access")

		if sendErr := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send unauthorized response")
		}
		return
	}

	progress, err := h.service.GetOverallProgress(ctx, userID)
	if err != nil {
		h.log.With(zap.Error(err)).Error("get overall progress service failed")
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
		}
		return
	}

	if sendErr := render.OK(w, toOverallProgressResponse(progress)); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send overall progress response")
	}
}

// GetProgressByTopicPercent godoc
// @Summary Get topic progress percentages
// @Description Returns completion percentages for each topic for the authenticated user
// @Tags Analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {array} topicProgressPercentResponse
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /analytics/topics [get]
//
// GetProgressByTopicPercent handles GET /analytics/topics requests.
// It returns completion percentages for each topic for the authenticated user.
func (h *Handler) GetProgressByTopicPercent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		h.log.Warn("unauthorized access")

		if sendErr := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send unauthorized response")
		}
		return
	}

	topicProgress, err := h.service.GetProgressByTopicPercent(ctx, userID)
	if err != nil {
		h.log.With(zap.Error(err)).Error("get topic progress percentages service failed")
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
		}
		return
	}

	resp := make([]topicProgressPercentResponse, len(topicProgress))
	for i, p := range topicProgress {
		resp[i] = toTopicProgressPercentResponse(p)
	}

	if sendErr := render.OK(w, resp); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send topic progress response")
	}
}

// ResetAllProgress godoc
// @Summary Reset all progress
// @Description Reset all question and topic progress for the authenticated user
// @Tags Analytics
// @Produce json
// @Security BearerAuth
// @Success 200
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /analytics/reset [delete]
//
// ResetAllProgress handles DELETE /analytics/reset requests.
// It resets all stored question and topic progress for the authenticated user.
func (h *Handler) ResetAllProgress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		h.log.Warn("unauthorized access")

		if sendErr := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send unauthorized response")
		}
		return
	}

	err := h.service.ResetAllProgress(ctx, userID)
	if err != nil {
		h.log.With(zap.Error(err)).Error("reset all progress service failed")
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
		}
		return
	}

	// Respond with an empty 200 OK or 204 No Content
	if sendErr := render.OK(w, nil); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send reset progress response")
	}
}
