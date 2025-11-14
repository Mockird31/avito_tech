package team

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
)

type IUsecase interface {
	AddTeam(ctx context.Context, team *entity.Team) (*entity.Team, error)
	GetTeam(ctx context.Context, teamName string) (*entity.Team, error)
}
