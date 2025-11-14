package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Mockird31/avito_tech/internal/team"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"go.uber.org/zap"
)

const (
	CheckTeamNameExistQuery = `
		SELECT 1
		FROM team
		WHERE name = $1;
	`

	CreateTeamQuery = `
		INSERT INTO team (name) VALUES ($1)
	`
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) team.IRepository {
	return &repository{
		db: db,
	}
}

func (r *repository) CheckTeamNameExist(ctx context.Context, teamName string) (bool, error) {
	logger := loggerPkg.LoggerFromContext(ctx)
	var isExist bool

	err := r.db.QueryRowContext(ctx, CheckTeamNameExistQuery, teamName).Scan(&isExist)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return isExist, nil
		}
		logger.Error("faile to check team name:", zap.Error(err))
		return isExist, err
	}

	return isExist, nil
}

func (r *repository) CreateTeam(ctx context.Context, teamName string) error {
	logger := loggerPkg.LoggerFromContext(ctx)
	if _, err := r.db.ExecContext(ctx, CreateTeamQuery, teamName); err != nil {
		logger.Error("failed to create team:", zap.Error(err))
		return err
	}
	return nil
}
