package usecase

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
	pullrequest "github.com/Mockird31/avito_tech/internal/pullRequest"
	"github.com/Mockird31/avito_tech/internal/user"
	"go.uber.org/zap"

	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
)

type usecase struct {
	UserRepository user.IRepository
	PRRepository   pullrequest.IRepository
}

func NewUsecase(userRepository user.IRepository, PRRepository pullrequest.IRepository) user.IUsecase {
	return &usecase{
		UserRepository: userRepository,
		PRRepository:   PRRepository,
	}
}

func (u *usecase) SetIsActive(ctx context.Context, userUpdateActive *entity.UserUpdateActive) (*entity.User, error) {
	isExist, err := u.UserRepository.CheckUserExistById(ctx, userUpdateActive.UserId)
	if err != nil {
		return nil, err
	}

	if !isExist {
		return nil, entity.ErrUserNotFound
	}

	err = u.UserRepository.SetIsActive(ctx, userUpdateActive.UserId, userUpdateActive.IsActive)
	if err != nil {
		return nil, err
	}

	user, err := u.UserRepository.GetUserById(ctx, userUpdateActive.UserId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *usecase) GetUserReview(ctx context.Context, userId string) ([]*entity.PullRequestShort, string, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	isExist, err := u.UserRepository.CheckUserExistById(ctx, userId)
	if err != nil {
		return nil, "", err
	}

	if !isExist {
		logger.Error("user not exist", zap.Error(err), zap.String("user_id", userId))
		return nil, "", entity.ErrUserNotFound
	}

	pullRequests, err := u.PRRepository.GetPullRequestsByReviewerId(ctx, userId)
	if err != nil {
		return nil, "", err
	}

	return pullRequests, userId, nil
}

func (u *usecase) DeactivateTeamUsers(ctx context.Context, deactivateUsers *entity.DeactivateUsers) (*entity.DeactivateUsers, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	if len(deactivateUsers.UserIds) == 0 {
		return &entity.DeactivateUsers{TeamName: deactivateUsers.TeamName, UserIds: []string{}}, nil
	}
	usersMap, err := u.UserRepository.GetUsersByIds(ctx, deactivateUsers.UserIds)
	if err != nil {
		return nil, err
	}

	if len(usersMap) != len(deactivateUsers.UserIds) {
		return nil, entity.ErrUserNotFound
	}

	for _, id := range deactivateUsers.UserIds {
		uinfo := usersMap[id]
		if uinfo == nil || uinfo.TeamName != deactivateUsers.TeamName {
			return nil, entity.ErrUsersNotSameTeam
		}
	}

	exclude := make([]string, 0, len(deactivateUsers.UserIds))
	exclude = append(exclude, deactivateUsers.UserIds...)

	for _, reviewerID := range deactivateUsers.UserIds {
		prs, err := u.PRRepository.GetPullRequestsByReviewerId(ctx, reviewerID)
		if err != nil {
			return nil, err
		}
		for _, pr := range prs {
			if pr.Status != "OPEN" {
				continue
			}
			newReviewerID, err := u.UserRepository.FindNewReviewerExcluding(ctx, pr.Id, pr.AuthorId, exclude)
			if err != nil {
				return nil, err
			}
			if newReviewerID != "" {
				if err := u.PRRepository.UpdateReviewerId(ctx, pr.Id, reviewerID, newReviewerID); err != nil {
					return nil, err
				}
			} else {
				logger.Info("no available reviewer (DeactivateTeamUsersWithList)", zap.String("pr_id", pr.Id), zap.String("old_reviewer_id", reviewerID))
			}
		}
	}

	if err := u.UserRepository.UpdateUsersIsActiveByIds(ctx, deactivateUsers.UserIds, false); err != nil {
		return nil, err
	}

	return &entity.DeactivateUsers{
		TeamName: deactivateUsers.TeamName,
		UserIds:  deactivateUsers.UserIds,
	}, nil
}
