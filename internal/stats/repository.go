package stats

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
)

type IRepository interface {
	GetAssignmentsStatsByReviewers(ctx context.Context) ([]*entity.UserAssignmentCount, error)
}
