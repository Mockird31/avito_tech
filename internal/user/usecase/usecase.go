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
