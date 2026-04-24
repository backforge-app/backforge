//go:build integration
// +build integration

package repository

import (
	"context"
	"database/sql"
	"log"
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
	"github.com/backforge-app/backforge/internal/repository/questiontag"
)

type QuestionTagRepoTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgCont      *postgres.PostgresContainer
	pool        *pgxpool.Pool
	repo        *questiontag.Repository
	testUserID  uuid.UUID
	testTopicID uuid.UUID
	testQID     uuid.UUID
	testTagIDs  []uuid.UUID
}

func TestQuestionTagRepositoryIntegration(t *testing.T) {
	suite.Run(t, new(QuestionTagRepoTestSuite))
}

func (s *QuestionTagRepoTestSuite) SetupSuite() {
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
	s.repo = questiontag.NewRepository(s.pool)

	s.createBaseFixtures()
}

func (s *QuestionTagRepoTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.pgCont != nil {
		if err := s.pgCont.Terminate(s.ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}
}

func (s *QuestionTagRepoTestSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, `TRUNCATE TABLE question_tags CASCADE`)
	s.Require().NoError(err)
}

func (s *QuestionTagRepoTestSuite) TestAddAndListTags() {
	err := s.repo.AddTagsToQuestion(s.ctx, s.testQID, s.testTagIDs)
	s.Require().NoError(err)

	tags, err := s.repo.ListTagsForQuestion(s.ctx, s.testQID)
	s.Require().NoError(err)
	s.Require().Len(tags, 2)
	s.Equal("Go", tags[0].Name)
	s.Equal("SQL", tags[1].Name)

	err = s.repo.AddTagsToQuestion(s.ctx, s.testQID, s.testTagIDs)
	s.Require().NoError(err)
}

func (s *QuestionTagRepoTestSuite) TestRemoveSpecificTags() {
	_ = s.repo.AddTagsToQuestion(s.ctx, s.testQID, s.testTagIDs)

	err := s.repo.RemoveTagsFromQuestion(s.ctx, s.testQID, []uuid.UUID{s.testTagIDs[0]})
	s.Require().NoError(err)

	tags, err := s.repo.ListTagsForQuestion(s.ctx, s.testQID)
	s.Require().NoError(err)
	s.Len(tags, 1)
}

func (s *QuestionTagRepoTestSuite) TestRemoveAllForQuestion() {
	_ = s.repo.AddTagsToQuestion(s.ctx, s.testQID, s.testTagIDs)

	err := s.repo.RemoveAllForQuestion(s.ctx, s.testQID)
	s.Require().NoError(err)

	tags, err := s.repo.ListTagsForQuestion(s.ctx, s.testQID)
	s.Require().NoError(err)
	s.Empty(tags)
}

func (s *QuestionTagRepoTestSuite) applyMigrations(dbURL string) {
	db, err := sql.Open("pgx", dbURL)
	s.Require().NoError(err)
	defer db.Close()

	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "../../..")
	migrationsPath := filepath.Join(projectRoot, "migrations")

	err = goose.SetDialect("postgres")
	s.Require().NoError(err)

	err = goose.Up(db, migrationsPath)
	s.Require().NoError(err)
}

func (s *QuestionTagRepoTestSuite) createBaseFixtures() {
	s.testUserID = uuid.New()
	s.testTopicID = uuid.New()
	s.testQID = uuid.New()

	_, err := s.pool.Exec(s.ctx, `
		INSERT INTO users (id, email, first_name) 
		VALUES ($1, 'qtag_test@example.com', 'Admin')
	`, s.testUserID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(s.ctx, `INSERT INTO topics (id, title, slug) VALUES ($1, 'T', 't')`, s.testTopicID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(s.ctx, `
		INSERT INTO questions (id, title, slug, content, level, topic_id, created_by, updated_by)
		VALUES ($1, 'Q', 'q', '{}', $2, $3, $4, $4)
	`, s.testQID, domain.QuestionLevelBeginner, s.testTopicID, s.testUserID)
	s.Require().NoError(err)

	s.testTagIDs = []uuid.UUID{uuid.New(), uuid.New()}
	_, err = s.pool.Exec(s.ctx, `
		INSERT INTO tags (id, name, created_by) 
		VALUES ($1, 'SQL', $3), ($2, 'Go', $3)
	`, s.testTagIDs[0], s.testTagIDs[1], s.testUserID)
	s.Require().NoError(err)
}
