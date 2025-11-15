package user

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
)

type IUsecase interface {
	SetIsActive(ctx context.Context, userUpdateActive *entity.UserUpdateActive) (*entity.User, error)
}
