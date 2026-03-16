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
	"github.com/backforge-app/backforge/internal/repository/session"
)

type SessionRepoTestSuite struct {
	suite.Suite
	ctx        context.Context
	pgCont     *postgres.PostgresContainer
	pool       *pgxpool.Pool
	repo       *session.Repository
	testUserID uuid.UUID
}

func TestSessionRepositoryIntegration(t *testing.T) {
	suite.Run(t, new(SessionRepoTestSuite))
}

func (s *SessionRepoTestSuite) SetupSuite() {
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
	s.repo = session.NewRepository(s.pool)

	s.createBaseFixtures()
}

func (s *SessionRepoTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if err := s.pgCont.Terminate(s.ctx); err != nil {
		log.Fatalf("failed to terminate container: %v", err)
	}
}

func (s *SessionRepoTestSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, `TRUNCATE TABLE sessions CASCADE`)
	s.Require().NoError(err)
}

func (s *SessionRepoTestSuite) TestCreateAndGet() {
	tokenHash := "some_very_long_secure_hash"
	expiresAt := time.Now().Add(24 * time.Hour).UTC()

	sess := &domain.Session{
		UserID:    s.testUserID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	err := s.repo.Create(s.ctx, sess)
	s.Require().NoError(err)

	err = s.repo.Create(s.ctx, sess)
	s.Require().ErrorIs(err, session.ErrSessionAlreadyExists)

	fetched, err := s.repo.GetByTokenHash(s.ctx, tokenHash)
	s.Require().NoError(err)
	s.Equal(s.testUserID, fetched.UserID)
	s.Equal(tokenHash, fetched.TokenHash)
	s.False(fetched.Revoked)
	s.WithinDuration(expiresAt, fetched.ExpiresAt, time.Second)

	_, err = s.repo.GetByTokenHash(s.ctx, "non_existent_hash")
	s.Require().ErrorIs(err, session.ErrSessionNotFound)
}

func (s *SessionRepoTestSuite) TestRevoke() {
	tokenHash := "revoke_me_hash"
	sess := domain.NewSession(s.testUserID, tokenHash, time.Now().Add(time.Hour))

	err := s.repo.Create(s.ctx, sess)
	s.Require().NoError(err)

	err = s.repo.Revoke(s.ctx, tokenHash)
	s.Require().NoError(err)

	fetched, err := s.repo.GetByTokenHash(s.ctx, tokenHash)
	s.Require().NoError(err)
	s.True(fetched.Revoked)
	s.WithinDuration(time.Now(), fetched.UpdatedAt, time.Second)

	err = s.repo.Revoke(s.ctx, "unknown_hash")
	s.Require().ErrorIs(err, session.ErrSessionNotFound)
}

func (s *SessionRepoTestSuite) applyMigrations(dbURL string) {
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

func (s *SessionRepoTestSuite) createBaseFixtures() {
	s.testUserID = uuid.New()
	_, err := s.pool.Exec(s.ctx, `
		INSERT INTO users (id, telegram_id, username, first_name) 
		VALUES ($1, 999999, 'session_owner', 'Owner')
	`, s.testUserID)
	s.Require().NoError(err)
}
