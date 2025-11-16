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

func TestUpdateUsersTeam_Success(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "teamX"
	users := []*entity.TeamMember{
		{UserID: "u1", Username: "alice", IsActive: true},
		{UserID: "u2", Username: "bob", IsActive: false},
	}

	mock.ExpectExec(regexp.QuoteMeta(UpdateUsersTeamQuery)).
		WithArgs(teamName, sqlmock.AnyArg()). // pq.Array(ids)
		WillReturnResult(sqlmock.NewResult(0, int64(len(users))))

	err := repo.UpdateUsersTeam(ctx, users, teamName)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUsersTeam_DBError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "teamY"
	users := []*entity.TeamMember{
		{UserID: "u1", Username: "alice", IsActive: true},
	}

	dbErr := errors.New("update failed")
	mock.ExpectExec(regexp.QuoteMeta(UpdateUsersTeamQuery)).
		WithArgs(teamName, sqlmock.AnyArg()). // pq.Array(ids)
		WillReturnError(dbErr)

	err := repo.UpdateUsersTeam(ctx, users, teamName)
	require.Error(t, err)
	assert.EqualError(t, err, dbErr.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUsersTeam_EmptyUsers(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "teamZ"
	var users []*entity.TeamMember

	mock.ExpectExec(regexp.QuoteMeta(UpdateUsersTeamQuery)).
		WithArgs(teamName, sqlmock.AnyArg()). // pq.Array(empty)
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateUsersTeam(ctx, users, teamName)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMembersByTeamName_Success(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "alpha"

	rows := sqlmock.NewRows([]string{"id", "username", "is_active"}).
		AddRow("u1", "alice", true).
		AddRow("u2", "bob", false)

	mock.ExpectQuery(regexp.QuoteMeta(GetMembersByTeamNameQuery)).
		WithArgs(teamName).
		WillReturnRows(rows)

	members, err := repo.GetMembersByTeamName(ctx, teamName)
	require.NoError(t, err)
	require.Len(t, members, 2)
	assert.Equal(t, "u1", members[0].UserID)
	assert.Equal(t, "alice", members[0].Username)
	assert.Equal(t, true, members[0].IsActive)
	assert.Equal(t, "u2", members[1].UserID)
	assert.Equal(t, "bob", members[1].Username)
	assert.Equal(t, false, members[1].IsActive)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMembersByTeamName_NoRows(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "empty-team"

	mock.ExpectQuery(regexp.QuoteMeta(GetMembersByTeamNameQuery)).
		WithArgs(teamName).
		WillReturnError(sql.ErrNoRows)

	members, err := repo.GetMembersByTeamName(ctx, teamName)
	require.Error(t, err)
	assert.Nil(t, members)
	assert.ErrorIs(t, err, entity.ErrTeamNoMembersByTeam)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMembersByTeamName_DBError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "team-err"
	dbErr := errors.New("db failure")

	mock.ExpectQuery(regexp.QuoteMeta(GetMembersByTeamNameQuery)).
		WithArgs(teamName).
		WillReturnError(dbErr)

	members, err := repo.GetMembersByTeamName(ctx, teamName)
	require.Error(t, err)
	assert.Nil(t, members)
	assert.EqualError(t, err, dbErr.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMembersByTeamName_ScanError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "bad-scan"

	rows := sqlmock.NewRows([]string{"id", "username", "is_active"}).
		AddRow("u1", "alice", "not_bool")

	mock.ExpectQuery(regexp.QuoteMeta(GetMembersByTeamNameQuery)).
		WithArgs(teamName).
		WillReturnRows(rows)

	members, err := repo.GetMembersByTeamName(ctx, teamName)
	require.Error(t, err)
	assert.Nil(t, members)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMembersByTeamName_RowsErr(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	teamName := "rows-err"

	rows := sqlmock.NewRows([]string{"id", "username", "is_active"}).
		AddRow("u1", "alice", true).
		RowError(0, errors.New("row iteration error"))

	mock.ExpectQuery(regexp.QuoteMeta(GetMembersByTeamNameQuery)).
		WithArgs(teamName).
		WillReturnRows(rows)

	members, err := repo.GetMembersByTeamName(ctx, teamName)
	require.Error(t, err)
	assert.Nil(t, members)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckUserExistById_UserExist(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	userId := "u1"
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

	mock.ExpectQuery(regexp.QuoteMeta(CheckUserExistByIdQuery)).
		WithArgs(userId).
		WillReturnRows(rows)

	isExist, err := repo.CheckUserExistById(ctx, userId)
	require.NoError(t, err)
	assert.True(t, isExist)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckUserExistById_UserNotExist(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	userId := "u-missing"

	mock.ExpectQuery(regexp.QuoteMeta(CheckUserExistByIdQuery)).
		WithArgs(userId).
		WillReturnError(sql.ErrNoRows)

	isExist, err := repo.CheckUserExistById(ctx, userId)
	require.NoError(t, err)
	assert.False(t, isExist)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckUserExistById_DBError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	userId := "u1"
	dbErr := errors.New("db failure")

	mock.ExpectQuery(regexp.QuoteMeta(CheckUserExistByIdQuery)).
		WithArgs(userId).
		WillReturnError(dbErr)

	isExist, err := repo.CheckUserExistById(ctx, userId)
	require.Error(t, err)
	assert.False(t, isExist)
	assert.EqualError(t, err, dbErr.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSetIsActive_Success(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	userId := "u1"
	isActive := true

	mock.ExpectExec(regexp.QuoteMeta(UpdateUserActiveQuery)).
		WithArgs(isActive, userId).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetIsActive(ctx, userId, isActive)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSetIsActive_DBError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	userId := "u1"
	isActive := false
	dbErr := errors.New("update failed")

	mock.ExpectExec(regexp.QuoteMeta(UpdateUserActiveQuery)).
		WithArgs(isActive, userId).
		WillReturnError(dbErr)

	err := repo.SetIsActive(ctx, userId, isActive)
	require.Error(t, err)
	assert.EqualError(t, err, dbErr.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserById_Success(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	userId := "u1"
	rows := sqlmock.NewRows([]string{"id", "username", "team_name", "is_active"}).
		AddRow("u1", "alice", "teamA", true)

	mock.ExpectQuery(regexp.QuoteMeta(GetUserByIdQuery)).
		WithArgs(userId).
		WillReturnRows(rows)

	u, err := repo.GetUserById(ctx, userId)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "u1", u.UserId)
	assert.Equal(t, "alice", u.Username)
	assert.Equal(t, "teamA", u.TeamName)
	assert.Equal(t, true, u.IsActive)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserById_NoRows(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	userId := "missing"

	mock.ExpectQuery(regexp.QuoteMeta(GetUserByIdQuery)).
		WithArgs(userId).
		WillReturnError(sql.ErrNoRows)

	u, err := repo.GetUserById(ctx, userId)
	require.Error(t, err)
	assert.Nil(t, u)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserById_DBError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	userId := "u1"
	dbErr := errors.New("db failure")

	mock.ExpectQuery(regexp.QuoteMeta(GetUserByIdQuery)).
		WithArgs(userId).
		WillReturnError(dbErr)

	u, err := repo.GetUserById(ctx, userId)
	require.Error(t, err)
	assert.Nil(t, u)
	assert.EqualError(t, err, dbErr.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserById_ScanError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	userId := "u1"
	rows := sqlmock.NewRows([]string{"id", "username", "team_name", "is_active"}).
		AddRow("u1", "alice", "teamA", "not_bool")

	mock.ExpectQuery(regexp.QuoteMeta(GetUserByIdQuery)).
		WithArgs(userId).
		WillReturnRows(rows)

	u, err := repo.GetUserById(ctx, userId)
	require.Error(t, err)
	assert.Nil(t, u)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindReviewers_Success(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	authorId := "author1"

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow("r1").
		AddRow("r2")

	mock.ExpectQuery(regexp.QuoteMeta(FindReviewersQuery)).
		WithArgs(authorId).
		WillReturnRows(rows)

	reviewers, err := repo.FindReviewers(ctx, authorId)
	require.NoError(t, err)
	assert.Equal(t, []string{"r1", "r2"}, reviewers)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindReviewers_NoRows(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	authorId := "author2"

	mock.ExpectQuery(regexp.QuoteMeta(FindReviewersQuery)).
		WithArgs(authorId).
		WillReturnError(sql.ErrNoRows)

	reviewers, err := repo.FindReviewers(ctx, authorId)
	require.NoError(t, err)
	assert.Empty(t, reviewers)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindReviewers_DBError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	authorId := "author3"
	dbErr := errors.New("db failure")

	mock.ExpectQuery(regexp.QuoteMeta(FindReviewersQuery)).
		WithArgs(authorId).
		WillReturnError(dbErr)

	reviewers, err := repo.FindReviewers(ctx, authorId)
	require.Error(t, err)
	assert.Nil(t, reviewers)
	assert.EqualError(t, err, dbErr.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindReviewers_RowsErr(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	authorId := "author5"

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow("r1").
		RowError(0, errors.New("row iteration error"))

	mock.ExpectQuery(regexp.QuoteMeta(FindReviewersQuery)).
		WithArgs(authorId).
		WillReturnRows(rows)

	reviewers, err := repo.FindReviewers(ctx, authorId)
	require.Error(t, err)
	assert.Nil(t, reviewers)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUsersIsActiveByIds_Success(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	ids := []string{"u1", "u2"}
	isActive := false

	mock.ExpectExec(regexp.QuoteMeta(UpdateUsersIsActiveByIdsQuery)).
		WithArgs(isActive, sqlmock.AnyArg()). // pq.Array(ids)
		WillReturnResult(sqlmock.NewResult(0, int64(len(ids))))

	err := repo.UpdateUsersIsActiveByIds(ctx, ids, isActive)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUsersIsActiveByIds_DBError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	ids := []string{"u1"}
	isActive := true
	dbErr := errors.New("bulk update failed")

	mock.ExpectExec(regexp.QuoteMeta(UpdateUsersIsActiveByIdsQuery)).
		WithArgs(isActive, sqlmock.AnyArg()). // pq.Array(ids)
		WillReturnError(dbErr)

	err := repo.UpdateUsersIsActiveByIds(ctx, ids, isActive)
	require.Error(t, err)
	assert.EqualError(t, err, dbErr.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUsersIsActiveByIds_EmptyIDs(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	var ids []string
	isActive := false

	mock.ExpectExec(regexp.QuoteMeta(UpdateUsersIsActiveByIdsQuery)).
		WithArgs(isActive, sqlmock.AnyArg()). // pq.Array(empty)
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateUsersIsActiveByIds(ctx, ids, isActive)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindNewReviewerExcluding_Success(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	prID := "pr1"
	authorID := "a1"
	exclude := []string{"u1", "u2"}

	rows := sqlmock.NewRows([]string{"id"}).AddRow("u3")

	mock.ExpectQuery(regexp.QuoteMeta(FindNewReviewerExcludingQuery)).
		WithArgs(authorID, prID, sqlmock.AnyArg()). // pq.Array(exclude)
		WillReturnRows(rows)

	reviewerID, err := repo.FindNewReviewerExcluding(ctx, prID, authorID, exclude)
	require.NoError(t, err)
	assert.Equal(t, "u3", reviewerID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindNewReviewerExcluding_NoRows(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	prID := "pr2"
	authorID := "a2"
	exclude := []string{"u1"}

	mock.ExpectQuery(regexp.QuoteMeta(FindNewReviewerExcludingQuery)).
		WithArgs(authorID, prID, sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	reviewerID, err := repo.FindNewReviewerExcluding(ctx, prID, authorID, exclude)
	require.NoError(t, err)
	assert.Empty(t, reviewerID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindNewReviewerExcluding_DBError(t *testing.T) {
	db, mock, repo := setupTest(t)
	defer db.Close()
	ctx := getTestContext()

	prID := "pr3"
	authorID := "a3"
	exclude := []string{"u9"}
	dbErr := errors.New("select failed")

	mock.ExpectQuery(regexp.QuoteMeta(FindNewReviewerExcludingQuery)).
		WithArgs(authorID, prID, sqlmock.AnyArg()).
		WillReturnError(dbErr)

	reviewerID, err := repo.FindNewReviewerExcluding(ctx, prID, authorID, exclude)
	require.Error(t, err)
	assert.Empty(t, reviewerID)
	assert.EqualError(t, err, dbErr.Error())

	require.NoError(t, mock.ExpectationsWereMet())
}
