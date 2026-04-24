// Package auth implements authentication, registration, and session management logic.
//
// It supports Email/Password login, GitHub OAuth, JWT access token issuance,
// refresh token rotation, email verification, and secure password reset flows.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/config"
	"github.com/backforge-app/backforge/internal/domain"
	oauthrepo "github.com/backforge-app/backforge/internal/repository/oauthconnection"
	sessionrepo "github.com/backforge-app/backforge/internal/repository/session"
	tokenrepo "github.com/backforge-app/backforge/internal/repository/verificationtoken"
	"github.com/backforge-app/backforge/internal/service/user"
)

// Service provides authentication and identity operations.
type Service struct {
	userService UserService
	sessionRepo SessionRepository
	oauthRepo   OAuthConnectionRepository
	tokenRepo   VerificationTokenRepository
	emailSender EmailSender
	oauthClient OAuthClient
	transactor  Transactor
	authConfig  *config.Auth
}

// NewService creates a new authentication service instance.
func NewService(
	userService UserService,
	sessionRepo SessionRepository,
	oauthRepo OAuthConnectionRepository,
	tokenRepo VerificationTokenRepository,
	emailSender EmailSender,
	oauthClient OAuthClient,
	transactor Transactor,
	authConfig *config.Auth,
) *Service {
	return &Service{
		userService: userService,
		sessionRepo: sessionRepo,
		oauthRepo:   oauthRepo,
		tokenRepo:   tokenRepo,
		emailSender: emailSender,
		oauthClient: oauthClient,
		transactor:  transactor,
		authConfig:  authConfig,
	}
}

// Register creates a new user, generates a verification token, and sends a welcome email.
func (s *Service) Register(ctx context.Context, input RegisterInput) error {
	var rawToken string
	var userEmail string

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		// 1. Create the user
		userID, err := s.userService.CreateWithPassword(txCtx, user.CreateWithPasswordInput{
			Email:     input.Email,
			Password:  input.Password,
			FirstName: input.FirstName,
			LastName:  input.LastName,
			Username:  input.Username,
		})
		if err != nil {
			return err // Will bubble up user.ErrUserEmailTaken, etc.
		}

		// 2. Generate verification token
		tokenStr, entity, err := domain.NewVerificationToken(userID, domain.TokenPurposeEmailVerification, s.authConfig.EmailVerificationTTL)
		if err != nil {
			return fmt.Errorf("generate token: %w", err)
		}

		// 3. Save token hash to DB
		if err := s.tokenRepo.Create(txCtx, entity); err != nil {
			return fmt.Errorf("save token: %w", err)
		}

		rawToken = tokenStr
		userEmail = input.Email
		return nil
	})

	if err != nil {
		return err
	}

	// 4. Send email (outside the DB transaction to avoid blocking)
	// In a high-load system, this should be published to a message queue (e.g., RabbitMQ).
	if err := s.emailSender.SendVerificationEmail(ctx, userEmail, rawToken); err != nil {
		// Log the error. We don't fail the registration, but the user will need to request a new link.
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}

// VerifyEmail validates the token and marks the user's email as verified.
func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	tokenHash := domain.HashVerificationToken(rawToken)

	return s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		token, err := s.tokenRepo.GetByHash(txCtx, tokenHash, domain.TokenPurposeEmailVerification)
		if err != nil {
			if errors.Is(err, tokenrepo.ErrTokenNotFound) {
				return ErrInvalidVerificationToken
			}
			return fmt.Errorf("get token: %w", err)
		}

		if err := s.userService.MarkEmailVerified(txCtx, token.UserID); err != nil {
			return fmt.Errorf("mark verified: %w", err)
		}

		// Invalidate the token so it can't be used again.
		if err := s.tokenRepo.Delete(txCtx, tokenHash); err != nil {
			return fmt.Errorf("delete used token: %w", err)
		}

		return nil
	})
}

