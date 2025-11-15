package pullrequest

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
)

type IUsecase interface {
	GetPullRequestById(ctx context.Context, prId string) (*entity.PullRequest, error)
	CreatePullRequest(ctx context.Context, pullRequestCreate *entity.PullRequest) (*entity.PullRequest, error)
}
