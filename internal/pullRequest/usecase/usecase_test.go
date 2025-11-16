package usecase

import (
	"context"
	"testing"

	"github.com/Mockird31/avito_tech/internal/entity"
	pullrequest "github.com/Mockird31/avito_tech/internal/pullRequest"
	mock_pullrequest "github.com/Mockird31/avito_tech/mocks/pullrequest"
	mock_team "github.com/Mockird31/avito_tech/mocks/team"
	mock_user "github.com/Mockird31/avito_tech/mocks/user"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (pullrequest.IUsecase, *mock_team.MockIRepository, *mock_user.MockIRepository, *mock_pullrequest.MockIRepository) {
	teamRepo := mock_team.NewMockIRepository(t)
	userRepo := mock_user.NewMockIRepository(t)
	prRepo := mock_pullrequest.NewMockIRepository(t)

	prUsecase := NewUsecase(prRepo, userRepo, teamRepo)
	return prUsecase, teamRepo, userRepo, prRepo
}

func getTestContext() context.Context {
	logger := zap.NewNop()
	ctx := context.Background()
	return loggerPkg.LoggerToContext(ctx, logger.Sugar())
}

func TestGetPullRequestById_Success(t *testing.T) {
	uc, _, _, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-1"
	base := &entity.PullRequest{
		Id:       prId,
		PrName:   "Add feature",
		AuthorId: "u1",
		Status:   "OPEN",
	}
	reviewers := []string{"r1", "r2"}

	prRepo.EXPECT().
		GetPullRequestById(mock.Anything, prId).
		Return(base, nil)
	prRepo.EXPECT().
		GetReviewersByPrId(mock.Anything, prId).
		Return(reviewers, nil)

	got, err := uc.GetPullRequestById(ctx, prId)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, prId, got.Id)
	assert.Equal(t, "Add feature", got.PrName)
	assert.Equal(t, "u1", got.AuthorId)
	assert.Equal(t, "OPEN", got.Status)
	assert.Equal(t, reviewers, got.AssignedReviewersIds)
}

func TestGetPullRequestById_PrRepoError(t *testing.T) {
	uc, _, _, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-err"
	prRepo.EXPECT().
		GetPullRequestById(mock.Anything, prId).
		Return((*entity.PullRequest)(nil), assert.AnError)

	got, err := uc.GetPullRequestById(ctx, prId)
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestGetPullRequestById_ReviewersError(t *testing.T) {
	uc, _, _, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-2"
	base := &entity.PullRequest{
		Id:       prId,
		PrName:   "Fix bug",
		AuthorId: "u2",
		Status:   "OPEN",
	}

	prRepo.EXPECT().
		GetPullRequestById(mock.Anything, prId).
		Return(base, nil)
	prRepo.EXPECT().
		GetReviewersByPrId(mock.Anything, prId).
		Return(([]string)(nil), assert.AnError)

	got, err := uc.GetPullRequestById(ctx, prId)
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestCreatePullRequest_Success_WithReviewers(t *testing.T) {
	uc, teamRepo, userRepo, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-new"
	prName := "Cool feature"
	authorId := "u1"
	author := &entity.User{UserId: authorId, TeamName: "teamA"}
	reviewers := []string{"r1", "r2"}

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(false, nil)
	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, authorId).
		Return(true, nil)
	userRepo.EXPECT().
		GetUserById(mock.Anything, authorId).
		Return(author, nil)
	teamRepo.EXPECT().
		CheckTeamNameExist(mock.Anything, "teamA").
		Return(true, nil)
	prRepo.EXPECT().
		CreatePullRequest(mock.Anything, prId, prName, authorId).
		Return(nil)
	userRepo.EXPECT().
		FindReviewers(mock.Anything, authorId).
		Return(reviewers, nil)
	prRepo.EXPECT().
		ConnectReviewersWithPullRequest(mock.Anything, prId, reviewers).
		Return(nil)

	req := &entity.PullRequest{Id: prId, PrName: prName, AuthorId: authorId}
	got, err := uc.CreatePullRequest(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, prId, got.Id)
	assert.Equal(t, prName, got.PrName)
	assert.Equal(t, authorId, got.AuthorId)
	assert.Equal(t, "OPEN", got.Status)
	assert.Equal(t, reviewers, got.AssignedReviewersIds)
}

func TestCreatePullRequest_Success_NoReviewers(t *testing.T) {
	uc, teamRepo, userRepo, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-empty"
	prName := "No reviewers case"
	authorId := "u2"
	author := &entity.User{UserId: authorId, TeamName: "teamB"}

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(false, nil)
	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, authorId).
		Return(true, nil)
	userRepo.EXPECT().
		GetUserById(mock.Anything, authorId).
		Return(author, nil)
	teamRepo.EXPECT().
		CheckTeamNameExist(mock.Anything, "teamB").
		Return(true, nil)
	prRepo.EXPECT().
		CreatePullRequest(mock.Anything, prId, prName, authorId).
		Return(nil)
	userRepo.EXPECT().
		FindReviewers(mock.Anything, authorId).
		Return([]string{}, nil)

	req := &entity.PullRequest{Id: prId, PrName: prName, AuthorId: authorId}
	got, err := uc.CreatePullRequest(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, []string{}, got.AssignedReviewersIds)

	prRepo.AssertNotCalled(t, "ConnectReviewersWithPullRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreatePullRequest_AlreadyExists(t *testing.T) {
	uc, _, userRepo, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-exist"
	authorId := "u1"

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(true, nil)

	req := &entity.PullRequest{Id: prId, PrName: "dup", AuthorId: authorId}
	got, err := uc.CreatePullRequest(ctx, req)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, entity.ErrPullRequestExist)

	userRepo.AssertNotCalled(t, "CheckUserExistById", mock.Anything, mock.Anything)
}

