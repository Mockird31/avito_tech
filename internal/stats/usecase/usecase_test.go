package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/stats"
	mock_stats "github.com/Mockird31/avito_tech/mocks/stats"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (stats.IUsecase, *mock_stats.MockIRepository) {
	repo := mock_stats.NewMockIRepository(t)
	uc := NewUsecase(repo)
	return uc, repo
}

func getTestContext() context.Context {
	logger := zap.NewNop()
	ctx := context.Background()
	return loggerPkg.LoggerToContext(ctx, logger.Sugar())
}

func TestGetAssignmentsStatsByReviewers_Success(t *testing.T) {
	uc, repo := setupTest(t)
	ctx := getTestContext()

	want := []*entity.UserAssignmentCount{
		{UserId: "u1", Count: 3},
		{UserId: "u2", Count: 1},
	}

	repo.EXPECT().
		GetAssignmentsStatsByReviewers(mock.Anything).
		Return(want, nil)

	got, err := uc.GetAssignmentsStatsByReviewers(ctx)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetAssignmentsStatsByReviewers_RepoError(t *testing.T) {
	uc, repo := setupTest(t)
	ctx := getTestContext()

	dbErr := errors.New("select failed")

	repo.EXPECT().
		GetAssignmentsStatsByReviewers(mock.Anything).
		Return(([]*entity.UserAssignmentCount)(nil), dbErr)

	got, err := uc.GetAssignmentsStatsByReviewers(ctx)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.EqualError(t, err, dbErr.Error())
}
