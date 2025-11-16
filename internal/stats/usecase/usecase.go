package usecase

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/stats"
)

type usecase struct {
	statsRepository stats.IRepository
}

func NewUsecase(statsRepo stats.IRepository) stats.IUsecase {
	return &usecase{
		statsRepository: statsRepo,
	}
}

func (u *usecase) GetAssignmentsStatsByReviewers(ctx context.Context) ([]*entity.UserAssignmentCount, error) {
	assignmentsStats, err := u.statsRepository.GetAssignmentsStatsByReviewers(ctx)
	if err != nil {
		return nil, err
	}
	return assignmentsStats, nil
}
