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
	"github.com/backforge-app/backforge/internal/repository/oauthconnection"
)

type OAuthConnectionRepoTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgCont      *postgres.PostgresContainer
	pool        *pgxpool.Pool
	repo        *oauthconnection.Repository
	testUser1ID uuid.UUID
	testUser2ID uuid.UUID
}

func TestOAuthConnectionRepositoryIntegration(t *testing.T) {
	suite.Run(t, new(OAuthConnectionRepoTestSuite))
}

func (s *OAuthConnectionRepoTestSuite) SetupSuite() {
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
	s.repo = oauthconnection.NewRepository(s.pool)

	s.createBaseFixtures()
}

func (s *OAuthConnectionRepoTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.pgCont != nil {
		if err := s.pgCont.Terminate(s.ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}
}

func (s *OAuthConnectionRepoTestSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, `TRUNCATE TABLE oauth_connections CASCADE`)
	s.Require().NoError(err)
}

func (s *OAuthConnectionRepoTestSuite) TestCreateAndGetByProviderUserID() {
	conn := &domain.OAuthConnection{
		UserID:         s.testUser1ID,
		Provider:       domain.OAuthProviderYandex,
		ProviderUserID: "yandex_user_123",
	}

	err := s.repo.Create(s.ctx, conn)
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, conn.ID)
	s.NotZero(conn.CreatedAt)

	err = s.repo.Create(s.ctx, &domain.OAuthConnection{
		UserID:         s.testUser2ID,
		Provider:       domain.OAuthProviderYandex,
		ProviderUserID: "yandex_user_123",
	})
	s.Require().ErrorIs(err, oauthconnection.ErrDuplicateConnection)

	fetched, err := s.repo.GetByProviderUserID(s.ctx, domain.OAuthProviderYandex, "yandex_user_123")
	s.Require().NoError(err)
	s.Equal(conn.ID, fetched.ID)
	s.Equal(s.testUser1ID, fetched.UserID)

	_, err = s.repo.GetByProviderUserID(s.ctx, domain.OAuthProviderYandex, "non_existent_user")
	s.Require().ErrorIs(err, oauthconnection.ErrConnectionNotFound)
}

func (s *OAuthConnectionRepoTestSuite) TestGetByUserID() {
	conn1 := &domain.OAuthConnection{
		UserID:         s.testUser1ID,
		Provider:       domain.OAuthProviderYandex,
		ProviderUserID: "yandex_1",
	}
	s.Require().NoError(s.repo.Create(s.ctx, conn1))

	time.Sleep(10 * time.Millisecond)

	conn2 := &domain.OAuthConnection{
		UserID:         s.testUser1ID,
		Provider:       domain.OAuthProvider("vk"),
		ProviderUserID: "vk_1",
	}
	s.Require().NoError(s.repo.Create(s.ctx, conn2))

	conn3 := &domain.OAuthConnection{
		UserID:         s.testUser2ID,
		Provider:       domain.OAuthProviderYandex,
		ProviderUserID: "yandex_2",
	}
	s.Require().NoError(s.repo.Create(s.ctx, conn3))

	user1Conns, err := s.repo.GetByUserID(s.ctx, s.testUser1ID)
	s.Require().NoError(err)
	s.Require().Len(user1Conns, 2)

	s.Equal(conn2.ID, user1Conns[0].ID)
	s.Equal(conn1.ID, user1Conns[1].ID)

	user2Conns, err := s.repo.GetByUserID(s.ctx, s.testUser2ID)
	s.Require().NoError(err)
	s.Require().Len(user2Conns, 1)
	s.Equal(conn3.ID, user2Conns[0].ID)

	emptyConns, err := s.repo.GetByUserID(s.ctx, uuid.New())
	s.Require().NoError(err)
	s.Empty(emptyConns)
}

func (s *OAuthConnectionRepoTestSuite) TestDelete() {
	conn := &domain.OAuthConnection{
		UserID:         s.testUser1ID,
		Provider:       domain.OAuthProviderYandex,
		ProviderUserID: "yandex_123",
	}

	err := s.repo.Create(s.ctx, conn)
	s.Require().NoError(err)

	err = s.repo.Delete(s.ctx, s.testUser1ID, domain.OAuthProviderYandex)
	s.Require().NoError(err)

	_, err = s.repo.GetByProviderUserID(s.ctx, domain.OAuthProviderYandex, "yandex_123")
	s.Require().ErrorIs(err, oauthconnection.ErrConnectionNotFound)

	err = s.repo.Delete(s.ctx, s.testUser1ID, domain.OAuthProviderYandex)
	s.Require().ErrorIs(err, oauthconnection.ErrConnectionNotFound)
}

func (s *OAuthConnectionRepoTestSuite) applyMigrations(dbURL string) {
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

func (s *OAuthConnectionRepoTestSuite) createBaseFixtures() {
	s.testUser1ID = uuid.New()
	s.testUser2ID = uuid.New()

	q := `INSERT INTO users (id, email, username, first_name) VALUES ($1, $2, $3, $4)`

	_, err := s.pool.Exec(s.ctx, q, s.testUser1ID, "user1@example.com", "user1", "User One")
	s.Require().NoError(err)

	_, err = s.pool.Exec(s.ctx, q, s.testUser2ID, "user2@example.com", "user2", "User Two")
	s.Require().NoError(err)
}
