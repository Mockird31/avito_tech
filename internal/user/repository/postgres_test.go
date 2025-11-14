package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/user"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, user.IRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewRepository(db)
	return db, mock, repo
}

func getTestContext() context.Context {
	logger := zap.NewNop()
	ctx := context.Background()
	return loggerPkg.LoggerToContext(ctx, logger.Sugar())
}

func TestGetExistentUsers_Success(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	members := []string{"u1", "u2", "u3"}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow("u1").
		AddRow("u3")

	mock.ExpectQuery(regexp.QuoteMeta(GetExistentUsersQuery)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	m, err := repo.GetExistentUsers(ctx, members)
	require.NoError(t, err)

	assert.Contains(t, m, "u1")
	assert.Contains(t, m, "u3")
	assert.NotContains(t, m, "u2")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetExistentUsers_EmptyResult(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	members := []string{"u1", "u2"}

	rows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery(regexp.QuoteMeta(GetExistentUsersQuery)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	m, err := repo.GetExistentUsers(ctx, members)
	require.NoError(t, err)
	assert.Empty(t, m)

	require.NoError(t, mock.ExpectationsWereMet())
}

func toDriverValues(args []any) []driver.Value {
	values := make([]driver.Value, len(args))
	for i, v := range args {
		values[i] = v
	}
	return values
}

func TestCreateUsers_Success(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "teamA"
	users := []*entity.TeamMember{
		{UserID: "u1", Username: "alice", IsActive: true},
		{UserID: "u2", Username: "bob", IsActive: false},
	}

	expectedQuery, expectedArgs, err := PrepareCreateUsersQuery(users, teamName)
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta(expectedQuery)).
		WithArgs(toDriverValues(expectedArgs)...).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err = repo.CreateUsers(ctx, users, teamName)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUsers_DBError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "teamB"
	users := []*entity.TeamMember{
		{UserID: "u1", Username: "alice", IsActive: true},
	}

	expectedQuery, expectedArgs, err := PrepareCreateUsersQuery(users, teamName)
	require.NoError(t, err)

	dbErr := errors.New("insert failed")
	mock.ExpectExec(regexp.QuoteMeta(expectedQuery)).
		WithArgs(toDriverValues(expectedArgs)...).
		WillReturnError(dbErr)

	err = repo.CreateUsers(ctx, users, teamName)
	require.Error(t, err)
	assert.EqualError(t, err, dbErr.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}
