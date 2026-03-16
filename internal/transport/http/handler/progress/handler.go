// Package progress implements HTTP handlers for tracking user learning progress.
package progress

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/backforge-app/backforge/internal/service/progress"
	"github.com/backforge-app/backforge/internal/transport/http/httputil"
	"github.com/backforge-app/backforge/internal/transport/http/middleware"
	"github.com/backforge-app/backforge/internal/transport/http/render"
)

// Handler handles progress-related HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates a new progress Handler with the provided service and logger.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{service: service, log: log}
}

// MarkKnown godoc
// @Summary Mark question as known
// @Description Mark a question as known for the authenticated user
// @Tags Progress
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param progress body markRequest true "Progress payload"
// @Success 200
// @Failure 400 {object} render.Error "Invalid request data"
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 404 {object} render.Error "Progress not found"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /progress/known [post]
//
// MarkKnown handles POST /progress/known requests.
func (h *Handler) MarkKnown(w http.ResponseWriter, r *http.Request) {
	h.processMarkAction(w, r, h.service.MarkKnown, "mark known")
}

// MarkLearned godoc
// @Summary Mark question as learned
// @Description Mark a question as learned for the authenticated user
// @Tags Progress
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param progress body markRequest true "Progress payload"
// @Success 200
// @Failure 400 {object} render.Error "Invalid request data"
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 404 {object} render.Error "Progress not found"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /progress/learned [post]
//
// MarkLearned handles POST /progress/learned requests.
func (h *Handler) MarkLearned(w http.ResponseWriter, r *http.Request) {
	h.processMarkAction(w, r, h.service.MarkLearned, "mark learned")
}

// MarkSkipped godoc
// @Summary Mark question as skipped
// @Description Mark a question as skipped for the authenticated user
// @Tags Progress
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param progress body markRequest true "Progress payload"
// @Success 200
// @Failure 400 {object} render.Error "Invalid request data"
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 404 {object} render.Error "Progress not found"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /progress/skipped [post]
//
// MarkSkipped handles POST /progress/skipped requests.
func (h *Handler) MarkSkipped(w http.ResponseWriter, r *http.Request) {
	h.processMarkAction(w, r, h.service.MarkSkipped, "mark skipped")
}

// GetTopicProgress godoc
// @Summary Get topic progress
// @Description Get aggregated progress for a topic for the authenticated user
// @Tags Progress
// @Produce json
// @Security BearerAuth
// @Param id path string true "Topic ID"
// @Success 200 {object} aggregateResponse
// @Failure 400 {object} render.Error "Invalid topic ID"
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 404 {object} render.Error "Progress not found"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /progress/topics/{id} [get]
//
// GetTopicProgress handles GET /progress/topics/{id} requests.
// It retrieves the aggregated progress for a specific topic for the authenticated user.
func (h *Handler) GetTopicProgress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		h.log.Warn("unauthorized access to topic progress")

		if err := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); err != nil {
			h.log.With(zap.Error(err)).Warn("failed to send unauthorized response")
		}
		return
	}

	topicID, ok := httputil.URLParamUUID(r, "id")
	if !ok {
		h.handleError(w, ErrInvalidTopicID, "get topic progress")
		return
	}

	agg, err := h.service.GetByTopic(ctx, userID, topicID)
	if err != nil {
		h.handleError(w, err, "get topic progress")
		return
	}

	resp := aggregateResponse{
		Known:           agg.Known,
		Learned:         agg.Learned,
		Skipped:         agg.Skipped,
		New:             agg.New,
		CurrentPosition: agg.CurrentPosition,
	}

	if err := render.OK(w, resp); err != nil {
		h.log.With(zap.Error(err)).Warn("failed to send topic progress response")
	}
}

