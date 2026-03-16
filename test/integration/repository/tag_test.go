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
	"github.com/backforge-app/backforge/internal/repository/tag"
)

type TagRepoTestSuite struct {
	suite.Suite
	ctx        context.Context
	pgCont     *postgres.PostgresContainer
	pool       *pgxpool.Pool
	repo       *tag.Repository
	testUserID uuid.UUID
}

func TestTagRepositoryIntegration(t *testing.T) {
	suite.Run(t, new(TagRepoTestSuite))
}

func (s *TagRepoTestSuite) SetupSuite() {
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
	s.repo = tag.NewRepository(s.pool)

	s.createBaseFixtures()
}

func (s *TagRepoTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if err := s.pgCont.Terminate(s.ctx); err != nil {
		log.Fatalf("failed to terminate container: %v", err)
	}
}

func (s *TagRepoTestSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, `TRUNCATE TABLE tags CASCADE`)
	s.Require().NoError(err)
}

func (s *TagRepoTestSuite) TestCreateAndGet() {
	tagName := "Golang"
	t := &domain.Tag{
		Name:      tagName,
		CreatedBy: &s.testUserID,
		UpdatedBy: &s.testUserID,
	}

	id, err := s.repo.Create(s.ctx, t)
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, id)

	_, err = s.repo.Create(s.ctx, t)
	s.Require().ErrorIs(err, tag.ErrTagAlreadyExists)

	fetched, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal(tagName, fetched.Name)
	s.Equal(s.testUserID, *fetched.CreatedBy)

	fetchedByName, err := s.repo.GetByName(s.ctx, tagName)
	s.Require().NoError(err)
	s.Equal(id, fetchedByName.ID)

	_, err = s.repo.GetByID(s.ctx, uuid.New())
	s.Require().ErrorIs(err, tag.ErrTagNotFound)
}

func (s *TagRepoTestSuite) TestDelete() {
	id, err := s.repo.Create(s.ctx, &domain.Tag{
		Name:      "ToDelete",
		CreatedBy: &s.testUserID,
	})
	s.Require().NoError(err)

	err = s.repo.Delete(s.ctx, id)
	s.Require().NoError(err)

	_, err = s.repo.GetByID(s.ctx, id)
	s.Require().ErrorIs(err, tag.ErrTagNotFound)

	err = s.repo.Delete(s.ctx, uuid.New())
	s.Require().ErrorIs(err, tag.ErrTagNotFound)
}

func (s *TagRepoTestSuite) TestList() {
	names := []string{"Postgres", "Docker", "Algorithms"}
	for _, name := range names {
		_, err := s.repo.Create(s.ctx, &domain.Tag{
			Name:      name,
			CreatedBy: &s.testUserID,
		})
		s.Require().NoError(err)
	}

	tags, err := s.repo.List(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(tags, 3)

	s.Equal("Algorithms", tags[0].Name)
	s.Equal("Docker", tags[1].Name)
	s.Equal("Postgres", tags[2].Name)
}

func (s *TagRepoTestSuite) applyMigrations(dbURL string) {
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

func (s *TagRepoTestSuite) createBaseFixtures() {
	s.testUserID = uuid.New()
	_, err := s.pool.Exec(s.ctx, `
		INSERT INTO users (id, telegram_id, username, first_name) 
		VALUES ($1, 555444, 'tag_admin', 'TagAdmin')
	`, s.testUserID)
	s.Require().NoError(err)
}