func TestCreatePullRequest_CheckExist_Error(t *testing.T) {
	uc, _, _, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-err"

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(false, assert.AnError)

	req := &entity.PullRequest{Id: prId, PrName: "x", AuthorId: "u1"}
	got, err := uc.CreatePullRequest(ctx, req)
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestCreatePullRequest_AuthorCheck_Error(t *testing.T) {
	uc, _, userRepo, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-1"
	authorId := "u1"

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(false, nil)
	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, authorId).
		Return(false, assert.AnError)

	req := &entity.PullRequest{Id: prId, PrName: "x", AuthorId: authorId}
	got, err := uc.CreatePullRequest(ctx, req)
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestCreatePullRequest_Author_NotExist(t *testing.T) {
	uc, _, userRepo, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-2"
	authorId := "missing"

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(false, nil)
	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, authorId).
		Return(false, nil)

	req := &entity.PullRequest{Id: prId, PrName: "x", AuthorId: authorId}
	got, err := uc.CreatePullRequest(ctx, req)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, entity.ErrAuthorOrTeamNotExist)

	userRepo.AssertNotCalled(t, "GetUserById", mock.Anything, mock.Anything)
}

func TestCreatePullRequest_GetUserById_Error(t *testing.T) {
	uc, _, userRepo, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-3"
	authorId := "u1"

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(false, nil)
	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, authorId).
		Return(true, nil)
	userRepo.EXPECT().
		GetUserById(mock.Anything, authorId).
		Return((*entity.User)(nil), assert.AnError)

	req := &entity.PullRequest{Id: prId, PrName: "x", AuthorId: authorId}
	got, err := uc.CreatePullRequest(ctx, req)
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestCreatePullRequest_TeamCheck_Error(t *testing.T) {
	uc, teamRepo, userRepo, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-4"
	authorId := "u1"
	author := &entity.User{UserId: authorId, TeamName: "teamX"}

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(false, nil)
	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, authorId).
		Return(true, nil)
	userRepo.EXPECT().
		GetUserById(mock.Anything, authorId).
		Return(author, nil)
	teamRepo.EXPECT().
		CheckTeamNameExist(mock.Anything, "teamX").
		Return(false, assert.AnError)

	req := &entity.PullRequest{Id: prId, PrName: "x", AuthorId: authorId}
	got, err := uc.CreatePullRequest(ctx, req)
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestCreatePullRequest_Team_NotExist(t *testing.T) {
	uc, teamRepo, userRepo, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-5"
	authorId := "u1"
	author := &entity.User{UserId: authorId, TeamName: "teamY"}

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(false, nil)
	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, authorId).
		Return(true, nil)
	userRepo.EXPECT().
		GetUserById(mock.Anything, authorId).
		Return(author, nil)
	teamRepo.EXPECT().
		CheckTeamNameExist(mock.Anything, "teamY").
		Return(false, nil)

	req := &entity.PullRequest{Id: prId, PrName: "x", AuthorId: authorId}
	got, err := uc.CreatePullRequest(ctx, req)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, entity.ErrAuthorOrTeamNotExist)
}

