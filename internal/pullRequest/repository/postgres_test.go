package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	pullrequest "github.com/Mockird31/avito_tech/internal/pullRequest"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, pullrequest.IRepository) {
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

// ...existing code...

func TestCheckPullRequestExistById_Exist(t *testing.T) {
    db, mock, repo := setupTest(t)
    defer db.Close()
    ctx := getTestContext()

    prID := "pr-1"
    rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

    mock.ExpectQuery(regexp.QuoteMeta(CheckPullRequestExistByIdQuery)).
        WithArgs(prID).
        WillReturnRows(rows)

    isExist, err := repo.CheckPullRequestExistById(ctx, prID)
    require.NoError(t, err)
    assert.True(t, isExist)

    require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckPullRequestExistById_NotExist(t *testing.T) {
    db, mock, repo := setupTest(t)
    defer db.Close()
    ctx := getTestContext()

    prID := "missing"

    mock.ExpectQuery(regexp.QuoteMeta(CheckPullRequestExistByIdQuery)).
        WithArgs(prID).
        WillReturnError(sql.ErrNoRows)

    isExist, err := repo.CheckPullRequestExistById(ctx, prID)
    require.NoError(t, err)
    assert.False(t, isExist)

    require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckPullRequestExistById_DBError(t *testing.T) {
    db, mock, repo := setupTest(t)
    defer db.Close()
    ctx := getTestContext()

    prID := "pr-err"
    dbErr := errors.New("db failure")

    mock.ExpectQuery(regexp.QuoteMeta(CheckPullRequestExistByIdQuery)).
        WithArgs(prID).
        WillReturnError(dbErr)

    isExist, err := repo.CheckPullRequestExistById(ctx, prID)
    require.Error(t, err)
    assert.False(t, isExist)
    assert.EqualError(t, err, dbErr.Error())

    require.NoError(t, mock.ExpectationsWereMet())
}

// ...existing code...

func TestGetPullRequestsByReviewerId_Success(t *testing.T) {
    db, mock, repo := setupTest(t)
    defer db.Close()
    ctx := getTestContext()

    reviewerId := "rev-1"

    rows := sqlmock.NewRows([]string{"id", "name", "author_id", "status", "merged_at"}).
        AddRow("pr1", "Fix bug", "author1", "OPEN", nil).
        AddRow("pr2", "Add feature", "author2", "MERGED", nil)

    mock.ExpectQuery(regexp.QuoteMeta(GetPullRequestsByReviewerIdQuery)).
        WithArgs(reviewerId).
        WillReturnRows(rows)

    prs, err := repo.GetPullRequestsByReviewerId(ctx, reviewerId)
    require.NoError(t, err)
    require.Len(t, prs, 2)

    assert.Equal(t, "pr1", prs[0].Id)
    assert.Equal(t, "Fix bug", prs[0].PrName)
    assert.Equal(t, "author1", prs[0].AuthorId)
    assert.Equal(t, "OPEN", prs[0].Status)
    assert.Nil(t, prs[0].MergedAt)

    assert.Equal(t, "pr2", prs[1].Id)
    assert.Equal(t, "Add feature", prs[1].PrName)
    assert.Equal(t, "author2", prs[1].AuthorId)
    assert.Equal(t, "MERGED", prs[1].Status)
    assert.Nil(t, prs[1].MergedAt)

    require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPullRequestsByReviewerId_NoRows(t *testing.T) {
    db, mock, repo := setupTest(t)
    defer db.Close()
    ctx := getTestContext()

    reviewerId := "rev-empty"

    mock.ExpectQuery(regexp.QuoteMeta(GetPullRequestsByReviewerIdQuery)).
        WithArgs(reviewerId).
        WillReturnError(sql.ErrNoRows)

    prs, err := repo.GetPullRequestsByReviewerId(ctx, reviewerId)
    require.NoError(t, err)
    assert.Empty(t, prs)

    require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPullRequestsByReviewerId_DBError(t *testing.T) {
    db, mock, repo := setupTest(t)
    defer db.Close()
    ctx := getTestContext()

    reviewerId := "rev-err"
    dbErr := errors.New("db failure")

    mock.ExpectQuery(regexp.QuoteMeta(GetPullRequestsByReviewerIdQuery)).
        WithArgs(reviewerId).
        WillReturnError(dbErr)

    prs, err := repo.GetPullRequestsByReviewerId(ctx, reviewerId)
    require.Error(t, err)
    assert.Nil(t, prs)
    assert.EqualError(t, err, dbErr.Error())

    require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPullRequestsByReviewerId_ScanError(t *testing.T) {
    db, mock, repo := setupTest(t)
    defer db.Close()
    ctx := getTestContext()

    reviewerId := "rev-scan"

    rows := sqlmock.NewRows([]string{"id", "name", "author_id", "status", "merged_at"}).
        AddRow("pr1", "Fix bug", "author1", "OPEN", "not_time")

    mock.ExpectQuery(regexp.QuoteMeta(GetPullRequestsByReviewerIdQuery)).
        WithArgs(reviewerId).
        WillReturnRows(rows)

    prs, err := repo.GetPullRequestsByReviewerId(ctx, reviewerId)
    require.Error(t, err)
    assert.Nil(t, prs)

    require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPullRequestsByReviewerId_RowsErr(t *testing.T) {
    db, mock, repo := setupTest(t)
    defer db.Close()
    ctx := getTestContext()

    reviewerId := "rev-rows-err"

    rows := sqlmock.NewRows([]string{"id", "name", "author_id", "status", "merged_at"}).
        AddRow("pr1", "Fix bug", "author1", "OPEN", nil).
        RowError(0, errors.New("row iteration error"))

    mock.ExpectQuery(regexp.QuoteMeta(GetPullRequestsByReviewerIdQuery)).
        WithArgs(reviewerId).
        WillReturnRows(rows)

    prs, err := repo.GetPullRequestsByReviewerId(ctx, reviewerId)
    require.Error(t, err)
    assert.Nil(t, prs)

    require.NoError(t, mock.ExpectationsWereMet())
}