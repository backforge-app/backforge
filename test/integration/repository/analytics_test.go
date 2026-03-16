//go:build integration
// +build integration

package repository

import (
	"context"
	"database/sql"
	"fmt"
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
	"github.com/backforge-app/backforge/internal/repository/analytics"
)

type AnalyticsRepoTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgCont      *postgres.PostgresContainer
	pool        *pgxpool.Pool
	repo        *analytics.Repository
	testUserID  uuid.UUID
	testTopicID uuid.UUID
}

func TestAnalyticsRepositoryIntegration(t *testing.T) {
	suite.Run(t, new(AnalyticsRepoTestSuite))
}

func (s *AnalyticsRepoTestSuite) SetupSuite() {
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
	s.repo = analytics.NewRepository(s.pool)

	s.createBaseFixtures()
}

func (s *AnalyticsRepoTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	_ = s.pgCont.Terminate(s.ctx)
}

func (s *AnalyticsRepoTestSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, `TRUNCATE TABLE user_question_progress CASCADE`)
	s.Require().NoError(err)
}

func (s *AnalyticsRepoTestSuite) TestGetOverallProgress() {
	p, err := s.repo.GetOverallProgress(s.ctx, s.testUserID)
	s.Require().NoError(err)
	s.Equal(3, p.Total)
	s.Equal(0, p.Known)

	s.setProgress(s.testUserID, s.getQuestionID(0), domain.ProgressStatusKnown)
	s.setProgress(s.testUserID, s.getQuestionID(1), domain.ProgressStatusLearned)

	pUpdated, err := s.repo.GetOverallProgress(s.ctx, s.testUserID)
	s.Require().NoError(err)
	s.Equal(3, pUpdated.Total)
	s.Equal(1, pUpdated.Known)
	s.Equal(1, pUpdated.Learned)
}

func (s *AnalyticsRepoTestSuite) TestGetTopicProgressPercent() {
	s.setProgress(s.testUserID, s.getQuestionID(0), domain.ProgressStatusKnown)

	results, err := s.repo.GetTopicProgressPercent(s.ctx, s.testUserID)
	s.Require().NoError(err)
	s.Require().Len(results, 1)

	res := results[0]
	s.Equal(s.testTopicID, res.TopicID)
	s.Equal(3, res.Total)
	s.Equal(1, res.Completed)
	s.InDelta(33.33, res.Percent, 0.01)
}

func (s *AnalyticsRepoTestSuite) setProgress(uID, qID uuid.UUID, status domain.ProgressStatus) {
	_, err := s.pool.Exec(s.ctx,
		`INSERT INTO user_question_progress (user_id, question_id, status) VALUES ($1, $2, $3)`,
		uID, qID, status)
	s.Require().NoError(err)
}

func (s *AnalyticsRepoTestSuite) getQuestionID(idx int) uuid.UUID {
	var ids []uuid.UUID
	rows, _ := s.pool.Query(s.ctx, "SELECT id FROM questions ORDER BY slug")
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids[idx]
}

func (s *AnalyticsRepoTestSuite) applyMigrations(dbURL string) {
	db, err := sql.Open("pgx", dbURL)
	s.Require().NoError(err)
	defer db.Close()
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "../../..")
	migrationsPath := filepath.Join(projectRoot, "migrations")
	goose.SetDialect("postgres")
	_ = goose.Up(db, migrationsPath)
}

func (s *AnalyticsRepoTestSuite) createBaseFixtures() {
	s.testUserID = uuid.New()
	s.testTopicID = uuid.New()

	_, _ = s.pool.Exec(s.ctx, `INSERT INTO users (id, telegram_id, first_name) VALUES ($1, 101, 'Analytic')`, s.testUserID)
	_, _ = s.pool.Exec(s.ctx, `INSERT INTO topics (id, title, slug) VALUES ($1, 'Go', 'go')`, s.testTopicID)

	for i := 1; i <= 3; i++ {
		_, _ = s.pool.Exec(s.ctx, `
			INSERT INTO questions (id, title, slug, content, level, topic_id, created_by, updated_by)
			VALUES ($1, $2, $3, '{}', 0, $4, $5, $5)
		`, uuid.New(), fmt.Sprintf("Q%d", i), fmt.Sprintf("q-%d", i), s.testTopicID, s.testUserID)
	}
}