func TestCreatePullRequest_Create_Error(t *testing.T) {
	uc, teamRepo, userRepo, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-6"
	prName := "x"
	authorId := "u1"
	author := &entity.User{UserId: authorId, TeamName: "teamZ"}

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(false, nil)
	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, authorId).
		Return(true, nil)
	userRepo.EXPECT().
		GetUserById(mock.Anything, authorId).
		Return(author, nil)
	teamRepo.EXPECT().
		CheckTeamNameExist(mock.Anything, "teamZ").
		Return(true, nil)
	prRepo.EXPECT().
		CreatePullRequest(mock.Anything, prId, prName, authorId).
		Return(assert.AnError)

	req := &entity.PullRequest{Id: prId, PrName: prName, AuthorId: authorId}
	got, err := uc.CreatePullRequest(ctx, req)
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestCreatePullRequest_FindReviewers_Error(t *testing.T) {
	uc, teamRepo, userRepo, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-7"
	prName := "x"
	authorId := "u1"
	author := &entity.User{UserId: authorId, TeamName: "teamA"}

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(false, nil)
	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, authorId).
		Return(true, nil)
	userRepo.EXPECT().
		GetUserById(mock.Anything, authorId).
		Return(author, nil)
	teamRepo.EXPECT().
		CheckTeamNameExist(mock.Anything, "teamA").
		Return(true, nil)
	prRepo.EXPECT().
		CreatePullRequest(mock.Anything, prId, prName, authorId).
		Return(nil)
	userRepo.EXPECT().
		FindReviewers(mock.Anything, authorId).
		Return(([]string)(nil), assert.AnError)

	req := &entity.PullRequest{Id: prId, PrName: prName, AuthorId: authorId}
	got, err := uc.CreatePullRequest(ctx, req)
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestCreatePullRequest_ConnectReviewers_Error(t *testing.T) {
	uc, teamRepo, userRepo, prRepo := setupTest(t)
	ctx := getTestContext()

	prId := "pr-8"
	prName := "x"
	authorId := "u1"
	author := &entity.User{UserId: authorId, TeamName: "teamA"}
	reviewers := []string{"r1", "r2"}

	prRepo.EXPECT().
		CheckPullRequestExistById(mock.Anything, prId).
		Return(false, nil)
	userRepo.EXPECT().
		CheckUserExistById(mock.Anything, authorId).
		Return(true, nil)
	userRepo.EXPECT().
		GetUserById(mock.Anything, authorId).
		Return(author, nil)
	teamRepo.EXPECT().
		CheckTeamNameExist(mock.Anything, "teamA").
		Return(true, nil)
	prRepo.EXPECT().
		CreatePullRequest(mock.Anything, prId, prName, authorId).
		Return(nil)
	userRepo.EXPECT().
		FindReviewers(mock.Anything, authorId).
		Return(reviewers, nil)
	prRepo.EXPECT().
		ConnectReviewersWithPullRequest(mock.Anything, prId, reviewers).
		Return(assert.AnError)

	req := &entity.PullRequest{Id: prId, PrName: prName, AuthorId: authorId}
	got, err := uc.CreatePullRequest(ctx, req)
	require.Error(t, err)
	assert.Nil(t, got)
}
