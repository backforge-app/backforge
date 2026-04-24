// Package auth implements HTTP handlers for user authentication and identity management.
package auth

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	serviceauth "github.com/backforge-app/backforge/internal/service/auth"
	serviceuser "github.com/backforge-app/backforge/internal/service/user"
	"github.com/backforge-app/backforge/internal/transport/http/httputil"
	"github.com/backforge-app/backforge/internal/transport/http/render"
)

// Handler handles authentication HTTP requests.
type Handler struct {
	service Service
	log     *zap.SugaredLogger
}

// NewHandler creates a new authentication Handler.
func NewHandler(service Service, log *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Registers a new user and sends an email verification link
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "Registration payload"
// @Success 201 {object} render.Message "Registration successful. Please check your email."
// @Failure 400 {object} render.Error "Validation failed"
// @Failure 409 {object} render.Error "Email or Username already taken"
// @Failure 500 {object} render.Error "Internal Server Error"
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "register") {
		return
	}

	input := serviceauth.RegisterInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
	}

	if err := h.service.Register(r.Context(), input); err != nil {
		switch {
		case errors.Is(err, serviceuser.ErrUserEmailTaken):
			if sendErr := render.Fail(w, http.StatusConflict, ErrEmailTaken); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send email conflict response")
			}
		case errors.Is(err, serviceuser.ErrUserUsernameTaken):
			if sendErr := render.Fail(w, http.StatusConflict, ErrUsernameTaken); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send username conflict response")
			}
		default:
			h.log.With(zap.Error(err)).Error("register service failed")
			if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send internal error response")
			}
		}
		return
	}

	if sendErr := render.Created(w, render.Message{Message: "Registration successful. Please check your email to verify your account."}); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send created response")
	}
}

// Login godoc
// @Summary Login using email and password
// @Description Authenticates a user and returns access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login payload"
// @Success 200 {object} tokenResponse
// @Failure 400 {object} render.Error "Validation failed"
// @Failure 401 {object} render.Error "Invalid credentials or unverified email"
// @Failure 500 {object} render.Error "Internal Server Error"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "login") {
		return
	}

	input := serviceauth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	access, refresh, err := h.service.Login(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, serviceauth.ErrInvalidCredentials):
			if sendErr := render.Fail(w, http.StatusUnauthorized, ErrInvalidCredentials); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send unauthorized response")
			}
		case errors.Is(err, serviceauth.ErrEmailNotVerified):
			if sendErr := render.Fail(w, http.StatusUnauthorized, ErrEmailNotVerified); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send email not verified response")
			}
		default:
			h.log.With(zap.Error(err)).Error("login service failed")
			if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send internal error response")
			}
		}
		return
	}

	if sendErr := render.OK(w, tokenResponse{AccessToken: access, RefreshToken: refresh}); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send login ok response")
	}
}

// VerifyEmail godoc
// @Summary Verify user email
// @Description Verifies a user's email using the token sent to their inbox
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body verifyEmailRequest true "Verification token"
// @Success 200 {object} render.Message "Email verified successfully"
// @Failure 400 {object} render.Error "Invalid or expired token"
// @Router /auth/verify-email [post]
//
//nolint:dupl // Handlers share standard validation and response boilerplate
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "verify email") {
		return
	}

	if err := h.service.VerifyEmail(r.Context(), req.Token); err != nil {
		if errors.Is(err, serviceauth.ErrInvalidVerificationToken) {
			if sendErr := render.Fail(w, http.StatusBadRequest, ErrInvalidToken); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
			}
			return
		}
		h.log.With(zap.Error(err)).Error("verify email service failed")
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal error response")
		}
		return
	}

	if sendErr := render.Msg(w, "email verified successfully"); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send verification ok response")
	}
}

// RequestPasswordReset godoc
// @Summary Request password reset email
// @Description Sends a password reset link to the provided email if it exists
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body requestResetRequest true "Email payload"
// @Success 200 {object} render.Message "If the email is registered, a reset link has been sent"
// @Router /auth/forgot-password [post]
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req requestResetRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "request reset") {
		return
	}

	if err := h.service.RequestPasswordReset(r.Context(), req.Email); err != nil {
		h.log.With(zap.Error(err)).Error("request password reset service failed")
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal error response")
		}
		return
	}

	// Always return success to prevent email enumeration.
	if sendErr := render.Msg(w, "if that email is registered, a password reset link has been sent"); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send reset requested response")
	}
}

