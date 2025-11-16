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
