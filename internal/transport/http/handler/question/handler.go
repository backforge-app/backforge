// Package question implements HTTP handlers for question management.
package question

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/backforge-app/backforge/internal/domain"
	servicequestion "github.com/backforge-app/backforge/internal/service/question"
	"github.com/backforge-app/backforge/internal/transport/http/httputil"
	"github.com/backforge-app/backforge/internal/transport/http/render"
)

// Handler handles question-related HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates a new question Handler.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{service: service, log: log}
}

// levelFromString converts string level to domain.QuestionLevel.
func levelFromString(s string) (domain.QuestionLevel, bool) {
	switch s {
	case "beginner":
		return domain.QuestionLevelBeginner, true
	case "medium":
		return domain.QuestionLevelMedium, true
	case "advanced":
		return domain.QuestionLevelAdvanced, true
	default:
		return domain.QuestionLevelBeginner, false
	}
}

// Create handles POST /questions requests.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req createRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "create question") {
		return
	}

	level, ok := levelFromString(req.Level)
	if !ok {
		h.log.Warn("invalid question level")
		if sendErr := render.Fail(w, http.StatusBadRequest, ErrQuestionInvalidData); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}
		return
	}

	input := servicequestion.CreateInput{
		Title:     req.Title,
		Slug:      req.Slug,
		Content:   req.Content,
		Level:     level,
		TopicID:   req.TopicID,
		IsFree:    req.IsFree,
		TagIDs:    req.TagIDs,
		CreatedBy: req.CreatedBy,
	}

	id, err := h.service.Create(ctx, input)
	if err != nil {
		h.handleError(w, err, "create question")
		return
	}

	resp := createResponse{ID: id}

	if sendErr := render.OK(w, resp); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send create response")
	}
}

// Update handles PUT /questions/{id} requests.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req updateRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "update question") {
		return
	}

	id, ok := httputil.URLParamUUID(r, "id")
	if !ok {
		h.log.Warn("invalid question ID")
		if sendErr := render.Fail(w, http.StatusBadRequest, ErrQuestionInvalidID); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}
		return
	}

	var level *domain.QuestionLevel

	if req.Level != nil {
		lvl, ok := levelFromString(*req.Level)
		if !ok {
			h.log.Warn("invalid question level")
			if sendErr := render.Fail(w, http.StatusBadRequest, ErrQuestionInvalidData); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
			}
			return
		}

		level = &lvl
	}

	input := servicequestion.UpdateInput{
		ID:        id,
		Title:     req.Title,
		Slug:      req.Slug,
		Content:   req.Content,
		Level:     level,
		TopicID:   req.TopicID,
		IsFree:    req.IsFree,
		TagIDs:    req.TagIDs,
		UpdatedBy: req.UpdatedBy,
	}

	err := h.service.Update(ctx, input)
	if err != nil {
		h.handleError(w, err, "update question")
		return
	}

	if sendErr := render.OK(w, nil); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send update response")
	}
}

// GetByID handles GET /admin/questions/{id} requests.
// It returns the full question including its content.
// This endpoint is intended for admin use.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, ok := httputil.URLParamUUID(r, "id")
	if !ok {
		h.log.Warn("invalid question ID in context")
		if sendErr := render.FailMessage(w, http.StatusBadRequest, "invalid question ID"); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}
		return
	}

	q, err := h.service.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, servicequestion.ErrQuestionNotFound):
			h.log.Warn("question not found")
			if sendErr := render.Fail(w, http.StatusNotFound, ErrQuestionNotFound); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send not found response")
			}

		default:
			h.log.With(zap.Error(err)).Error("get question service failed")
			if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
			}
		}

		return
	}

	if sendErr := render.OK(w, toResponse(q)); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send get question response")
	}
}

// GetBySlug handles GET /questions/{slug} requests.
// It retrieves a question by its slug.
func (h *Handler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slug := chi.URLParam(r, "slug")
	if slug == "" {
		h.log.Warn("invalid slug")
		if sendErr := render.FailMessage(w, http.StatusBadRequest, "invalid slug"); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}
		return
	}

	q, err := h.service.GetBySlug(ctx, slug)
	if err != nil {
		switch {
		case errors.Is(err, servicequestion.ErrQuestionNotFound):
			h.log.Warn("question not found")
			if sendErr := render.Fail(w, http.StatusNotFound, ErrQuestionNotFound); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send not found response")
			}

		default:
			h.log.With(zap.Error(err)).Error("get question service failed")
			if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
			}
		}
		return
	}

	if sendErr := render.OK(w, toResponse(q)); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send get question response")
	}
}

