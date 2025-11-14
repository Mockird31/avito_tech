package usecase

import (
	"context"
	"testing"

	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/team"
	mock_team "github.com/Mockird31/avito_tech/mocks/team"
	mock_user "github.com/Mockird31/avito_tech/mocks/user"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (team.IUsecase, *mock_team.MockIRepository, *mock_user.MockIRepository) {
	teamRepo := mock_team.NewMockIRepository(t)
	userRepo := mock_user.NewMockIRepository(t)

	teamUsecase := NewUsecase(teamRepo, userRepo)
	return teamUsecase, teamRepo, userRepo
}

func getTestContext() context.Context {
	logger := zap.NewNop()
	ctx := context.Background()
	return loggerPkg.LoggerToContext(ctx, logger.Sugar())
}

func TestAddTeam_NameExists(t *testing.T) {
	ctx := getTestContext()

	uc, teamRepo, _ := setupTest(t)

	teamName := "alpha"
	req := &entity.Team{
		TeamName: teamName,
		Members:  []*entity.TeamMember{},
	}

	teamRepo.EXPECT().
		CheckTeamNameExist(mock.Anything, teamName).
		Return(true, nil)

	res, err := uc.AddTeam(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, entity.ErrTeamNameExist)

	teamRepo.AssertNotCalled(t, "CreateTeam", mock.Anything, mock.Anything)
}

func TestAddTeam_Success_MixedExistent(t *testing.T) {
	ctx := context.Background()

	teamRepo := mock_team.NewMockIRepository(t)
	userRepo := mock_user.NewMockIRepository(t)

	uc := NewUsecase(teamRepo, userRepo)

	teamName := "gamma"
	m1 := &entity.TeamMember{UserID: "u1", Username: "alice", IsActive: true} // существует
	m2 := &entity.TeamMember{UserID: "u2", Username: "bob", IsActive: false}  // новый
	req := &entity.Team{
		TeamName: teamName,
		Members:  []*entity.TeamMember{m1, m2},
	}

	teamRepo.EXPECT().
		CheckTeamNameExist(mock.Anything, teamName).
		Return(false, nil)
	teamRepo.EXPECT().
		CreateTeam(mock.Anything, teamName).
		Return(nil)

	userRepo.EXPECT().
		GetExistentUsers(mock.Anything, []string{"u1", "u2"}).
		Return(map[string]struct{}{"u1": {}}, nil)
	userRepo.EXPECT().
		CreateUsers(mock.Anything, []*entity.TeamMember{m2}, teamName).
		Return(nil)
	userRepo.EXPECT().
		UpdateUsersTeam(mock.Anything, []*entity.TeamMember{m1}, teamName).
		Return(nil)

	res, err := uc.AddTeam(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, req, res)
}
