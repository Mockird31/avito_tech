package usecase

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
	pullrequest "github.com/Mockird31/avito_tech/internal/pullRequest"
	"github.com/Mockird31/avito_tech/internal/team"
	"github.com/Mockird31/avito_tech/internal/user"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"go.uber.org/zap"
)

type usecase struct {
	PRRepository   pullrequest.IRepository
	UserRepository user.IRepository
	TeamRepository team.IRepository
}

func NewUsecase(PRRepository pullrequest.IRepository, UserRepository user.IRepository, TeamRepository team.IRepository) pullrequest.IUsecase {
	return &usecase{
		PRRepository:   PRRepository,
		UserRepository: UserRepository,
		TeamRepository: TeamRepository,
	}
}

func (u *usecase) GetPullRequestById(ctx context.Context, prId string) (*entity.PullRequest, error) {
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

func (u *usecase) CreatePullRequest(ctx context.Context, pullRequestCreate *entity.PullRequest) (*entity.PullRequest, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	isExist, err := u.PRRepository.CheckPullRequestExistById(ctx, pullRequestCreate.Id)
	if err != nil {
		return nil, err
	}

	if isExist {
		logger.Error("pull request with id is exist (CreatePullRequest)", zap.String("pr_id", pullRequestCreate.Id))
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

func (u *usecase) MergePullRequest(ctx context.Context, pullRequestMerge *entity.PullRequest) (*entity.PullRequest, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	isExist, err := u.PRRepository.CheckPullRequestExistById(ctx, pullRequestMerge.Id)
	if err != nil {
		return nil, err
	}

	if !isExist {
		logger.Error("pull request with id is not exist (MergePullRequest)", zap.Error(err), zap.String("pr_id", pullRequestMerge.Id))
		return nil, entity.ErrPullRequestNotExist
	}

	isMerged, err := u.PRRepository.CheckPullRequestIsMergedById(ctx, pullRequestMerge.Id)
	if err != nil {
		return nil, err
	}

	if !isMerged {
		err := u.PRRepository.MergePullRequest(ctx, pullRequestMerge.Id)
		if err != nil {
			return nil, err
		}
	}

	pullRequest, err := u.GetPullRequestById(ctx, pullRequestMerge.Id)
	if err != nil {
		return nil, err
	}

	return pullRequest, nil
}

func (u *usecase) ReassignPullRequest(ctx context.Context, pullRequestReassign *entity.PullRequestReassignRequest) (*entity.PullRequest, string, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	isExist, err := u.PRRepository.CheckPullRequestExistById(ctx, pullRequestReassign.Id)
	if err != nil {
		return nil, "", err
	}

	if !isExist {
		logger.Error("pull request with id is not exist (ReassignPullRequest)", zap.Error(err), zap.String("pr_id", pullRequestReassign.Id))
		return nil, "", entity.ErrPullRequestNotExist
	}

	isOldReviewerExist, err := u.UserRepository.CheckUserExistById(ctx, pullRequestReassign.OldReviewerId)
	if err != nil {
		return nil, "", err
	}

	if !isOldReviewerExist {
		return nil, "", entity.ErrUserNotFound
	}

	isMerged, err := u.PRRepository.CheckPullRequestIsMergedById(ctx, pullRequestReassign.Id)
	if err != nil {
		return nil, "", err
	}

	if isMerged {
		return nil, "", entity.ErrRequestAlreadyMerged
	}

	authorId, err := u.PRRepository.GetAuthorIdByPRId(ctx, pullRequestReassign.Id)
	if err != nil {
		return nil, "", err
	}

	newReviewerId, err := u.UserRepository.FindNewReviewer(ctx, pullRequestReassign.Id, authorId, pullRequestReassign.OldReviewerId)
	if err != nil {
		return nil, "", err
	}

	if newReviewerId == "" {
		logger.Info("no available reviewer (ReassignPullRequest)")
		pullRequest, err := u.GetPullRequestById(ctx, pullRequestReassign.Id)
		if err != nil {
			return nil, "", err
		}
		return pullRequest, "", nil
	}

	err = u.PRRepository.UpdateReviewerId(ctx, pullRequestReassign.Id, pullRequestReassign.OldReviewerId, newReviewerId)
	if err != nil {
		return nil, "", err
	}

	pullRequest, err := u.GetPullRequestById(ctx, pullRequestReassign.Id)
	if err != nil {
		return nil, "", err
	}

	return pullRequest, newReviewerId, nil
}
