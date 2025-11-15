package pullrequest

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
)

type IRepository interface {
	CreatePullRequest(ctx context.Context, prId string, prName string, authorId string) error
	CheckPullRequestExistById(ctx context.Context, prId string) (bool, error)
	ConnectReviewersWithPullRequest(ctx context.Context, prId string, reviewersIds []string) error
	GetPullRequestById(ctx context.Context, prId string) (*entity.PullRequest, error)
	GetReviewersByPrId(ctx context.Context, prId string) ([]string, error)
	MergePullRequest(ctx context.Context, prId string) error
	CheckPullRequestIsMergedById(ctx context.Context, prId string) (bool, error)
	GetAuthorIdByPRId(ctx context.Context, oldReviewerId string) (string, error)
	UpdateReviewerId(ctx context.Context, prId string, oldReviewerId string, newReviewerId string) error
	GetPullRequestsByReviewerId(ctx context.Context, reviewerId string) ([]*entity.PullRequestShort, error)
}
