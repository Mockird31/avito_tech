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
}
