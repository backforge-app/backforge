// Package app provides the application wiring and startup logic.
// It assembles configuration, logger, repositories, services, HTTP handlers,
// and runs the HTTP server with graceful shutdown.
package app

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/backforge-app/backforge/internal/config"
	inframailer "github.com/backforge-app/backforge/internal/infra/mailer"
	infraoauth "github.com/backforge-app/backforge/internal/infra/oauth"
	"github.com/backforge-app/backforge/internal/infra/postgres"
	"github.com/backforge-app/backforge/internal/logger"
	repoanalytics "github.com/backforge-app/backforge/internal/repository/analytics"
	repooauth "github.com/backforge-app/backforge/internal/repository/oauthconnection"
	repoprogress "github.com/backforge-app/backforge/internal/repository/progress"
	repoquestion "github.com/backforge-app/backforge/internal/repository/question"
	repoquestiontag "github.com/backforge-app/backforge/internal/repository/questiontag"
	reposession "github.com/backforge-app/backforge/internal/repository/session"
	repotag "github.com/backforge-app/backforge/internal/repository/tag"
	repotopic "github.com/backforge-app/backforge/internal/repository/topic"
	repouser "github.com/backforge-app/backforge/internal/repository/user"
	repotoken "github.com/backforge-app/backforge/internal/repository/verificationtoken"
	svcanalytics "github.com/backforge-app/backforge/internal/service/analytics"
	svcauth "github.com/backforge-app/backforge/internal/service/auth"
	svcprogress "github.com/backforge-app/backforge/internal/service/progress"
	svcquestion "github.com/backforge-app/backforge/internal/service/question"
	svctag "github.com/backforge-app/backforge/internal/service/tag"
	svctopic "github.com/backforge-app/backforge/internal/service/topic"
	svcuser "github.com/backforge-app/backforge/internal/service/user"
	transporthttp "github.com/backforge-app/backforge/internal/transport/http"
	"github.com/backforge-app/backforge/internal/transport/http/handler/analytics"
	"github.com/backforge-app/backforge/internal/transport/http/handler/auth"
	"github.com/backforge-app/backforge/internal/transport/http/handler/progress"
	"github.com/backforge-app/backforge/internal/transport/http/handler/question"
	"github.com/backforge-app/backforge/internal/transport/http/handler/tag"
	"github.com/backforge-app/backforge/internal/transport/http/handler/topic"
	"github.com/backforge-app/backforge/internal/transport/http/handler/user"
	pkgmailer "github.com/backforge-app/backforge/pkg/mailer"
	"github.com/backforge-app/backforge/pkg/transactor"
	pkgyandex "github.com/backforge-app/backforge/pkg/yandex"
)

// App holds all application components.
type App struct {
	Config       *config.Config
	Logger       *zap.SugaredLogger
	Repositories *Repositories
	Services     *Services
	Server       *transporthttp.Server
}

// Repositories holds all repository instances.
type Repositories struct {
	User                 *repouser.Repository
	Question             *repoquestion.Repository
	Topic                *repotopic.Repository
	Tag                  *repotag.Repository
	Analytics            *repoanalytics.Repository
	UserQuestionProgress *repoprogress.UserQuestionRepository
	UserTopicProgress    *repoprogress.UserTopicRepository
	QuestionTag          *repoquestiontag.Repository
	Session              *reposession.Repository
	OAuthConnection      *repooauth.Repository
	VerificationToken    *repotoken.Repository
}

// Services holds all service instances.
type Services struct {
	User      *svcuser.Service
	Auth      *svcauth.Service
	Question  *svcquestion.Service
	Topic     *svctopic.Service
	Tag       *svctag.Service
	Analytics *svcanalytics.Service
	Progress  *svcprogress.Service
}

