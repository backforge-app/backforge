// Package auth implements HTTP handlers for user authentication.
package auth

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	serviceauth "github.com/backforge-app/backforge/internal/service/auth"
	"github.com/backforge-app/backforge/internal/transport/http/httputil"
	"github.com/backforge-app/backforge/internal/transport/http/render"
)

// Handler handles authentication HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates a new authentication Handler with the provided service and logger.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{service: service, log: log}
}

// Login godoc
// @Summary Login using Telegram
// @Description Login endpoint for Telegram auth
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body loginRequest true "Login payload"
// @Success 200 {object} loginResponse
// @Failure 400 {object} render.Error "Bad Request"
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 500 {object} render.Error "Internal Server Error"
// @Router /auth/login [post]
//
// Login handles POST /login requests using Telegram login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req loginRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "login") {
		return
	}

	input := serviceauth.TelegramLoginInput{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		PhotoURL:  req.PhotoURL,
		AuthDate:  req.AuthDate,
		Hash:      req.Hash,
	}

	accessToken, refreshToken, err := h.service.LoginWithTelegram(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, serviceauth.ErrInvalidTelegramAuth):
			h.log.Warn("invalid telegram auth data")
			if sendErr := render.Fail(w, http.StatusUnauthorized, ErrInvalidCredentials); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send unauthorized response")
			}

		case errors.Is(err, serviceauth.ErrTelegramAuthExpired):
			h.log.Warn("telegram auth expired")
			if sendErr := render.Fail(w, http.StatusUnauthorized, ErrInvalidCredentials); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send unauthorized response")
			}

		default:
			h.log.With(zap.Error(err)).
				Error("login service failed")
			if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
			}
		}

		return
	}

	resp := loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	if err := render.OK(w, resp); err != nil {
		h.log.With(zap.Error(err)).
			Warn("failed to send login response")
	}
}

// Refresh godoc
// @Summary Refresh access token
// @Description Exchange a refresh token for a new access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh body refreshRequest true "Refresh token payload"
// @Success 200 {object} refreshResponse "New access and refresh tokens"
// @Failure 401 {object} render.Error "Unauthorized"
// @Failure 500 {object} render.Error "Internal server error"
// @Router /api/v1/auth/refresh [post]
//
// Refresh handles POST /refresh requests.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req refreshRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "refresh") {
		return
	}

	accessToken, newRefreshToken, err := h.service.Refresh(ctx, req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, serviceauth.ErrRefreshTokenInvalid):
			h.log.Warn("invalid refresh token")
			if sendErr := render.Fail(w, http.StatusUnauthorized, ErrInvalidCredentials); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send unauthorized response")
			}

		case errors.Is(err, serviceauth.ErrRefreshTokenRevoked):
			h.log.Warn("revoked refresh token used")
			if sendErr := render.Fail(w, http.StatusUnauthorized, ErrInvalidCredentials); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send unauthorized response")
			}

		case errors.Is(err, serviceauth.ErrRefreshTokenAlreadyExists):
			h.log.With(zap.Error(err)).
				Error("refresh token collision")
			if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
			}

		default:
			h.log.With(zap.Error(err)).
				Error("refresh service failed")
			if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send internal server error response")
			}
		}

		return
	}

	resp := refreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}

	if err := render.OK(w, resp); err != nil {
		h.log.With(zap.Error(err)).
			Warn("failed to send refresh response")
	}
}
