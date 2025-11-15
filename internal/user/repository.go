package user

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
)

type IRepository interface {
	GetExistentUsers(ctx context.Context, membersIds []string) (map[string]struct{}, error)
	CreateUsers(ctx context.Context, users []*entity.TeamMember, teamName string) error
	UpdateUsersTeam(ctx context.Context, users []*entity.TeamMember, teamName string) error
	GetMembersByTeamName(ctx context.Context, teamName string) ([]*entity.TeamMember, error)
	SetIsActive(ctx context.Context, userId string, isActive bool) error
	CheckUserExistById(ctx context.Context, userId string) (bool, error)
	GetUserById(ctx context.Context, userId string) (*entity.User, error)
	FindReviewers(ctx context.Context, authorId string) ([]string, error)
}
