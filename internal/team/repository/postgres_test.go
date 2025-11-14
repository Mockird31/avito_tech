package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Mockird31/avito_tech/internal/team"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, team.IRepository) {
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

func TestCheckTeamNameExist_TeamNotExist(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "team1"

	mock.ExpectQuery(regexp.QuoteMeta(CheckTeamNameExistQuery)).WithArgs(teamName).WillReturnError(sql.ErrNoRows)

	isExist, err := repo.CheckTeamNameExist(ctx, teamName)
	require.NoError(t, err)
	assert.False(t, isExist)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckTeamNameExist_TeamExist(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "team2"
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

	mock.ExpectQuery(regexp.QuoteMeta(CheckTeamNameExistQuery)).WithArgs(teamName).WillReturnRows(rows)

	isExist, err := repo.CheckTeamNameExist(ctx, teamName)
	require.NoError(t, err)
	assert.True(t, isExist)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckTeamNameExist_DBError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "team2"
	dbError := errors.New("db failure")

	mock.ExpectQuery(regexp.QuoteMeta(CheckTeamNameExistQuery)).WithArgs(teamName).WillReturnError(dbError)

	isExist, err := repo.CheckTeamNameExist(ctx, teamName)
	require.Error(t, err)
	assert.False(t, isExist)
	assert.EqualError(t, err, dbError.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTeam_Successfull(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "team2"

	mock.ExpectExec(regexp.QuoteMeta(CreateTeamQuery)).WithArgs(teamName).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.CreateTeam(ctx, teamName)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTeam_Failure(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "team2"
	dbErr := errors.New("insert failed")

	mock.ExpectExec(regexp.QuoteMeta(CreateTeamQuery)).WithArgs(teamName).WillReturnError(dbErr)

	err := repo.CreateTeam(ctx, teamName)
	require.Error(t, err)
	assert.EqualError(t, err, dbErr.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}
