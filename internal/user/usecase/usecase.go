package usecase

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/user"
)

type Usecase struct {
	UserRepository user.IRepository
}

func NewUsecase(userRepository user.IRepository) user.IUsecase {
	return &Usecase{
		UserRepository: userRepository,
	}
}

func (u *Usecase) SetIsActive(ctx context.Context, userUpdateActive *entity.UserUpdateActive) (*entity.User, error) {
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
