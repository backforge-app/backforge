// Package user implements HTTP handlers for user management.
package user

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	serviceuser "github.com/backforge-app/backforge/internal/service/user"
	"github.com/backforge-app/backforge/internal/transport/http/middleware"
	"github.com/backforge-app/backforge/internal/transport/http/render"
)

// Handler handles user-related HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates a new user Handler.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Returns the authenticated user's profile
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} userResponse
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 404 {object} render.Error "User not found"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /users/me [get]
//
// GetProfile handles GET /users/me requests.
// It returns the authenticated user's profile based on the ID in the context.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		h.log.Warn("unauthorized access attempt to profile")
		if sendErr := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send unauthorized response")
		}
		return
	}

	u, err := h.service.GetByID(ctx, userID)
	if err != nil {
		h.handleError(w, err, "get profile")
		return
	}

	if sendErr := render.OK(w, toUserResponse(u)); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send profile response")
	}
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Updates the authenticated user's mutable fields (e.g., name, username, photo)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body updateProfileRequest true "Profile update payload"
// @Success 200 {object} render.Message "Profile updated successfully"
// @Failure 400 {object} render.Error "Validation failed or invalid JSON"
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 409 {object} render.Error "Username already taken"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /users/me [patch]
//
// UpdateProfile handles PATCH /users/me requests.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		h.log.Warn("unauthorized access attempt to update profile")
		if sendErr := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send unauthorized response")
		}
		return
	}

	var req updateProfileRequest
	if err := render.Decode(r, &req); err != nil {
		// If Decode fails due to validation rules, render.ValidationErrors will extract them.
		validationErrs := render.ValidationErrors(err)
		if len(validationErrs) > 0 {
			if sendErr := render.FailWithDetails(w, http.StatusBadRequest, "validation failed", validationErrs); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send validation error response")
			}
			return
		}

		h.log.With(zap.Error(err)).Warn("failed to decode update profile request")
		if sendErr := render.FailMessage(w, http.StatusBadRequest, "invalid request body"); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}
		return
	}

	input := serviceuser.UpdateInput{
		ID:        userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		PhotoURL:  req.PhotoURL,
	}

	if err := h.service.Update(ctx, input); err != nil {
		h.handleError(w, err, "update profile")
		return
	}

	if sendErr := render.Msg(w, "profile updated successfully"); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send profile update response")
	}
}

// handleError maps service-level errors to appropriate HTTP responses.
func (h *Handler) handleError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, serviceuser.ErrUserNotFound):
		h.log.Warn("user not found")
		if sendErr := render.Fail(w, http.StatusNotFound, ErrUserNotFound); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send not found response")
		}

	case errors.Is(err, serviceuser.ErrUserUsernameTaken):
		h.log.Warn("username conflict during update")
		if sendErr := render.Fail(w, http.StatusConflict, ErrUsernameTaken); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send conflict response")
		}

	default:
		h.log.With(zap.Error(err)).Errorf("%s service failed", action)
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
		}
	}
}
