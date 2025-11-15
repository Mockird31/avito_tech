package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Mockird31/avito_tech/internal/entity"
	pullrequest "github.com/Mockird31/avito_tech/internal/pullRequest"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"go.uber.org/zap"
)

const (
	CheckPullRequestExistByIdQuery = `
		SELECT 1 
		FROM pull_request
		WHERE id = $1;
	`
	GetPullRequestByIdQuery = `
		SELECT id, name, author_id, status, merged_at
		FROM pull_request
		WHERE id = $1;
	`
	GetReviewersByPrId = `
		SELECT u.id
		FROM "user"
		JOIN pull_request_reviewers prr ON prr.reviewer_id = u.id
		WHERE prr.pull_request_id = $1;
	`
	CreatePullRequestQuery = `
		INSERT INTO pull_request
		(id, name, author_id, status)
		VALUES ($1, $2, $3, $4);
	`
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) pullrequest.IRepository {
	return &repository{
		db: db,
	}
}

func (r *repository) CheckPullRequestExistById(ctx context.Context, prId string) (bool, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	var isExist bool

	err := r.db.QueryRowContext(ctx, CheckPullRequestExistByIdQuery, prId).Scan(&isExist)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("pull request not found by id", zap.String("pr_id", prId))
			return isExist, nil
		}
		logger.Error("failed to check is pull request exist (CheckPullRequestExistById)", zap.Error(err))
		return isExist, err
	}
	return isExist, nil
}

func (r *repository) GetReviewersByPrId(ctx context.Context, prId string) ([]string, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	rows, err := r.db.QueryContext(ctx, GetReviewersByPrId, prId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("no reviewers to pr", "pr_id", prId)
			return []string{}, nil
		}
		logger.Error("failed to get reviewers (GetReviewersByPrId) by pr_id", "pr_id", prId, "error", zap.Error(err))
		return nil, err
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	reviewersIds := make([]string, 0)
	for rows.Next() {
		var reviewerId string
		err := rows.Scan(&reviewerId)
		if err != nil {
			logger.Error("failed to scan reviewerId (GetReviewersByPrId)", zap.Error(err))
			return nil, err
		}
		reviewersIds = append(reviewersIds, reviewerId)
	}

	if err = rows.Err(); err != nil {
		logger.Error("failed while iterate through rows (GetReviewersByPrId)", zap.Error(err))
		return nil, err
	}

	return reviewersIds, nil
}

func (r *repository) GetPullRequestById(ctx context.Context, prId string) (*entity.PullRequest, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	var pullRequest entity.PullRequest
	var mergedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, GetPullRequestByIdQuery, prId).Scan(&pullRequest.Id, &pullRequest.PrName, &pullRequest.AuthorId, &pullRequest.Status, &mergedAt)
	if err != nil {
		logger.Error("failed to get pull request by id", "id", prId, "error", zap.Error(err))
		return nil, err
	}

	if mergedAt.Valid {
		pullRequest.MergedAt = &mergedAt.Time
	} else {
		pullRequest.MergedAt = nil
	}

	return &pullRequest, nil
}

func (r *repository) CreatePullRequest(ctx context.Context, prId string, prName string, authorId string) error {
	logger := loggerPkg.LoggerFromContext(ctx)

	_, err := r.db.ExecContext(ctx, CreatePullRequestQuery, prId, prName, authorId, entity.StatusOpen)
	if err != nil {
		logger.Error("failed to create pull request", "pr_id", prId, "error", zap.Error(err))
		return err
	}
	return nil
}

func PrepareConnectReviewersQuery(ctx context.Context, prId string, reviewersIds []string) (string, []any, error) {
	logger := loggerPkg.LoggerFromContext(ctx)
	var sb strings.Builder
	_, err := sb.WriteString(`INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id) VALUES`)
	if err != nil {
		return "", nil, err
	}

	args := make([]any, 0, len(reviewersIds)+1)
	args = append(args, prId)

	for i, reviewerId := range reviewersIds {
		if i > 0 {
			_, err := sb.WriteString(",")
			if err != nil {
				logger.Error("failed to create query (ConnectReviewersWithPullRequest)", zap.Error(err))
				return "", nil, err
			}
		}
		_, err := sb.WriteString(fmt.Sprintf(" ($1, $%d)", i+2))
		if err != nil {
			logger.Error("failed to create query (ConnectReviewersWithPullRequest)", zap.Error(err))
			return "", nil, err
		}
		args = append(args, reviewerId)
	}

	_, err = sb.WriteString(";")
	if err != nil {
		logger.Error("failed to create query (ConnectReviewersWithPullRequest)", zap.Error(err))
		return "", nil, err
	}

	return sb.String(), args, nil
}

func (r *repository) ConnectReviewersWithPullRequest(ctx context.Context, prId string, reviewersIds []string) error {
	logger := loggerPkg.LoggerFromContext(ctx)

	query, args, err := PrepareConnectReviewersQuery(ctx, prId, reviewersIds)
	if err != nil {
		logger.Error("failed to prepare query (ConnectReviewersWithPullRequest)", zap.Error(err))
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Error("failed to connect reviewers with pull request (ConnectReviewersWithPullRequest)", "pr_id", prId, "error", zap.Error(err))
		return err
	}

	return nil
}
