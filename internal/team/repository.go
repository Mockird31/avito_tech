package team

import (
	"context"
)

type IRepository interface {
	CheckTeamNameExist(ctx context.Context, teamName string) (bool, error)
	CreateTeam(ctx context.Context, teamName string) error
}