// ResetPassword godoc
// @Summary Set a new password
// @Description Sets a new password using a valid reset token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body resetPasswordRequest true "Token and new password payload"
// @Success 200 {object} render.Message "Password reset successfully"
// @Failure 400 {object} render.Error "Invalid or expired token"
// @Router /auth/reset-password [post]
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "reset password") {
		return
	}

	if err := h.service.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		if errors.Is(err, serviceauth.ErrInvalidVerificationToken) {
			if sendErr := render.Fail(w, http.StatusBadRequest, ErrInvalidToken); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
			}
			return
		}
		h.log.With(zap.Error(err)).Error("reset password service failed")
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal error response")
		}
		return
	}

	if sendErr := render.Msg(w, "password reset successfully"); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send reset success response")
	}
}

// YandexCallback godoc
// @Summary Yandex OAuth Callback
// @Description Exchanges Yandex authorization code for JWT tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body yandexCallbackRequest true "Yandex code payload"
// @Success 200 {object} tokenResponse
// @Failure 400 {object} render.Error "Invalid code or OAuth failure"
// @Router /auth/yandex/callback [post]
func (h *Handler) YandexCallback(w http.ResponseWriter, r *http.Request) {
	var req yandexCallbackRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "yandex callback") {
		return
	}

	access, refresh, err := h.service.LoginWithYandex(r.Context(), req.Code)
	if err != nil {
		h.log.With(zap.Error(err)).Warn("yandex oauth failed")
		if sendErr := render.Fail(w, http.StatusBadRequest, ErrOAuthFailed); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send bad request response")
		}
		return
	}

	if sendErr := render.OK(w, tokenResponse{AccessToken: access, RefreshToken: refresh}); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send oauth ok response")
	}
}

// Refresh godoc
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new pair of access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body refreshRequest true "Refresh token payload"
// @Success 200 {object} tokenResponse
// @Failure 401 {object} render.Error "Invalid or revoked token"
// @Router /auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "refresh") {
		return
	}

	access, refresh, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, serviceauth.ErrRefreshTokenInvalid), errors.Is(err, serviceauth.ErrRefreshTokenRevoked):
			h.log.Warn("invalid or revoked refresh token")
			if sendErr := render.Fail(w, http.StatusUnauthorized, ErrInvalidCredentials); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send unauthorized response")
			}
		default:
			h.log.With(zap.Error(err)).Error("refresh token service failed")
			if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send internal error response")
			}
		}
		return
	}

	if sendErr := render.OK(w, tokenResponse{AccessToken: access, RefreshToken: refresh}); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send refresh ok response")
	}
}

// ResendVerification godoc
// @Summary Resend verification email
// @Description Sends a new email verification link if the account exists and is not yet verified
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body resendVerificationRequest true "Email payload"
// @Success 200 {object} render.Message "If the account exists and needs verification, an email has been sent"
// @Failure 400 {object} render.Error "Email is already verified"
// @Failure 500 {object} render.Error "Internal Server Error"
// @Router /auth/resend-verification [post]
//
//nolint:dupl // Handlers share standard validation and response boilerplate
func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req resendVerificationRequest
	if httputil.DecodeAndValidate(r, w, h.log, &req, "resend verification") {
		return
	}

	if err := h.service.ResendVerificationEmail(r.Context(), req.Email); err != nil {
		if errors.Is(err, serviceauth.ErrEmailAlreadyVerified) {
			if sendErr := render.Fail(w, http.StatusBadRequest, ErrAlreadyVerified); sendErr != nil {
				h.log.With(zap.Error(sendErr)).Warn("failed to send already verified response")
			}
			return
		}

		h.log.With(zap.Error(err)).Error("resend verification service failed")
		if sendErr := render.FailMessage(w, http.StatusInternalServerError, ErrInternalServer.Error()); sendErr != nil {
			h.log.With(zap.Error(sendErr)).Warn("failed to send internal error response")
		}
		return
	}

	// Blind success response to prevent email enumeration (similar to forgot-password)
	if sendErr := render.Msg(w, "if the account exists and is unverified, a new link has been sent"); sendErr != nil {
		h.log.With(zap.Error(sendErr)).Warn("failed to send resend success response")
	}
}
