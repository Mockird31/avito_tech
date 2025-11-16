package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Mockird31/avito_tech/internal/entity"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *repository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return db, mock, &repository{db: db}
}

func getTestContext() context.Context {
	logger := zap.NewNop()
	ctx := context.Background()
	return loggerPkg.LoggerToContext(ctx, logger.Sugar())
}

func TestGetAssignmentsStatsByReviewers_Success(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	rows := sqlmock.NewRows([]string{"reviewer_id", "cnt"}).
		AddRow("u1", 3).
		AddRow("u2", 1)

	mock.ExpectQuery(regexp.QuoteMeta(GetAssignmentsStatsByReviewersQuery)).
		WillReturnRows(rows)

	stats, err := repo.GetAssignmentsStatsByReviewers(ctx)
	require.NoError(t, err)
	require.Len(t, stats, 2)

	assert.Equal(t, &entity.UserAssignmentCount{UserId: "u1", Count: 3}, stats[0])
	assert.Equal(t, &entity.UserAssignmentCount{UserId: "u2", Count: 1}, stats[1])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAssignmentsStatsByReviewers_NoRows(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	mock.ExpectQuery(regexp.QuoteMeta(GetAssignmentsStatsByReviewersQuery)).
		WillReturnError(sql.ErrNoRows)

	stats, err := repo.GetAssignmentsStatsByReviewers(ctx)
	require.NoError(t, err)
	assert.Empty(t, stats)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAssignmentsStatsByReviewers_DBError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	dbErr := errors.New("select failed")
	mock.ExpectQuery(regexp.QuoteMeta(GetAssignmentsStatsByReviewersQuery)).
		WillReturnError(dbErr)

	stats, err := repo.GetAssignmentsStatsByReviewers(ctx)
	require.Error(t, err)
	assert.Nil(t, stats)
	assert.EqualError(t, err, dbErr.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAssignmentsStatsByReviewers_ScanError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	rows := sqlmock.NewRows([]string{"reviewer_id", "cnt"}).
		AddRow("u1", "not_int")

	mock.ExpectQuery(regexp.QuoteMeta(GetAssignmentsStatsByReviewersQuery)).
		WillReturnRows(rows)

	stats, err := repo.GetAssignmentsStatsByReviewers(ctx)
	require.Error(t, err)
	assert.Nil(t, stats)

	require.NoError(t, mock.ExpectationsWereMet())
}