// New creates and wires all application dependencies.
// It does NOT start the HTTP server yet.
func New(ctx context.Context) (*App, error) {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// 2. Create logger
	log, err := logger.New(logger.Config{
		Environment: cfg.Env,
		Level:       cfg.Logging.Level,
	})
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	// 3. Create PostgreSQL pool
	pgPool, err := postgres.NewPool(ctx, cfg.Postgres.ConnectionURL, cfg.Postgres.Pool)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	// 4. Wrap pool with transactor
	tx := transactor.NewTransactor(postgres.NewPoolAdapter(pgPool), log)

	// 5. Setup generic external clients & infrastructure adapters
	mailerCfg := pkgmailer.Config{
		Host:        cfg.SMTP.Host,
		Port:        cfg.SMTP.Port,
		Username:    cfg.SMTP.Username,
		Password:    cfg.SMTP.Password,
		FromAddress: cfg.SMTP.FromAddress,
	}
	genericMailer := pkgmailer.New(mailerCfg)

	// AuthSender parses HTML templates. If it fails (e.g., bad syntax), the app won't start.
	authSender, err := inframailer.NewAuthSender(genericMailer, cfg.Client.URL)
	if err != nil {
		return nil, fmt.Errorf("initialize auth email sender: %w", err)
	}

	yandexCfg := pkgyandex.Config{
		ClientID:     cfg.OAuth.Yandex.ClientID,
		ClientSecret: cfg.OAuth.Yandex.ClientSecret,
	}
	genericYandexClient := pkgyandex.New(yandexCfg)
	oauthAdapter := infraoauth.NewYandexAdapter(genericYandexClient)

	// 6. Setup repositories
	repos := &Repositories{
		User:                 repouser.NewRepository(pgPool),
		Question:             repoquestion.NewRepository(pgPool),
		Topic:                repotopic.NewRepository(pgPool),
		Tag:                  repotag.NewRepository(pgPool),
		Analytics:            repoanalytics.NewRepository(pgPool),
		UserQuestionProgress: repoprogress.NewUserQuestionRepository(pgPool),
		UserTopicProgress:    repoprogress.NewUserTopicRepository(pgPool),
		QuestionTag:          repoquestiontag.NewRepository(pgPool),
		Session:              reposession.NewRepository(pgPool),
		OAuthConnection:      repooauth.NewRepository(pgPool),
		VerificationToken:    repotoken.NewRepository(pgPool),
	}

	// 7. Setup services
	userSvc := svcuser.NewService(repos.User, tx)

	svcs := &Services{
		User: userSvc,
		Auth: svcauth.NewService(
			userSvc,
			repos.Session,
			repos.OAuthConnection,
			repos.VerificationToken,
			authSender,
			oauthAdapter,
			tx,
			&cfg.Auth,
		),
		Question:  svcquestion.NewService(repos.Question, repos.QuestionTag, tx),
		Topic:     svctopic.NewService(repos.Topic, tx),
		Tag:       svctag.NewService(repos.Tag),
		Analytics: svcanalytics.NewService(repos.Analytics, repos.UserQuestionProgress, repos.UserTopicProgress),
		Progress:  svcprogress.NewService(repos.UserQuestionProgress, repos.UserTopicProgress),
	}

	// 8. Setup HTTP handlers
	handlers := transporthttp.Handlers{
		Auth:      auth.NewHandler(svcs.Auth, log),
		User:      user.NewHandler(svcs.User, log),
		Question:  question.NewHandler(svcs.Question, log),
		Topic:     topic.NewHandler(svcs.Topic, log),
		Tag:       tag.NewHandler(svcs.Tag, log),
		Progress:  progress.NewHandler(svcs.Progress, log),
		Analytics: analytics.NewHandler(svcs.Analytics, log),
	}

	// 9. Setup HTTP router
	router := transporthttp.NewRouter(cfg, log, handlers, userSvc)

	// 10. Create HTTP server
	server := transporthttp.NewServer(cfg, log, router)

	return &App{
		Config:       cfg,
		Logger:       log,
		Repositories: repos,
		Services:     svcs,
		Server:       server,
	}, nil
}

// Run starts the HTTP server and handles graceful shutdown.
func (a *App) Run(ctx context.Context) error {
	a.Logger.Infof("starting server on %s", a.Config.HTTP.Port)
	return a.Server.Run(ctx)
}
