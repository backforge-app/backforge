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
	"github.com/backforge-app/backforge/internal/repository/user"
)

type UserRepoTestSuite struct {
	suite.Suite
	ctx    context.Context
	pgCont *postgres.PostgresContainer
	pool   *pgxpool.Pool
	repo   *user.Repository
}

func TestUserRepositoryIntegration(t *testing.T) {
	suite.Run(t, new(UserRepoTestSuite))
}

func (s *UserRepoTestSuite) SetupSuite() {
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
	s.repo = user.NewRepository(s.pool)
}

func (s *UserRepoTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if err := s.pgCont.Terminate(s.ctx); err != nil {
		log.Fatalf("failed to terminate container: %v", err)
	}
}

func (s *UserRepoTestSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, `TRUNCATE TABLE users CASCADE`)
	s.Require().NoError(err)
}

func (s *UserRepoTestSuite) TestCreateAndGet() {
	u := &domain.User{
		TelegramID: 111222333,
		FirstName:  "John",
		LastName:   ptr("Doe"),
		Username:   ptr("johndoe"),
		Role:       domain.UserRoleUser,
	}

	id, err := s.repo.Create(s.ctx, u)
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, id)

	fetchedByID, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal(id, fetchedByID.ID)
	s.Equal(u.TelegramID, fetchedByID.TelegramID)
	s.Equal(u.FirstName, fetchedByID.FirstName)

	s.Require().NotNil(fetchedByID.LastName)
	s.Equal(*u.LastName, *fetchedByID.LastName)

	s.Require().NotNil(fetchedByID.Username)
	s.Equal(*u.Username, *fetchedByID.Username)

	s.Equal(domain.UserRoleUser, fetchedByID.Role)

	fetchedByTG, err := s.repo.GetByTelegramID(s.ctx, u.TelegramID)
	s.Require().NoError(err)
	s.Equal(id, fetchedByTG.ID)
}

func (s *UserRepoTestSuite) TestUpdate() {
	u := &domain.User{
		TelegramID: 444555666,
		FirstName:  "OldName",
		Username:   ptr("old_nick"),
		Role:       domain.UserRoleUser,
	}

	id, err := s.repo.Create(s.ctx, u)
	s.Require().NoError(err)

	u.ID = id
	u.FirstName = "NewName"
	u.Username = ptr("new_nick")
	u.Role = domain.UserRoleAdmin

	err = s.repo.Update(s.ctx, u)
	s.Require().NoError(err)

	updated, err := s.repo.GetByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal("NewName", updated.FirstName)

	s.Require().NotNil(updated.Username)
	s.Equal("new_nick", *updated.Username)
	s.Equal(domain.UserRoleAdmin, updated.Role)
}

func (s *UserRepoTestSuite) TestIsAdmin() {
	adminID, err := s.repo.Create(s.ctx, &domain.User{
		TelegramID: 100,
		FirstName:  "Admin",
		Role:       domain.UserRoleAdmin,
	})
	s.Require().NoError(err)

	userID, err := s.repo.Create(s.ctx, &domain.User{
		TelegramID: 200,
		FirstName:  "User",
		Role:       domain.UserRoleUser,
	})
	s.Require().NoError(err)

	isAdmin, err := s.repo.IsAdmin(s.ctx, adminID)
	s.Require().NoError(err)
	s.True(isAdmin)

	isNotAdmin, err := s.repo.IsAdmin(s.ctx, userID)
	s.Require().NoError(err)
	s.False(isNotAdmin)
}

func (s *UserRepoTestSuite) applyMigrations(dbURL string) {
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

func ptr[T any](v T) *T {
	return &v
}
