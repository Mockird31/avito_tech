package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/user"
	mock_pullrequest "github.com/Mockird31/avito_tech/mocks/pullrequest"
	mock_user "github.com/Mockird31/avito_tech/mocks/user"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (user.IUsecase, *mock_user.MockIRepository, *mock_pullrequest.MockIRepository) {
	userRepo := mock_user.NewMockIRepository(t)
	prRepo := mock_pullrequest.NewMockIRepository(t)

	userUsecase := NewUsecase(userRepo, prRepo)
	return userUsecase, userRepo, prRepo
}

func getTestContext() context.Context {
	logger := zap.NewNop()
	ctx := context.Background()
	return loggerPkg.LoggerToContext(ctx, logger.Sugar())
}

func TestSetIsActive_UserNotExist(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, _ := setupTest(t)

	req := &entity.UserUpdateActive{UserId: "u1", IsActive: true}

	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, "u1").
		Return(false, nil)

	res, err := uc.SetIsActive(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, entity.ErrUserNotFound)

	userRepo.AssertNotCalled(t, "SetIsActive", mock.Anything, mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "GetUserById", mock.Anything, mock.Anything)
}

func TestSetIsActive_Success(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, _ := setupTest(t)

	req := &entity.UserUpdateActive{UserId: "u1", IsActive: true}
	want := &entity.User{UserId: "u1", Username: "alice", TeamName: "teamA", IsActive: true}

	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, "u1").
		Return(true, nil)
	userRepo.EXPECT().
		SetIsActive(mock.Anything, "u1", true).
		Return(nil)
	userRepo.EXPECT().
		GetUserById(mock.Anything, "u1").
		Return(want, nil)

	res, err := uc.SetIsActive(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, want, res)
}

func TestSetIsActive_GetUser_DBError(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, _ := setupTest(t)

	req := &entity.UserUpdateActive{UserId: "u1", IsActive: true}
	dbErr := errors.New("select failed")

	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, "u1").
		Return(true, nil)
	userRepo.EXPECT().
		SetIsActive(mock.Anything, "u1", true).
		Return(nil)
	userRepo.EXPECT().
		GetUserById(mock.Anything, "u1").
		Return(nil, dbErr)

	res, err := uc.SetIsActive(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.EqualError(t, err, dbErr.Error())
}

func TestGetUserReview_Success(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	userId := "u1"
	want := []*entity.PullRequestShort{
		{Id: "pr1", PrName: "PR 1"},
		{Id: "pr2", PrName: "PR 2"},
	}

	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, userId).
		Return(true, nil)
	prRepo.EXPECT().
		GetPullRequestsByReviewerId(mock.Anything, userId).
		Return(want, nil)

	got, id, err := uc.GetUserReview(ctx, userId)
	require.NoError(t, err)
	assert.Equal(t, userId, id)
	assert.Equal(t, want, got)
}

func TestGetUserReview_CheckExist_DBError(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	userId := "u1"
	dbErr := errors.New("db failure")

	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, userId).
		Return(false, dbErr)

	got, id, err := uc.GetUserReview(ctx, userId)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Empty(t, id)
	assert.EqualError(t, err, dbErr.Error())

	prRepo.AssertNotCalled(t, "GetPullRequestsByReviewerId", mock.Anything, mock.Anything)
}

func TestGetUserReview_UserNotExist(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	userId := "u-missing"

	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, userId).
		Return(false, nil)

	got, id, err := uc.GetUserReview(ctx, userId)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Empty(t, id)
	assert.ErrorIs(t, err, entity.ErrUserNotFound)

	prRepo.AssertNotCalled(t, "GetPullRequestsByReviewerId", mock.Anything, mock.Anything)
}

func TestGetUserReview_PRRepoError(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	userId := "u1"
	dbErr := errors.New("select failed")

	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, userId).
		Return(true, nil)
	prRepo.EXPECT().
		GetPullRequestsByReviewerId(mock.Anything, userId).
		Return(nil, dbErr)

	got, id, err := uc.GetUserReview(ctx, userId)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Empty(t, id)
	assert.EqualError(t, err, dbErr.Error())
}

