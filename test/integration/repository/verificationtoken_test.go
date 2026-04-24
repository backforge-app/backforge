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
	"github.com/backforge-app/backforge/internal/repository/verificationtoken"
)

type VerificationTokenRepoTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgCont      *postgres.PostgresContainer
	pool        *pgxpool.Pool
	repo        *verificationtoken.Repository
	testUser1ID uuid.UUID
	testUser2ID uuid.UUID
}

func TestVerificationTokenRepositoryIntegration(t *testing.T) {
	suite.Run(t, new(VerificationTokenRepoTestSuite))
}

func (s *VerificationTokenRepoTestSuite) SetupSuite() {
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
	s.repo = verificationtoken.NewRepository(s.pool)

	s.createBaseFixtures()
}

func (s *VerificationTokenRepoTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.pgCont != nil {
		if err := s.pgCont.Terminate(s.ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}
}

func (s *VerificationTokenRepoTestSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, `TRUNCATE TABLE verification_tokens CASCADE`)
	s.Require().NoError(err)
}

func (s *VerificationTokenRepoTestSuite) TestCreateAndGetByHash() {
	validTokenHash := "hash_valid_123"

	token := &domain.VerificationToken{
		TokenHash: validTokenHash,
		UserID:    s.testUser1ID,
		Purpose:   domain.TokenPurposeEmailVerification,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	err := s.repo.Create(s.ctx, token)
	s.Require().NoError(err)
	s.NotZero(token.CreatedAt)

	fetched, err := s.repo.GetByHash(s.ctx, validTokenHash, domain.TokenPurposeEmailVerification)
	s.Require().NoError(err)
	s.Equal(token.UserID, fetched.UserID)
	s.Equal(token.Purpose, fetched.Purpose)

	s.WithinDuration(token.ExpiresAt, fetched.ExpiresAt, time.Second)

	_, err = s.repo.GetByHash(s.ctx, validTokenHash, domain.TokenPurposePasswordReset)
	s.Require().ErrorIs(err, verificationtoken.ErrTokenNotFound)

	expiredHash := "hash_expired_456"
	expiredToken := &domain.VerificationToken{
		TokenHash: expiredHash,
		UserID:    s.testUser1ID,
		Purpose:   domain.TokenPurposeEmailVerification,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Истек час назад
	}
	s.Require().NoError(s.repo.Create(s.ctx, expiredToken))

	_, err = s.repo.GetByHash(s.ctx, expiredHash, domain.TokenPurposeEmailVerification)
	s.Require().ErrorIs(err, verificationtoken.ErrTokenNotFound)
}

func (s *VerificationTokenRepoTestSuite) TestDelete() {
	tokenHash := "hash_to_delete_789"
	token := &domain.VerificationToken{
		TokenHash: tokenHash,
		UserID:    s.testUser1ID,
		Purpose:   domain.TokenPurposePasswordReset,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	s.Require().NoError(s.repo.Create(s.ctx, token))

	err := s.repo.Delete(s.ctx, tokenHash)
	s.Require().NoError(err)

	_, err = s.repo.GetByHash(s.ctx, tokenHash, domain.TokenPurposePasswordReset)
	s.Require().ErrorIs(err, verificationtoken.ErrTokenNotFound)

	err = s.repo.Delete(s.ctx, tokenHash)
	s.Require().ErrorIs(err, verificationtoken.ErrTokenNotFound)
}

func (s *VerificationTokenRepoTestSuite) TestDeleteAllForUser() {
	s.Require().NoError(s.repo.Create(s.ctx, &domain.VerificationToken{
		TokenHash: "u1_email_1", UserID: s.testUser1ID, Purpose: domain.TokenPurposeEmailVerification, ExpiresAt: time.Now().Add(time.Hour),
	}))
	s.Require().NoError(s.repo.Create(s.ctx, &domain.VerificationToken{
		TokenHash: "u1_email_2", UserID: s.testUser1ID, Purpose: domain.TokenPurposeEmailVerification, ExpiresAt: time.Now().Add(time.Hour),
	}))
	s.Require().NoError(s.repo.Create(s.ctx, &domain.VerificationToken{
		TokenHash: "u1_reset_1", UserID: s.testUser1ID, Purpose: domain.TokenPurposePasswordReset, ExpiresAt: time.Now().Add(time.Hour),
	}))

	s.Require().NoError(s.repo.Create(s.ctx, &domain.VerificationToken{
		TokenHash: "u2_email_1", UserID: s.testUser2ID, Purpose: domain.TokenPurposeEmailVerification, ExpiresAt: time.Now().Add(time.Hour),
	}))

	err := s.repo.DeleteAllForUser(s.ctx, s.testUser1ID, domain.TokenPurposeEmailVerification)
	s.Require().NoError(err)

	_, err = s.repo.GetByHash(s.ctx, "u1_email_1", domain.TokenPurposeEmailVerification)
	s.Require().ErrorIs(err, verificationtoken.ErrTokenNotFound)

	_, err = s.repo.GetByHash(s.ctx, "u1_email_2", domain.TokenPurposeEmailVerification)
	s.Require().ErrorIs(err, verificationtoken.ErrTokenNotFound)

	_, err = s.repo.GetByHash(s.ctx, "u1_reset_1", domain.TokenPurposePasswordReset)
	s.Require().NoError(err)

	_, err = s.repo.GetByHash(s.ctx, "u2_email_1", domain.TokenPurposeEmailVerification)
	s.Require().NoError(err)
}

func (s *VerificationTokenRepoTestSuite) applyMigrations(dbURL string) {
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

func (s *VerificationTokenRepoTestSuite) createBaseFixtures() {
	s.testUser1ID = uuid.New()
	s.testUser2ID = uuid.New()

	q := `INSERT INTO users (id, email, username, first_name) VALUES ($1, $2, $3, $4)`

	_, err := s.pool.Exec(s.ctx, q, s.testUser1ID, "user1@example.com", "user1", "User One")
	s.Require().NoError(err)

	_, err = s.pool.Exec(s.ctx, q, s.testUser2ID, "user2@example.com", "user2", "User Two")
	s.Require().NoError(err)
}