// Login authenticates a user by email and password.
func (s *Service) Login(ctx context.Context, input LoginInput) (string, string, error) {
	u, err := s.userService.GetByEmail(ctx, input.Email)
	if err != nil {
		// Return a generic credential error to prevent email enumeration attacks.
		return "", "", ErrInvalidCredentials
	}

	if !u.CheckPassword(input.Password) {
		return "", "", ErrInvalidCredentials
	}

	if !u.IsEmailVerified {
		return "", "", ErrEmailNotVerified
	}

	return s.issueTokens(ctx, u)
}

// RequestPasswordReset generates a reset token and sends an email to the user.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	u, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		// Security: If user is not found, we silently return nil to prevent email enumeration.
		// A malicious actor shouldn't know if the email exists in our DB.
		return nil
	}

	var rawToken string

	err = s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		// Invalidate any previously requested, unused reset tokens
		if err := s.tokenRepo.DeleteAllForUser(txCtx, u.ID, domain.TokenPurposePasswordReset); err != nil {
			return fmt.Errorf("cleanup old tokens: %w", err)
		}

		tokenStr, entity, err := domain.NewVerificationToken(
			u.ID, domain.TokenPurposePasswordReset, s.authConfig.PasswordResetTTL,
		)
		if err != nil {
			return fmt.Errorf("generate token: %w", err)
		}

		if err := s.tokenRepo.Create(txCtx, entity); err != nil {
			return fmt.Errorf("save token: %w", err)
		}

		rawToken = tokenStr
		return nil
	})

	if err != nil {
		return err
	}

	if err := s.emailSender.SendPasswordResetEmail(ctx, u.Email, rawToken); err != nil {
		return fmt.Errorf("send reset email: %w", err)
	}

	return nil
}

// ResetPassword validates the reset token and updates the user's password.
func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	tokenHash := domain.HashVerificationToken(rawToken)

	return s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		token, err := s.tokenRepo.GetByHash(txCtx, tokenHash, domain.TokenPurposePasswordReset)
		if err != nil {
			if errors.Is(err, tokenrepo.ErrTokenNotFound) {
				return ErrInvalidVerificationToken
			}
			return fmt.Errorf("get token: %w", err)
		}

		if err := s.userService.SetNewPassword(txCtx, token.UserID, newPassword); err != nil {
			// Bubbles up domain limit errors (e.g., user.ErrPasswordTooLong).
			return err
		}

		if err := s.tokenRepo.Delete(txCtx, tokenHash); err != nil {
			return fmt.Errorf("delete used token: %w", err)
		}

		return nil
	})
}

