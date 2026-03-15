// Package topic implements HTTP handlers for topic management.
package topic

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/backforge-app/backforge/internal/domain"
	servicetopic "github.com/backforge-app/backforge/internal/service/topic"
	"github.com/backforge-app/backforge/internal/transport/http/httputil"
	"github.com/backforge-app/backforge/internal/transport/http/render"
)

// Handler handles topic-related HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates a new topic Handler with the provided service and logger.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{service: service, log: log}
}

// Create handles POST /topics requests.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req createRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "create topic") {
		return
	}

	var description string
	if req.Description != nil {
		description = *req.Description
	}

	input := servicetopic.CreateInput{
		Title:       req.Title,
		Slug:        req.Slug,
		Description: description,
		CreatedBy:   req.CreatedBy,
	}

	id, err := h.service.Create(ctx, input)
	if err != nil {
		h.handleError(w, err, "create topic")
		return
	}

	resp := createResponse{ID: id}
	if sendErr := render.OK(w, resp); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send create response")
	}
}

// Update handles PUT /topics/{id} requests.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req updateRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "update topic") {
		return
	}

	id, ok := httputil.URLParamUUID(r, "id")
	if !ok {
		h.log.Warn("invalid topic ID in URL")
		if sendErr := render.Fail(w, http.StatusBadRequest, ErrTopicInvalidID); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}
		return
	}

	input := servicetopic.UpdateInput{
		ID:          id,
		Title:       req.Title,
		Slug:        req.Slug,
		Description: req.Description,
		UpdatedBy:   req.UpdatedBy,
	}

	err := h.service.Update(ctx, input)
	if err != nil {
		h.handleError(w, err, "update topic")
		return
	}

	if sendErr := render.OK(w, nil); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send update response")
	}
}

// GetByID handles GET /topics/{id} requests.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, ok := httputil.URLParamUUID(r, "id")
	if !ok {
		h.log.Warn("invalid topic ID in URL")
		if sendErr := render.Fail(w, http.StatusBadRequest, ErrTopicInvalidID); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}
		return
	}

	t, err := h.service.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, servicetopic.ErrTopicNotFound):
			h.log.Warn("topic not found")
			if sendErr := render.Fail(w, http.StatusNotFound, ErrTopicNotFound); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send not found response")
			}

		default:
			h.log.With(zap.Error(err)).Error("get topic by id service failed")
			if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
			}
		}
		return
	}

	if sendErr := render.OK(w, toTopicResponse(t)); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send get topic response")
	}
}

// GetBySlug handles GET /topics/slug/{slug} requests.
func (h *Handler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slug := chi.URLParam(r, "slug")
	if slug == "" {
		h.log.Warn("invalid slug in URL")
		if sendErr := render.FailMessage(w, http.StatusBadRequest, "invalid slug"); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}
		return
	}

	t, err := h.service.GetBySlug(ctx, slug)
	if err != nil {
		switch {
		case errors.Is(err, servicetopic.ErrTopicNotFound):
			h.log.Warn("topic not found")
			if sendErr := render.Fail(w, http.StatusNotFound, ErrTopicNotFound); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send not found response")
			}

		default:
			h.log.With(zap.Error(err)).Error("get topic by slug service failed")
			if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
			}
		}
		return
	}

	if sendErr := render.OK(w, toTopicResponse(t)); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send get topic response")
	}
}

// ListRows handles GET /topics requests.
// It retrieves a list of topics formatted as rows, including question counts.
func (h *Handler) ListRows(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := h.service.ListRows(ctx)
	if err != nil {
		h.log.With(zap.Error(err)).Error("list topic rows service failed")
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
		}
		return
	}

	resp := make([]topicRowResponse, len(rows))
	for i, row := range rows {
		resp[i] = topicRowResponse{
			ID:            row.ID,
			Title:         row.Title,
			Slug:          row.Slug,
			QuestionCount: row.QuestionCount,
		}
	}

	if sendErr := render.OK(w, resp); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send list topics response")
	}
}

// toTopicResponse converts a domain.Topic to a topicResponse DTO.
func toTopicResponse(t *domain.Topic) topicResponse {
	var desc *string
	if t.Description != "" {
		d := t.Description
		desc = &d
	}

	return topicResponse{
		ID:          t.ID,
		Title:       t.Title,
		Slug:        t.Slug,
		Description: desc,
		CreatedBy:   t.CreatedBy,
		UpdatedBy:   t.UpdatedBy,
	}
}

// handleError maps service-level errors to HTTP responses for topic operations.
func (h *Handler) handleError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, servicetopic.ErrTopicNotFound):
		h.log.Warn("topic not found")
		if sendErr := render.Fail(w, http.StatusNotFound, ErrTopicNotFound); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send not found response")
		}

	case errors.Is(err, servicetopic.ErrTopicAlreadyExists):
		h.log.Warn("topic already exists")
		if sendErr := render.Fail(w, http.StatusConflict, ErrTopicAlreadyExists); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send conflict response")
		}

	case errors.Is(err, servicetopic.ErrTopicInvalidData):
		h.log.Warn("invalid topic data")
		if sendErr := render.Fail(w, http.StatusBadRequest, ErrTopicInvalidData); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}

	default:
		h.log.With(zap.Error(err)).Errorf("%s service failed", action)
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
		}
	}
}