func TestDeactivateTeamUsers_EmptyList(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	req := &entity.DeactivateUsers{TeamName: "teamA", UserIds: []string{}}

	res, err := uc.DeactivateTeamUsers(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, &entity.DeactivateUsers{TeamName: "teamA", UserIds: []string{}}, res)

	prRepo.AssertNotCalled(t, "GetPullRequestsByReviewerId", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "GetUsersByIds", mock.Anything, mock.Anything)
}

func TestDeactivateTeamUsers_GetUsersByIds_Error(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	req := &entity.DeactivateUsers{TeamName: "teamA", UserIds: []string{"u1", "u2"}}
	dbErr := errors.New("db failure")

	userRepo.EXPECT().
		GetUsersByIds(mock.Anything, []string{"u1", "u2"}).
		Return(nil, dbErr)

	res, err := uc.DeactivateTeamUsers(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.EqualError(t, err, dbErr.Error())

	prRepo.AssertNotCalled(t, "GetPullRequestsByReviewerId", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "UpdateUsersIsActiveByIds", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeactivateTeamUsers_UserNotFound(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	req := &entity.DeactivateUsers{TeamName: "teamA", UserIds: []string{"u1", "u2"}}

	userRepo.EXPECT().
		GetUsersByIds(mock.Anything, []string{"u1", "u2"}).
		Return(map[string]*entity.User{
			"u1": {UserId: "u1", TeamName: "teamA", IsActive: true},
		}, nil)

	res, err := uc.DeactivateTeamUsers(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, entity.ErrUserNotFound)

	prRepo.AssertNotCalled(t, "GetPullRequestsByReviewerId", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "UpdateUsersIsActiveByIds", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeactivateTeamUsers_UsersNotSameTeam(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	req := &entity.DeactivateUsers{TeamName: "teamA", UserIds: []string{"u1", "u2"}}

	userRepo.EXPECT().
		GetUsersByIds(mock.Anything, []string{"u1", "u2"}).
		Return(map[string]*entity.User{
			"u1": {UserId: "u1", TeamName: "teamA", IsActive: true},
			"u2": {UserId: "u2", TeamName: "teamB", IsActive: true},
		}, nil)

	res, err := uc.DeactivateTeamUsers(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, entity.ErrUsersNotSameTeam)

	prRepo.AssertNotCalled(t, "GetPullRequestsByReviewerId", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "UpdateUsersIsActiveByIds", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeactivateTeamUsers_Success_ReassignAndDeactivate(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	req := &entity.DeactivateUsers{TeamName: "teamA", UserIds: []string{"u1", "u2"}}

	userRepo.EXPECT().
		GetUsersByIds(mock.Anything, []string{"u1", "u2"}).
		Return(map[string]*entity.User{
			"u1": {UserId: "u1", TeamName: "teamA", IsActive: true},
			"u2": {UserId: "u2", TeamName: "teamA", IsActive: true},
		}, nil)

	prRepo.EXPECT().
		GetPullRequestsByReviewerId(mock.Anything, "u1").
		Return([]*entity.PullRequestShort{
			{Id: "pr1", PrName: "A", AuthorId: "a1", Status: "OPEN"},
			{Id: "pr2", PrName: "B", AuthorId: "a2", Status: "MERGED"},
		}, nil)

	userRepo.EXPECT().
		FindNewReviewerExcluding(mock.Anything, "pr1", "a1", []string{"u1", "u2"}).
		Return("u3", nil)
	prRepo.EXPECT().
		UpdateReviewerId(mock.Anything, "pr1", "u1", "u3").
		Return(nil)

	prRepo.EXPECT().
		GetPullRequestsByReviewerId(mock.Anything, "u2").
		Return([]*entity.PullRequestShort{
			{Id: "pr3", PrName: "C", AuthorId: "a3", Status: "OPEN"},
		}, nil)
	userRepo.EXPECT().
		FindNewReviewerExcluding(mock.Anything, "pr3", "a3", []string{"u1", "u2"}).
		Return("", nil)

	userRepo.EXPECT().
		UpdateUsersIsActiveByIds(mock.Anything, []string{"u1", "u2"}, false).
		Return(nil)

	res, err := uc.DeactivateTeamUsers(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, req, res)
}

func TestDeactivateTeamUsers_GetPRs_Error(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	req := &entity.DeactivateUsers{TeamName: "teamA", UserIds: []string{"u1"}}
	dbErr := errors.New("select prs failed")

	userRepo.EXPECT().
		GetUsersByIds(mock.Anything, []string{"u1"}).
		Return(map[string]*entity.User{"u1": {UserId: "u1", TeamName: "teamA"}}, nil)

	prRepo.EXPECT().
		GetPullRequestsByReviewerId(mock.Anything, "u1").
		Return(nil, dbErr)

	res, err := uc.DeactivateTeamUsers(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.EqualError(t, err, dbErr.Error())
}

func TestDeactivateTeamUsers_FindNewReviewer_Error(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	req := &entity.DeactivateUsers{TeamName: "teamA", UserIds: []string{"u1"}}
	dbErr := errors.New("find reviewer failed")

	userRepo.EXPECT().
		GetUsersByIds(mock.Anything, []string{"u1"}).
		Return(map[string]*entity.User{"u1": {UserId: "u1", TeamName: "teamA"}}, nil)

	prRepo.EXPECT().
		GetPullRequestsByReviewerId(mock.Anything, "u1").
		Return([]*entity.PullRequestShort{
			{Id: "pr1", PrName: "A", AuthorId: "a1", Status: "OPEN"},
		}, nil)

	userRepo.EXPECT().
		FindNewReviewerExcluding(mock.Anything, "pr1", "a1", []string{"u1"}).
		Return("", dbErr)

	res, err := uc.DeactivateTeamUsers(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.EqualError(t, err, dbErr.Error())
}

func TestDeactivateTeamUsers_UpdateReviewer_Error(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	req := &entity.DeactivateUsers{TeamName: "teamA", UserIds: []string{"u1"}}
	dbErr := errors.New("update reviewer failed")

	userRepo.EXPECT().
		GetUsersByIds(mock.Anything, []string{"u1"}).
		Return(map[string]*entity.User{"u1": {UserId: "u1", TeamName: "teamA"}}, nil)

	prRepo.EXPECT().
		GetPullRequestsByReviewerId(mock.Anything, "u1").
		Return([]*entity.PullRequestShort{
			{Id: "pr1", PrName: "A", AuthorId: "a1", Status: "OPEN"},
		}, nil)

	userRepo.EXPECT().
		FindNewReviewerExcluding(mock.Anything, "pr1", "a1", []string{"u1"}).
		Return("u3", nil)

	prRepo.EXPECT().
		UpdateReviewerId(mock.Anything, "pr1", "u1", "u3").
		Return(dbErr)

	res, err := uc.DeactivateTeamUsers(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.EqualError(t, err, dbErr.Error())
}

func TestDeactivateTeamUsers_UpdateUsersIsActive_Error(t *testing.T) {
	ctx := getTestContext()
	uc, userRepo, prRepo := setupTest(t)

	req := &entity.DeactivateUsers{TeamName: "teamA", UserIds: []string{"u1"}}
	dbErr := errors.New("bulk deactivate failed")

	userRepo.EXPECT().
		GetUsersByIds(mock.Anything, []string{"u1"}).
		Return(map[string]*entity.User{"u1": {UserId: "u1", TeamName: "teamA"}}, nil)

	prRepo.EXPECT().
		GetPullRequestsByReviewerId(mock.Anything, "u1").
		Return([]*entity.PullRequestShort{}, nil)

	userRepo.EXPECT().
		UpdateUsersIsActiveByIds(mock.Anything, []string{"u1"}, false).
		Return(dbErr)

	res, err := uc.DeactivateTeamUsers(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.EqualError(t, err, dbErr.Error())
}
