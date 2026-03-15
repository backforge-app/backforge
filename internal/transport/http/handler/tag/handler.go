// Package tag implements HTTP handlers for tag management.
package tag

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	servicetag "github.com/backforge-app/backforge/internal/service/tag"
	"github.com/backforge-app/backforge/internal/transport/http/httputil"
	"github.com/backforge-app/backforge/internal/transport/http/render"
)

// Handler handles tag-related HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates a new tag Handler.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{service: service, log: log}
}

// Create handles POST /tags requests.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req createRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "create tag") {
		return
	}

	id, err := h.service.Create(ctx, req.Name, req.CreatedBy)
	if err != nil {
		h.handleError(w, err, "create tag")
		return
	}

	resp := createResponse{ID: id}
	if err := render.OK(w, resp); err != nil {
		h.log.With(zap.Error(err)).Warn("failed to send create response")
	}
}

// GetByID handles GET /tags/{id} requests.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, ok := httputil.URLParamUUID(r, "id")
	if !ok {
		h.log.Warn("invalid tag ID in URL")
		if err := render.Fail(w, http.StatusBadRequest, ErrTagInvalidID); err != nil {
			h.log.With(zap.Error(err)).Warn("failed to send bad request response")
		}
		return
	}

	t, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.handleError(w, err, "get tag")
		return
	}

	if err := render.OK(w, tagResponse{ID: t.ID, Name: t.Name}); err != nil {
		h.log.With(zap.Error(err)).Warn("failed to send get tag response")
	}
}

// List handles GET /tags requests.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tags, err := h.service.List(ctx)
	if err != nil {
		h.handleError(w, err, "list tags")
		return
	}

	resp := make([]tagResponse, len(tags))
	for i, t := range tags {
		resp[i] = tagResponse{ID: t.ID, Name: t.Name}
	}

	if err := render.OK(w, resp); err != nil {
		h.log.With(zap.Error(err)).Warn("failed to send list tags response")
	}
}

// Delete handles DELETE /tags/{id} requests.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, ok := httputil.URLParamUUID(r, "id")
	if !ok {
		h.log.Warn("invalid tag ID in URL")
		if err := render.Fail(w, http.StatusBadRequest, ErrTagInvalidID); err != nil {
			h.log.With(zap.Error(err)).Warn("failed to send bad request response")
		}
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.handleError(w, err, "delete tag")
		return
	}

	if err := render.OK(w, nil); err != nil {
		h.log.With(zap.Error(err)).Warn("failed to send delete response")
	}
}

// handleError maps service-level errors to HTTP responses for tag operations.
func (h *Handler) handleError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, servicetag.ErrTagNotFound):
		h.log.Warn("tag not found")
		if sendErr := render.Fail(w, http.StatusNotFound, ErrTagNotFound); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send not found response")
		}

	case errors.Is(err, servicetag.ErrTagAlreadyExists):
		h.log.Warn("tag already exists")
		if sendErr := render.Fail(w, http.StatusConflict, ErrTagAlreadyExists); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send conflict response")
		}

	case errors.Is(err, servicetag.ErrTagInvalidData):
		h.log.Warn("invalid tag data")
		if sendErr := render.Fail(w, http.StatusBadRequest, ErrTagInvalidData); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}

	default:
		h.log.With(zap.Error(err)).Errorf("%s service failed", action)
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
		}
	}
}
