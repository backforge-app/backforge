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
	"github.com/backforge-app/backforge/internal/repository/topic"
)

type TopicRepoTestSuite struct {
	suite.Suite
	ctx        context.Context
	pgCont     *postgres.PostgresContainer
	pool       *pgxpool.Pool
	repo       *topic.Repository
	testUserID uuid.UUID
}

func TestTopicRepositoryIntegration(t *testing.T) {
	suite.Run(t, new(TopicRepoTestSuite))
}

func (s *TopicRepoTestSuite) SetupSuite() {
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
	s.repo = topic.NewRepository(s.pool)

	s.createBaseFixtures()
}

func (s *TopicRepoTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if err := s.pgCont.Terminate(s.ctx); err != nil {
		log.Fatalf("failed to terminate container: %v", err)
	}
}

func (s *TopicRepoTestSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, `TRUNCATE TABLE topics CASCADE`)
	s.Require().NoError(err)
}

func (s *TopicRepoTestSuite) TestCreateAndGet() {
	t := &domain.Topic{
		Title:       "Concurrency in Go",
		Slug:        "go-concurrency",
		Description: "Learn about goroutines and channels",
		CreatedBy:   &s.testUserID,
		UpdatedBy:   &s.testUserID,
	}

	id, err := s.repo.Create(s.ctx, t)
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, id)

	_, err = s.repo.Create(s.ctx, t)
	s.Require().ErrorIs(err, topic.ErrTopicAlreadyExists)

	fetched, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal(t.Title, fetched.Title)
	s.Equal(t.Slug, fetched.Slug)
	s.Equal(t.Description, fetched.Description)
	s.Equal(*t.CreatedBy, *fetched.CreatedBy)

	fetchedSlug, err := s.repo.GetBySlug(s.ctx, t.Slug)
	s.Require().NoError(err)
	s.Equal(id, fetchedSlug.ID)

	_, err = s.repo.GetByID(s.ctx, uuid.New())
	s.Require().ErrorIs(err, topic.ErrTopicNotFound)
}

func (s *TopicRepoTestSuite) TestUpdate() {
	id, err := s.repo.Create(s.ctx, &domain.Topic{
		Title:     "Old Title",
		Slug:      "old-slug",
		CreatedBy: &s.testUserID,
		UpdatedBy: &s.testUserID,
	})
	s.Require().NoError(err)

	updatedTopic := &domain.Topic{
		ID:          id,
		Title:       "New Title",
		Slug:        "new-slug",
		Description: "Updated description",
		UpdatedBy:   &s.testUserID,
	}

	err = s.repo.Update(s.ctx, updatedTopic)
	s.Require().NoError(err)

	fetched, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal("New Title", fetched.Title)
	s.Equal("new-slug", fetched.Slug)
	s.Equal("Updated description", fetched.Description)
}

func (s *TopicRepoTestSuite) TestListRows() {
	t1ID, err := s.repo.Create(s.ctx, &domain.Topic{
		Title:     "Topic 1",
		Slug:      "topic-1",
		CreatedBy: &s.testUserID,
		UpdatedBy: &s.testUserID,
	})
	s.Require().NoError(err)

	t2ID, err := s.repo.Create(s.ctx, &domain.Topic{
		Title:     "Topic 2",
		Slug:      "topic-2",
		CreatedBy: &s.testUserID,
		UpdatedBy: &s.testUserID,
	})
	s.Require().NoError(err)

	const insertQuestion = `
		INSERT INTO questions (title, slug, content, level, topic_id, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = s.pool.Exec(s.ctx, insertQuestion, "Q1", "q-1", "{}", 0, t1ID, s.testUserID, s.testUserID)
	s.Require().NoError(err)
	_, err = s.pool.Exec(s.ctx, insertQuestion, "Q2", "q-2", "{}", 0, t1ID, s.testUserID, s.testUserID)
	s.Require().NoError(err)

	rows, err := s.repo.ListRows(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(rows, 2)

	var row1, row2 *domain.TopicRow
	for _, r := range rows {
		if r.ID == t1ID {
			row1 = r
		} else if r.ID == t2ID {
			row2 = r
		}
	}

	s.Require().NotNil(row1)
	s.Equal(2, row1.QuestionCount)
	s.Equal("Topic 1", row1.Title)

	s.Require().NotNil(row2)
	s.Equal(0, row2.QuestionCount)
	s.Equal("Topic 2", row2.Title)
}

func (s *TopicRepoTestSuite) applyMigrations(dbURL string) {
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

func (s *TopicRepoTestSuite) createBaseFixtures() {
	s.testUserID = uuid.New()
	_, err := s.pool.Exec(s.ctx, `
		INSERT INTO users (id, telegram_id, username, first_name) 
		VALUES ($1, 12345, 'tester', 'Tester')
	`, s.testUserID)
	s.Require().NoError(err)
}
