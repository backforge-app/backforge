//go:build integration
// +build integration

package repository

import (
	"context"
	"database/sql"
	"fmt"
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
	"github.com/backforge-app/backforge/internal/repository/question"
)

type QuestionRepoTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgCont      *postgres.PostgresContainer
	pool        *pgxpool.Pool
	repo        *question.Repository
	testUserID  uuid.UUID
	testTopicID uuid.UUID
	testTagID   uuid.UUID
}

func TestQuestionRepositoryIntegration(t *testing.T) {
	suite.Run(t, new(QuestionRepoTestSuite))
}

func (s *QuestionRepoTestSuite) SetupSuite() {
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
	s.repo = question.NewRepository(s.pool)

	s.createBaseFixtures()
}

func (s *QuestionRepoTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.pgCont != nil {
		if err := s.pgCont.Terminate(s.ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}
}

func (s *QuestionRepoTestSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, `TRUNCATE TABLE questions, question_tags CASCADE`)
	s.Require().NoError(err)
}

func (s *QuestionRepoTestSuite) TestCreateAndGet() {
	q := &domain.Question{
		Title:     "What is a Goroutine?",
		Slug:      "what-is-goroutine",
		Content:   map[string]interface{}{"blocks": []string{"text block"}},
		Level:     domain.QuestionLevelMedium,
		TopicID:   &s.testTopicID,
		IsFree:    true,
		CreatedBy: &s.testUserID,
		UpdatedBy: &s.testUserID,
	}

	id, err := s.repo.Create(s.ctx, q)
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, id)

	_, err = s.repo.Create(s.ctx, q)
	s.Require().ErrorIs(err, question.ErrQuestionAlreadyExists)

	fetchedByID, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal(q.Title, fetchedByID.Title)
	s.Equal(q.Slug, fetchedByID.Slug)
	s.Equal(*q.TopicID, *fetchedByID.TopicID)
	s.Equal(domain.QuestionLevelMedium, fetchedByID.Level)
	s.True(fetchedByID.IsFree)

	fetchedBySlug, err := s.repo.GetBySlug(s.ctx, q.Slug)
	s.Require().NoError(err)
	s.Equal(id, fetchedBySlug.ID)

	_, err = s.repo.GetByID(s.ctx, uuid.New())
	s.Require().ErrorIs(err, question.ErrQuestionNotFound)
}

func (s *QuestionRepoTestSuite) TestUpdate() {
	id, err := s.repo.Create(s.ctx, &domain.Question{
		Title:     "Old Title",
		Slug:      "old-slug",
		Content:   map[string]interface{}{"test": "old"},
		Level:     domain.QuestionLevelBeginner,
		TopicID:   &s.testTopicID,
		IsFree:    true,
		CreatedBy: &s.testUserID,
		UpdatedBy: &s.testUserID,
	})
	s.Require().NoError(err)

	qUpdate := &domain.Question{
		ID:        id,
		Title:     "New Title",
		Slug:      "new-slug",
		Content:   map[string]interface{}{"test": "new"},
		Level:     domain.QuestionLevelAdvanced,
		TopicID:   &s.testTopicID,
		IsFree:    false,
		UpdatedBy: &s.testUserID,
	}

	err = s.repo.Update(s.ctx, qUpdate)
	s.Require().NoError(err)

	updated, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal("New Title", updated.Title)
	s.Equal("new-slug", updated.Slug)
	s.Equal(domain.QuestionLevelAdvanced, updated.Level)
	s.False(updated.IsFree)

	qUpdate.ID = uuid.New()
	err = s.repo.Update(s.ctx, qUpdate)
	s.Require().ErrorIs(err, question.ErrQuestionNotFound)
}

func (s *QuestionRepoTestSuite) TestListCards() {
	levels := []domain.QuestionLevel{
		domain.QuestionLevelBeginner,
		domain.QuestionLevelMedium,
		domain.QuestionLevelAdvanced,
	}

	for i, lvl := range levels {
		qID, err := s.repo.Create(s.ctx, &domain.Question{
			Title:     fmt.Sprintf("Question %d", i+1),
			Slug:      fmt.Sprintf("question-%d", i+1),
			Content:   map[string]interface{}{},
			Level:     lvl,
			TopicID:   &s.testTopicID,
			CreatedBy: &s.testUserID,
			UpdatedBy: &s.testUserID,
		})
		s.Require().NoError(err)

		if lvl == domain.QuestionLevelMedium {
			_, err = s.pool.Exec(s.ctx, `INSERT INTO question_tags (question_id, tag_id) VALUES ($1, $2)`, qID, s.testTagID)
			s.Require().NoError(err)
		}
	}

	cards, err := s.repo.ListCards(s.ctx, question.ListOptions{Limit: 10, Offset: 0})
	s.Require().NoError(err)
	s.Require().Len(cards, 3)

	mediumLvl := domain.QuestionLevelMedium
	cards, err = s.repo.ListCards(s.ctx, question.ListOptions{
		Limit:  10,
		Offset: 0,
		Level:  &mediumLvl,
	})
	s.Require().NoError(err)
	s.Require().Len(cards, 1)
	s.Equal(domain.QuestionLevelMedium, cards[0].Level)

	cards, err = s.repo.ListCards(s.ctx, question.ListOptions{
		Limit:  10,
		Offset: 0,
		Tags:   []string{"Golang"},
	})
	s.Require().NoError(err)
	s.Require().Len(cards, 1)
	s.Contains(cards[0].Tags, "Golang")
}

func (s *QuestionRepoTestSuite) TestListByTopic() {
	otherTopicID := uuid.New()
	_, err := s.pool.Exec(s.ctx, `INSERT INTO topics (id, title, slug) VALUES ($1, 'Other', 'other')`, otherTopicID)
	s.Require().NoError(err)

	_, err = s.repo.Create(s.ctx, &domain.Question{
		Title:     "Target Topic Question",
		Slug:      "target-topic-q",
		Content:   map[string]interface{}{},
		Level:     domain.QuestionLevelBeginner,
		TopicID:   &s.testTopicID,
		CreatedBy: &s.testUserID,
		UpdatedBy: &s.testUserID,
	})
	s.Require().NoError(err)

	questions, err := s.repo.ListByTopic(s.ctx, s.testTopicID)
	s.Require().NoError(err)
	s.Require().Len(questions, 1)
	s.Equal("Target Topic Question", questions[0].Title)
}

func (s *QuestionRepoTestSuite) applyMigrations(dbURL string) {
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

func (s *QuestionRepoTestSuite) createBaseFixtures() {
	s.testUserID = uuid.New()
	s.testTopicID = uuid.New()
	s.testTagID = uuid.New()

	_, err := s.pool.Exec(s.ctx, `
		INSERT INTO users (id, email, username, first_name) 
		VALUES ($1, 'johndoe@example.com', 'johndoe', 'John')
	`, s.testUserID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(s.ctx, `
		INSERT INTO topics (id, title, slug) 
		VALUES ($1, 'Golang Basics', 'golang-basics')
	`, s.testTopicID)
	s.Require().NoError(err)

	_, err = s.pool.Exec(s.ctx, `
		INSERT INTO tags (id, name, created_by, updated_by) 
		VALUES ($1, 'Golang', $2, $2)
	`, s.testTagID, s.testUserID)
	s.Require().NoError(err)
}
