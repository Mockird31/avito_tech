package usecase

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
	pullrequest "github.com/Mockird31/avito_tech/internal/pullRequest"
	"github.com/Mockird31/avito_tech/internal/team"
	"github.com/Mockird31/avito_tech/internal/user"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
)

type Usecase struct {
	PRRepository   pullrequest.IRepository
	UserRepository user.IRepository
	TeamRepository team.IRepository
}

func NewUsecase(PRRepository pullrequest.IRepository, UserRepository user.IRepository, TeamRepository team.IRepository) pullrequest.IUsecase {
	return &Usecase{
		PRRepository:   PRRepository,
		UserRepository: UserRepository,
		TeamRepository: TeamRepository,
	}
}

func (u *Usecase) GetPullRequestById(ctx context.Context, prId string) (*entity.PullRequest, error) {
	pullrequest, err := u.PRRepository.GetPullRequestById(ctx, prId)
	if err != nil {
		return nil, err
	}

	reviewersId, err := u.PRRepository.GetReviewersByPrId(ctx, prId)
	if err != nil {
		return nil, err
	}

	pullrequest.AssignedReviewersIds = reviewersId

	return pullrequest, nil
}

func (u *Usecase) CreatePullRequest(ctx context.Context, pullRequestCreate *entity.PullRequest) (*entity.PullRequest, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	isExist, err := u.PRRepository.CheckPullRequestExistById(ctx, pullRequestCreate.Id)
	if err != nil {
		return nil, err
	}

	if isExist {
		logger.Error("pull request with id is exist (CreatePullRequest)", "pr_id", pullRequestCreate.Id)
		return nil, entity.ErrPullRequestExist
	}

	isAuthorExist, err := u.UserRepository.CheckUserExistById(ctx, pullRequestCreate.AuthorId)
	if err != nil {
		return nil, err
	}
	if !isAuthorExist {
		return nil, entity.ErrAuthorOrTeamNotExist
	}

	author, err := u.UserRepository.GetUserById(ctx, pullRequestCreate.AuthorId)
	if err != nil {
		return nil, err
	}
	isTeamExist, err := u.TeamRepository.CheckTeamNameExist(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}
	if !isTeamExist {
		return nil, entity.ErrAuthorOrTeamNotExist
	}

	err = u.PRRepository.CreatePullRequest(ctx, pullRequestCreate.Id, pullRequestCreate.PrName, pullRequestCreate.AuthorId)
	if err != nil {
		return nil, err
	}

	reviewersIds, err := u.UserRepository.FindReviewers(ctx, pullRequestCreate.AuthorId)
	if err != nil {
		return nil, err
	}

	if len(reviewersIds) > 0 {
		err = u.PRRepository.ConnectReviewersWithPullRequest(ctx, pullRequestCreate.Id, reviewersIds)
		if err != nil {
			return nil, err
		}
	}

	pullRequest := &entity.PullRequest{
		Id:                   pullRequestCreate.Id,
		PrName:               pullRequestCreate.PrName,
		AuthorId:             pullRequestCreate.AuthorId,
		Status:               "OPEN",
		AssignedReviewersIds: reviewersIds,
	}
	return pullRequest, nil
}
