package user

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
)

type IUsecase interface {
	SetIsActive(ctx context.Context, userUpdateActive *entity.UserUpdateActive) (*entity.User, error)
	GetUserReview(ctx context.Context, userId string) ([]*entity.PullRequestShort, string, error)
	DeactivateTeamUsers(ctx context.Context, deactivateUsers *entity.DeactivateUsers) (*entity.DeactivateUsers, error)
}