// LoginWithYandex handles the OAuth flow, linking existing accounts by email
// or creating new ones, then issuing tokens.
func (s *Service) LoginWithYandex(ctx context.Context, code string) (string, string, error) {
	profile, err := s.oauthClient.ExchangeCode(ctx, code)
	if err != nil {
		return "", "", ErrOAuthExchangeFailed
	}

	var targetUser *domain.User

	err = s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		// 1. Check if we already have an OAuth connection for this Yandex ID.
		conn, err := s.oauthRepo.GetByProviderUserID(txCtx, domain.OAuthProviderYandex, profile.ProviderID)
		if err == nil {
			// User exists and is linked.
			targetUser, err = s.userService.GetByID(txCtx, conn.UserID)
			return err
		}
		if !errors.Is(err, oauthrepo.ErrConnectionNotFound) {
			return fmt.Errorf("query oauth connection: %w", err)
		}

		// 2. Connection not found. Check if an account with this email already exists.
		targetUser, err = s.userService.GetByEmail(txCtx, profile.Email)
		if err != nil {
			if !errors.Is(err, user.ErrUserNotFound) {
				return fmt.Errorf("query user by email: %w", err)
			}

			// 3. User does not exist. Create a new OAuth user.
			userID, createErr := s.userService.CreateOAuthUser(txCtx, user.CreateOAuthInput{
				Email:           profile.Email,
				FirstName:       profile.Name,
				PhotoURL:        &profile.AvatarURL,
				IsEmailVerified: true, // trusted from Yandex
			})
			if createErr != nil {
				return fmt.Errorf("create oauth user: %w", createErr)
			}

			targetUser, err = s.userService.GetByID(txCtx, userID)
			if err != nil {
				return fmt.Errorf("get newly created oauth user: %w", err)
			}
		}

		// 4. Link the Yandex account to the local user (new or existing)
		newConn := domain.NewOAuthConnection(targetUser.ID, domain.OAuthProviderYandex, profile.ProviderID)
		if err := s.oauthRepo.Create(txCtx, newConn); err != nil {
			return fmt.Errorf("link oauth connection: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", "", err
	}

	return s.issueTokens(ctx, targetUser)
}

// Refresh exchanges a valid refresh token for new tokens.
func (s *Service) Refresh(ctx context.Context, oldRawToken string) (string, string, error) {
	var newAccessToken, newRefreshToken string

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		tokenHash := hashToken(oldRawToken)

		session, err := s.sessionRepo.GetByTokenHash(txCtx, tokenHash)
		if err != nil {
			if errors.Is(err, sessionrepo.ErrSessionNotFound) {
				return ErrRefreshTokenInvalid
			}
			return fmt.Errorf("get refresh token: %w", err)
		}

		if session.Revoked {
			return ErrRefreshTokenRevoked
		}
		if time.Now().After(session.ExpiresAt) {
			return ErrRefreshTokenInvalid
		}

		u, err := s.userService.GetByID(txCtx, session.UserID)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		newAccessToken, err = s.generateAccessToken(u)
		if err != nil {
			return fmt.Errorf("generate access token: %w", err)
		}

		if err := s.sessionRepo.Revoke(txCtx, tokenHash); err != nil {
			return fmt.Errorf("revoke old refresh token: %w", err)
		}

		newRefreshToken, err = s.generateRefreshToken(txCtx, u.ID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// ResendVerificationEmail invalidates existing verification tokens and issues a new one.
func (s *Service) ResendVerificationEmail(ctx context.Context, email string) error {
	u, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		// Security: If user is not found, silently return nil to prevent email enumeration.
		return nil
	}

	if u.IsEmailVerified {
		// It's safe to return an error here so the frontend can redirect them to login.
		return ErrEmailAlreadyVerified
	}

	var rawToken string

	err = s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		// 1. Invalidate any previously requested, unused verification tokens
		if err := s.tokenRepo.DeleteAllForUser(txCtx, u.ID, domain.TokenPurposeEmailVerification); err != nil {
			return fmt.Errorf("cleanup old tokens: %w", err)
		}

		// 2. Generate new token
		tokenStr, entity, err := domain.NewVerificationToken(
			u.ID, domain.TokenPurposeEmailVerification, s.authConfig.EmailVerificationTTL,
		)
		if err != nil {
			return fmt.Errorf("generate token: %w", err)
		}

		// 3. Save new token
		if err := s.tokenRepo.Create(txCtx, entity); err != nil {
			return fmt.Errorf("save token: %w", err)
		}

		rawToken = tokenStr
		return nil
	})

	if err != nil {
		return err
	}

	// 4. Send the new email
	if err := s.emailSender.SendVerificationEmail(ctx, u.Email, rawToken); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}

// issueTokens is an internal helper to generate both Access and Refresh tokens.
func (s *Service) issueTokens(ctx context.Context, u *domain.User) (string, string, error) {
	accessToken, err := s.generateAccessToken(u)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(ctx, u.ID)
	if err != nil {
		return "", "", err // bubble up potential ErrRefreshTokenAlreadyExists
	}

	return accessToken, refreshToken, nil
}

// generateAccessToken creates a signed JWT access token.
func (s *Service) generateAccessToken(u *domain.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  u.ID.String(),
		"role": u.Role,
		"iss":  "backforge",
		"aud":  "backforge-client",
		"exp":  now.Add(s.authConfig.AccessTokenTTL).Unix(),
		"iat":  now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.authConfig.Secret))
}

// generateRefreshToken generates a secure opaque token and persists its hash.
func (s *Service) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("crypto rand read: %w", err)
	}

	rawToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(s.authConfig.RefreshTokenTTL)

	session := domain.NewSession(userID, tokenHash, expiresAt)
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		if errors.Is(err, sessionrepo.ErrSessionAlreadyExists) {
			return "", ErrRefreshTokenAlreadyExists
		}
		return "", fmt.Errorf("create session: %w", err)
	}

	return rawToken, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
