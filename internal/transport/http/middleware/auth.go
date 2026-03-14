// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/backforge-app/backforge/internal/transport/http/render"
)

// ContextKey is a private type to prevent collisions in the context map.
type ContextKey struct{}

var (
	// UserIDKey is the key used to store/retrieve the User ID from the request context.
	UserIDKey = ContextKey{}
)

var (
	// ErrUnauthorized indicates the request lacks valid authentication credentials.
	ErrUnauthorized = errors.New("unauthorized: missing or invalid token")
)

// Auth returns a middleware that validates authentication for incoming HTTP requests.
// It expects a Bearer token containing a valid UUID representing the user.
// If the token is valid, the user ID is injected into the request context.
func Auth(log *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Extract the Authorization header.
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warn("missing authorization header")
				if err := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); err != nil {
					log.With(zap.Error(err)).Warn("failed to send unauthorized response")
				}
				return
			}

			// 2. Expect format "Bearer <UUID>" and safely split only once.
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				log.With(zap.String("authorization", authHeader)).Warn("invalid authorization header format")
				if err := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); err != nil {
					log.With(zap.Error(err)).Warn("failed to send unauthorized response")
				}
				return
			}

			// 3. Parse the user ID as UUID.
			userID, parseErr := uuid.Parse(parts[1])
			if parseErr != nil {
				log.With(
					zap.String("authorization", authHeader),
					zap.Error(parseErr),
				).Warn("failed to parse user id from token")
				if err := render.Fail(w, http.StatusUnauthorized, ErrUnauthorized); err != nil {
					log.With(zap.Error(err)).Warn("failed to send unauthorized response")
				}
				return
			}

			// 4. Inject User ID into context.
			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			// 5. Call next handler with new context.
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext safely retrieves the User ID from the context.
// It returns uuid.Nil and false if the ID is not found or has the wrong type.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}
