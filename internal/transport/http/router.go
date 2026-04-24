package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpswagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	_ "github.com/backforge-app/backforge/api/swagger"

	"github.com/backforge-app/backforge/internal/config"
	"github.com/backforge-app/backforge/internal/transport/http/handler/analytics"
	"github.com/backforge-app/backforge/internal/transport/http/handler/auth"
	"github.com/backforge-app/backforge/internal/transport/http/handler/progress"
	"github.com/backforge-app/backforge/internal/transport/http/handler/question"
	"github.com/backforge-app/backforge/internal/transport/http/handler/tag"
	"github.com/backforge-app/backforge/internal/transport/http/handler/topic"
	"github.com/backforge-app/backforge/internal/transport/http/handler/user"
	mw "github.com/backforge-app/backforge/internal/transport/http/middleware"
	"github.com/backforge-app/backforge/internal/transport/http/render"
)

// Handlers bundles all handler instances for easy router wiring.
type Handlers struct {
	Auth      *auth.Handler
	User      *user.Handler
	Question  *question.Handler
	Topic     *topic.Handler
	Tag       *tag.Handler
	Progress  *progress.Handler
	Analytics *analytics.Handler
}

// NewRouter creates a production-ready HTTP router.
func NewRouter(
	cfg *config.Config,
	log *zap.SugaredLogger,
	handlers Handlers,
	userSvc mw.UserRoleChecker,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware (RequestID, Logger, Recovery, etc.).
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(mw.Logger(log))
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))

	if cfg.RateLimit.Enabled {
		r.Use(mw.RateLimiter(
			log,
			rate.Limit(cfg.RateLimit.Global.Limit),
			cfg.RateLimit.Global.Burst,
			cfg.RateLimit.CleanupInterval,
		))
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.Client.URL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		// Auth Routes (Public).
		r.Route("/auth", func(r chi.Router) {
			// Apply stricter rate limiting to auth endpoints to prevent brute-force attacks.
			if cfg.RateLimit.Enabled {
				r.Use(mw.RateLimiter(
					log,
					rate.Limit(cfg.RateLimit.Auth.Limit),
					cfg.RateLimit.Auth.Burst,
					cfg.RateLimit.CleanupInterval,
				))
			}

			r.Post("/register", handlers.Auth.Register)
			r.Post("/login", handlers.Auth.Login)
			r.Post("/verify-email", handlers.Auth.VerifyEmail)
			r.Post("/forgot-password", handlers.Auth.RequestPasswordReset)
			r.Post("/reset-password", handlers.Auth.ResetPassword)
			r.Post("/yandex/callback", handlers.Auth.YandexCallback)
			r.Post("/refresh", handlers.Auth.Refresh)
			r.Post("/resend-verification", handlers.Auth.ResendVerification)
		})

		// Public Discovery Routes (No Auth Required).
		// These allow unauthorized users to browse the catalog.
		r.Route("/questions", func(r chi.Router) {
			r.Get("/", handlers.Question.ListCards)
			r.Get("/{slug}", handlers.Question.GetBySlug)
		})

		r.Route("/topics", func(r chi.Router) {
			r.Get("/", handlers.Topic.ListRows)
			r.Get("/slug/{slug}", handlers.Topic.GetBySlug)
			r.Get("/{id}/questions", handlers.Question.ListByTopic)
		})

		r.Route("/tags", func(r chi.Router) {
			r.Get("/", handlers.Tag.List)
		})

		// Protected Routes (Auth Required).
		protected := chi.NewRouter()
		protected.Use(mw.Auth(cfg.Auth.Secret, log))

		// User profile.
		protected.Route("/users", func(r chi.Router) {
			r.Get("/me", handlers.User.GetProfile)
			r.Patch("/me", handlers.User.UpdateProfile)
		})

		// Question & Topic exploration.
		protected.Route("/questions", func(r chi.Router) {
			r.Get("/{id}", handlers.Question.GetByID)
		})
		protected.Route("/topics", func(r chi.Router) {
			r.Get("/{id}", handlers.Topic.GetByID)
		})

		// Progress tracking.
		protected.Route("/progress", func(r chi.Router) {
			r.Post("/known", handlers.Progress.MarkKnown)
			r.Post("/learned", handlers.Progress.MarkLearned)
			r.Post("/skipped", handlers.Progress.MarkSkipped)

			r.Get("/topics/{id}", handlers.Progress.GetTopicProgress)
			r.Delete("/topics/{id}", handlers.Progress.ResetTopic)
			r.Get("/questions/{id}", handlers.Progress.GetQuestionProgress)
		})

		// Analytics & Dashboard.
		protected.Route("/analytics", func(r chi.Router) {
			r.Get("/overall", handlers.Analytics.GetOverallProgress)
			r.Get("/topics", handlers.Analytics.GetProgressByTopicPercent)
			r.Delete("/reset", handlers.Analytics.ResetAllProgress)
		})

		// Admin routes.
		protected.Route("/admin", func(r chi.Router) {
			r.Use(mw.AdminOnly(log, userSvc))

			r.Route("/questions", func(r chi.Router) {
				r.Post("/", handlers.Question.Create)
				r.Get("/{id}", handlers.Question.GetByID)
				r.Put("/{id}", handlers.Question.Update)
			})
			r.Route("/topics", func(r chi.Router) {
				r.Post("/", handlers.Topic.Create)
				r.Put("/{id}", handlers.Topic.Update)
			})
			r.Route("/tags", func(r chi.Router) {
				r.Post("/", handlers.Tag.Create)
				r.Delete("/{id}", handlers.Tag.Delete)
			})
		})

		r.Mount("/", protected)
	})

	// Swagger UI.
	r.Get("/swagger/*", httpswagger.Handler(
		httpswagger.URL("/swagger/doc.json"),
	))

	// Health check.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := render.JSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		}); err != nil {
			log.Warnw("failed to write health response", "error", err)
		}
	})

	return r
}
