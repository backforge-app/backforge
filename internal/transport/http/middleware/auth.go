// Package middleware provides HTTP middleware components for the Backforge API.
// It includes utilities for logging, authentication, authorization, and rate limiting.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/backforge-app/backforge/internal/transport/http/render"
)

// ContextKey is a private type to prevent collisions in the context map.
type ContextKey string

var (
	// UserIDKey is the context key used to store/retrieve the authenticated user ID.
	UserIDKey ContextKey = "userID"

	// ErrUnauthorized is returned when the request lacks valid authentication credentials.
	ErrUnauthorized = errors.New("unauthorized: missing or invalid token")
)

// Auth returns a middleware that validates JWT access tokens for incoming HTTP requests.
//
// The middleware expects an Authorization header in the format:
//
//	Authorization: Bearer <JWT>
//
// It verifies the token signature using the provided secret, checks the expiration,
// extracts the user ID (UUID) from the "sub" claim, and injects it into the request context.
//
// Any request with a missing, malformed, expired, or invalid token will receive
// a 401 Unauthorized response.
func Auth(secret string, log *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Extract the Authorization header.
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warn("missing authorization header")
				renderFailUnauthorized(w, log)
				return
			}

			// 2. Expect the format "Bearer <JWT>".
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				log.Warn("invalid authorization header format")
				renderFailUnauthorized(w, log)
				return
			}

			tokenString := parts[1]

			// 3. Validate the token and extract the userID.
			userID, err := validateToken(tokenString, secret)
			if err != nil {
				log.With(zap.Error(err)).Warn("failed to validate JWT token")
				renderFailUnauthorized(w, log)
				return
			}

			// 4. Inject userID into the request context.
			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			// 5. Call the next handler with the updated context.
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateToken parses and validates a JWT token using the provided secret.
//
// It checks the token signature, expiration, and extracts the "sub" claim as the userID (UUID).
// Returns an error if the token is invalid or the "sub" claim is missing/incorrect.
func validateToken(tokenString, secret string) (uuid.UUID, error) {
	// Parse the JWT token.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid JWT token")
	}

	// Extract claims.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("failed to read JWT claims")
	}

	// Extract the "sub" claim as userID.
	subRaw, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, errors.New("JWT missing sub claim")
	}

	userID, err := uuid.Parse(subRaw)
	if err != nil {
		return uuid.Nil, errors.New("invalid userID in sub claim")
	}

	return userID, nil
}

// UserIDFromContext retrieves the authenticated user's UUID from the request context.
//
// Returns uuid.Nil and false if the user ID is missing or has an incorrect type.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

// renderFailUnauthorized sends a 401 Unauthorized JSON response and logs the event.
func renderFailUnauthorized(w http.ResponseWriter, log *zap.SugaredLogger) {
	if err := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); err != nil {
		log.With(zap.Error(err)).Warn("failed to send unauthorized response")
	}
}