// GetQuestionProgress godoc
// @Summary Get question progress
// @Description Get progress of a specific question for the authenticated user
// @Tags Progress
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question ID"
// @Success 200 {object} domain.UserQuestionProgress
// @Failure 400 {object} render.Error "Invalid question ID"
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 404 {object} render.Error "Progress not found"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /progress/questions/{id} [get]
//
// GetQuestionProgress handles GET /progress/questions/{id} requests.
// It retrieves the progress of a specific question for the authenticated user.
func (h *Handler) GetQuestionProgress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		h.log.Warn("unauthorized access to question progress")

		if err := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); err != nil {
			h.log.With(zap.Error(err)).Warn("failed to send unauthorized response")
		}
		return
	}

	questionID, ok := httputil.URLParamUUID(r, "id")
	if !ok {
		h.handleError(w, ErrInvalidQuestionID, "get question progress")
		return
	}

	progressObj, err := h.service.GetByUserAndQuestion(ctx, userID, questionID)
	if err != nil {
		h.handleError(w, err, "get question progress")
		return
	}

	if progressObj == nil {
		h.log.Warnf("progress object is nil for user %s question %s", userID, questionID)
		if err := render.Fail(w, http.StatusNotFound, ErrProgressNotFound); err != nil {
			h.log.With(zap.Error(err)).Warn("failed to send not found response")
		}
		return
	}

	if err := render.OK(w, progressObj); err != nil {
		h.log.With(zap.Error(err)).Warn("failed to send question progress response")
	}
}

// ResetTopic godoc
// @Summary Reset topic progress
// @Description Reset all progress data for a topic for the authenticated user
// @Tags Progress
// @Produce json
// @Security BearerAuth
// @Param id path string true "Topic ID"
// @Success 200
// @Failure 400 {object} render.Error "Invalid topic ID"
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 404 {object} render.Error "Progress not found"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /progress/topics/{id} [delete]
//
// ResetTopic handles DELETE /progress/topics/{id} requests.
// It resets all progress data for the specified topic for the authenticated user.
func (h *Handler) ResetTopic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		h.log.Warn("unauthorized access to reset topic progress")

		if err := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); err != nil {
			h.log.With(zap.Error(err)).Warn("failed to send unauthorized response")
		}
		return
	}

	topicID, ok := httputil.URLParamUUID(r, "id")
	if !ok {
		h.handleError(w, ErrInvalidTopicID, "reset topic progress")
		return
	}

	if err := h.service.ResetTopicProgress(ctx, userID, topicID); err != nil {
		h.handleError(w, err, "reset topic progress")
		return
	}

	if err := render.OK(w, nil); err != nil {
		h.log.With(zap.Error(err)).Warn("failed to send reset response")
	}
}

// processMarkAction encapsulates common logic for marking question progress.
func (h *Handler) processMarkAction(
	w http.ResponseWriter,
	r *http.Request,
	actionFn func(ctx context.Context, input progress.MarkQuestionInput) error,
	actionName string,
) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		h.log.Warnf("unauthorized access attempt for %s", actionName)

		if err := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); err != nil {
			h.log.With(zap.Error(err)).Warn("failed to send unauthorized response")
		}
		return
	}

	var req markRequest

	if httputil.DecodeAndValidate(r, w, h.log, &req, actionName) {
		return
	}

	input := progress.MarkQuestionInput{
		UserID:     userID,
		TopicID:    req.TopicID,
		QuestionID: req.QuestionID,
	}

	if err := actionFn(ctx, input); err != nil {
		h.handleError(w, err, actionName)
		return
	}

	if err := render.OK(w, nil); err != nil {
		h.log.With(zap.Error(err)).Warnf("failed to send %s response", actionName)
	}
}

// handleError maps service-level errors to HTTP responses with appropriate status codes and logging.
func (h *Handler) handleError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, progress.ErrInvalidProgressStatus):
		h.log.Warn("invalid progress status")

		if sendErr := render.Fail(w, http.StatusBadRequest, ErrInvalidProgressStatus); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}

	case errors.Is(err, progress.ErrQuestionProgressNotFound):
		h.log.Warn("question progress not found")

		if sendErr := render.Fail(w, http.StatusNotFound, ErrProgressNotFound); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send not found response")
		}

	case errors.Is(err, progress.ErrTopicProgressNotFound):
		h.log.Warn("topic progress not found")

		if sendErr := render.Fail(w, http.StatusNotFound, ErrProgressNotFound); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send not found response")
		}

	case errors.Is(err, ErrInvalidTopicID):
		h.log.Warn("invalid topic id")

		if sendErr := render.Fail(w, http.StatusBadRequest, ErrInvalidTopicID); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}

	default:
		h.log.With(zap.Error(err)).Errorf("%s service failed", action)

		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal error response")
		}
	}
}
