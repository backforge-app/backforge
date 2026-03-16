//go:build integration
// +build integration

package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/repository/progress"
)

type ProgressRepoTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgCont      *postgres.PostgresContainer
	pool        *pgxpool.Pool
	qRepo       *progress.UserQuestionRepository
	tRepo       *progress.UserTopicRepository
	testUserID  uuid.UUID
	testTopicID uuid.UUID
	testQID     uuid.UUID
}

func TestProgressRepositoryIntegration(t *testing.T) {
	suite.Run(t, new(ProgressRepoTestSuite))
}

func (s *ProgressRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()

	pgContainer, err := postgres.Run(s.ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("backforge_test"),
		postgres.WithUsername("tester"),
		postgres.WithPassword("testerpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(15*time.Second),
		),
	)
	s.Require().NoError(err)
	s.pgCont = pgContainer

	connString, err := pgContainer.ConnectionString(s.ctx, "sslmode=disable")
	s.Require().NoError(err)

	s.applyMigrations(connString)

	pool, err := pgxpool.New(s.ctx, connString)
	s.Require().NoError(err)
	s.pool = pool

	s.qRepo = progress.NewUserQuestionRepository(s.pool)
	s.tRepo = progress.NewUserTopicRepository(s.pool)

	s.createBaseFixtures()
}

func (s *ProgressRepoTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	_ = s.pgCont.Terminate(s.ctx)
}

func (s *ProgressRepoTestSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, `TRUNCATE TABLE user_question_progress, user_topic_progress CASCADE`)
	s.Require().NoError(err)
}

func (s *ProgressRepoTestSuite) TestUserQuestion_SetAndGet() {
	err := s.qRepo.SetStatus(s.ctx, s.testUserID, s.testQID, domain.ProgressStatusLearned)
	s.Require().NoError(err)

	p, err := s.qRepo.GetByUserAndQuestion(s.ctx, s.testUserID, s.testQID)
	s.Require().NoError(err)
	s.Equal(domain.ProgressStatusLearned, p.Status)

	err = s.qRepo.SetStatus(s.ctx, s.testUserID, s.testQID, domain.ProgressStatusSkipped)
	s.Require().NoError(err)

	pUpdated, _ := s.qRepo.GetByUserAndQuestion(s.ctx, s.testUserID, s.testQID)
	s.Equal(domain.ProgressStatusSkipped, pUpdated.Status)
}

func (s *ProgressRepoTestSuite) TestUserQuestion_ListAndReset() {
	_ = s.qRepo.SetStatus(s.ctx, s.testUserID, s.testQID, domain.ProgressStatusKnown)

	list, err := s.qRepo.ListByUserAndTopic(s.ctx, s.testUserID, s.testTopicID)
	s.Require().NoError(err)
	s.Len(list, 1)

	err = s.qRepo.ResetByTopic(s.ctx, s.testUserID, s.testTopicID)
	s.Require().NoError(err)

	listEmpty, _ := s.qRepo.ListByUserAndTopic(s.ctx, s.testUserID, s.testTopicID)
	s.Empty(listEmpty)
}

func (s *ProgressRepoTestSuite) TestUserTopic_Position() {
	err := s.tRepo.SetPosition(s.ctx, s.testUserID, s.testTopicID, 10)
	s.Require().NoError(err)

	p, err := s.tRepo.GetByUserAndTopic(s.ctx, s.testUserID, s.testTopicID)
	s.Require().NoError(err)
	s.Equal(10, p.CurrentPosition)

	err = s.tRepo.SetPosition(s.ctx, s.testUserID, s.testTopicID, 20)
	s.Require().NoError(err)

	pUpdated, _ := s.tRepo.GetByUserAndTopic(s.ctx, s.testUserID, s.testTopicID)
	s.Equal(20, pUpdated.CurrentPosition)
}

func (s *ProgressRepoTestSuite) TestUserTopic_Reset() {
	_ = s.tRepo.SetPosition(s.ctx, s.testUserID, s.testTopicID, 50)

	err := s.tRepo.ResetByTopic(s.ctx, s.testUserID, s.testTopicID)
	s.Require().NoError(err)

	p, _ := s.tRepo.GetByUserAndTopic(s.ctx, s.testUserID, s.testTopicID)
	s.Equal(0, p.CurrentPosition)
}

func (s *ProgressRepoTestSuite) applyMigrations(dbURL string) {
	db, err := sql.Open("pgx", dbURL)
	s.Require().NoError(err)
	defer db.Close()
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "../../..")
	migrationsPath := filepath.Join(projectRoot, "migrations")
	goose.SetDialect("postgres")
	_ = goose.Up(db, migrationsPath)
}

func (s *ProgressRepoTestSuite) createBaseFixtures() {
	s.testUserID = uuid.New()
	s.testTopicID = uuid.New()
	s.testQID = uuid.New()

	_, _ = s.pool.Exec(s.ctx, `INSERT INTO users (id, telegram_id, first_name) VALUES ($1, 888, 'ProgressUser')`, s.testUserID)
	_, _ = s.pool.Exec(s.ctx, `INSERT INTO topics (id, title, slug) VALUES ($1, 'T', 't')`, s.testTopicID)
	_, _ = s.pool.Exec(s.ctx, `
		INSERT INTO questions (id, title, slug, content, level, topic_id, created_by, updated_by)
		VALUES ($1, 'Q', 'q', '{}', 0, $2, $3, $3)
	`, s.testQID, s.testTopicID, s.testUserID)
}