// ListCards handles GET /questions.
// Returns a list of questions with filtering and pagination via query params.
func (h *Handler) ListCards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := httputil.ParseQueryInt(r, "limit", 20)
	offset := httputil.ParseQueryInt(r, "offset", 0)
	search := httputil.GetQueryPtrString(r, "search")
	levelStr := httputil.GetQueryPtrString(r, "level")
	tagsParam := r.URL.Query().Get("tags")

	var tags []string
	if tagsParam != "" {
		for _, t := range strings.Split(tagsParam, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	var level *domain.QuestionLevel
	if levelStr != nil {
		lvl, ok := levelFromString(*levelStr)
		if !ok {
			h.log.Warn("invalid question level in query params")
			if sendErr := render.Fail(w, http.StatusBadRequest, ErrQuestionInvalidData); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
			}
			return
		}
		level = &lvl
	}

	input := servicequestion.ListInput{
		Limit:  limit,
		Offset: offset,
		Search: search,
		Level:  level,
		Tags:   tags,
	}

	cards, err := h.service.ListCards(ctx, input)
	if err != nil {
		h.log.With(zap.Error(err)).Error("list question cards service failed")
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
		}
		return
	}

	resp := make([]listCardResponse, len(cards))
	for i, q := range cards {
		resp[i] = listCardResponse{
			ID:     q.ID,
			Title:  q.Title,
			Slug:   q.Slug,
			Level:  q.Level.String(),
			Tags:   q.Tags,
			IsNew:  q.IsNew,
			IsFree: q.IsFree,
		}
	}

	if sendErr := render.OK(w, resp); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send list questions response")
	}
}

// ListByTopic handles GET /topics/{id}/questions.
// Returns all questions on the topic.
func (h *Handler) ListByTopic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	topicID, ok := httputil.URLParamUUID(r, "id")
	if !ok {
		h.log.Warn("invalid topic ID")
		if sendErr := render.FailMessage(w, http.StatusBadRequest, "invalid topic ID"); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}
		return
	}

	questions, err := h.service.ListByTopic(ctx, topicID)
	if err != nil {
		h.log.With(zap.Error(err)).Error("list questions by topic service failed")
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
		}
		return
	}

	resp := make([]questionResponse, len(questions))
	for i, q := range questions {
		resp[i] = toResponse(q)
	}

	if sendErr := render.OK(w, resp); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send list questions response")
	}
}

// toResponse converts a domain.Question to questionResponse DTO.
func toResponse(q *domain.Question) questionResponse {
	tagIDs := make([]uuid.UUID, len(q.Tags))
	for i, t := range q.Tags {
		tagIDs[i] = t.ID
	}

	return questionResponse{
		ID:        q.ID,
		Title:     q.Title,
		Slug:      q.Slug,
		Content:   q.Content,
		Level:     q.Level.String(),
		TopicID:   q.TopicID,
		IsFree:    q.IsFree,
		TagIDs:    tagIDs,
		CreatedBy: q.CreatedBy,
		UpdatedBy: q.UpdatedBy,
	}
}

// handleError maps service-level errors to HTTP responses for question operations.
func (h *Handler) handleError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, servicequestion.ErrQuestionNotFound):
		h.log.Warn("question not found")
		if sendErr := render.Fail(w, http.StatusNotFound, ErrQuestionNotFound); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send not found response")
		}

	case errors.Is(err, servicequestion.ErrQuestionAlreadyExists):
		h.log.Warn("question already exists")
		if sendErr := render.Fail(w, http.StatusConflict, ErrQuestionAlreadyExists); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send conflict response")
		}

	case errors.Is(err, servicequestion.ErrQuestionInvalidData):
		h.log.Warn("invalid question data")
		if sendErr := render.Fail(w, http.StatusBadRequest, ErrQuestionInvalidData); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}

	default:
		h.log.With(zap.Error(err)).Errorf("%s service failed", action)
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
		}
	}
}
