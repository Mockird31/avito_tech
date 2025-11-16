package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/stats"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"go.uber.org/zap"
)

const (
	GetAssignmentsStatsByReviewersQuery = `
        SELECT prr.reviewer_id, COUNT(*) as cnt
        FROM pull_request_reviewers prr
        JOIN pull_request p ON p.id = prr.pull_request_id
        WHERE p.status = 'OPEN'
        GROUP BY prr.reviewer_id
        ORDER BY prr.reviewer_id;
    `
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) stats.IRepository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetAssignmentsStatsByReviewers(ctx context.Context) ([]*entity.UserAssignmentCount, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	rows, err := r.db.QueryContext(ctx, GetAssignmentsStatsByReviewersQuery)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("no assignments by reviewers not found")
			return []*entity.UserAssignmentCount{}, nil
		}
		logger.Error("failed to get assignments by reviewers", zap.Error(err))
		return nil, err
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil && err == nil {
			logger.Error("failed to close rows", zap.Error(err))
			err = closeErr
		}
	}()

	assignmentsStats := make([]*entity.UserAssignmentCount, 0)
	for rows.Next() {
		var assignmentStat entity.UserAssignmentCount
		err := rows.Scan(&assignmentStat.UserId, &assignmentStat.Count)
		if err != nil {
			logger.Error("failed to scan data", zap.Error(err))
			return nil, err
		}
		assignmentsStats = append(assignmentsStats, &assignmentStat)
	}

	if err := rows.Close(); err != nil {
		logger.Error("failed to pass through rows", zap.Error(err))
		return nil, err
	}
	return assignmentsStats, nil
}
