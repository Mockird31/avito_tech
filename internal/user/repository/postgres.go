package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/user"
	loggerPkg "github.com/Mockird31/avito_tech/pkg/logger"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	GetExistentUsersQuery = `
		SELECT id
		FROM "user"
		WHERE id = ANY($1);
	`
	UpdateUsersTeamQuery = `
		UPDATE "user"
		SET team_name = $1
		WHERE id = ANY($2);
	`
	GetMembersByTeamNameQuery = `
		SELECT id, username, is_active
		FROM "user"
		WHERE team_name = $1;
	`
	UpdateUserActiveQuery = `
		UPDATE "user"
		SET is_active = $1
		WHERE id = $2;
	`
	CheckUserExistByIdQuery = `
		SELECT 1
		FROM "user"
		WHERE id = $1;
	`
	GetUserByIdQuery = `
		SELECT id, username, team_name, is_active
		FROM "user"
		WHERE id = $1;
	`
	FindReviewersQuery = `
        SELECT u.id
        FROM "user" u
        WHERE u.team_name = (SELECT team_name FROM "user" WHERE id = $1)
          AND u.id <> $1
          AND u.is_active = TRUE
        ORDER BY random()
        LIMIT 2;
    `
	FindNewReviewerQuery = `
        SELECT u.id
        FROM "user" u
        WHERE u.team_name = (SELECT team_name FROM "user" WHERE id = $1)
          AND u.id <> $1
          AND u.id <> $3
          AND u.is_active = TRUE
          AND u.id NOT IN (
              SELECT reviewer_id
              FROM pull_request_reviewers
              WHERE pull_request_id = $2
          )
        ORDER BY random()
        LIMIT 1;
    `
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) user.IRepository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetExistentUsers(ctx context.Context, membersIds []string) (map[string]struct{}, error) {
	logger := loggerPkg.LoggerFromContext(ctx)
	existingUsersMap := make(map[string]struct{})

	rows, err := r.db.QueryContext(ctx, GetExistentUsersQuery, pq.Array(membersIds))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("existent users not found")
			return existingUsersMap, nil
		}
		logger.Error("failed to get non-existent users:", zap.Error(err))
		return nil, err
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil && err == nil {
			err = closeErr
			logger.Error("failed to close rows (GetNonExistentUsers):", zap.Error(err))
		}
	}()

	for rows.Next() {
		var userId string
		err := rows.Scan(&userId)
		if err != nil {
			logger.Error("failed to scan (GetNonExistentUsers)", zap.Error(err))
			return nil, err
		}
		existingUsersMap[userId] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		logger.Error("failed to iterate through rows (GetNonExistentUsers)", zap.Error(err))
		return nil, err
	}

	return existingUsersMap, nil
}

func PrepareCreateUsersQuery(users []*entity.TeamMember, teamName string) (string, []any, error) {
	var sb strings.Builder

	args := make([]any, 0, len(users)*3)

	_, err := sb.WriteString(`INSERT INTO "user" (id, username, team_name, is_active) VALUES`)
	if err != nil {
		return "", nil, err
	}

	p := 1
	for i, u := range users {
		if i > 0 {
			sb.WriteString(",")
		}
		_, err := sb.WriteString(fmt.Sprintf(" ($%d, $%d, $%d, $%d)", p, p+1, p+2, p+3))
		if err != nil {
			return "", nil, err
		}
		args = append(args, u.UserID, u.Username, teamName, u.IsActive)
		p += 4
	}

	sb.WriteString(";")

	return sb.String(), args, nil

}

func (r *repository) CreateUsers(ctx context.Context, users []*entity.TeamMember, teamName string) error {
	logger := loggerPkg.LoggerFromContext(ctx)
	query, args, err := PrepareCreateUsersQuery(users, teamName)
	if err != nil {
		logger.Error("failed to prepare query (CreateUsers):", zap.Error(err))
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Error("failed to create users (CreateUsers):", zap.Error(err))
		return err
	}
	return nil
}

func (r *repository) UpdateUsersTeam(ctx context.Context, users []*entity.TeamMember, teamName string) error {
	logger := loggerPkg.LoggerFromContext(ctx)
	ids := make([]string, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.UserID)
	}

	_, err := r.db.ExecContext(ctx, UpdateUsersTeamQuery, teamName, pq.Array(ids))
	if err != nil {
		logger.Error("failed to update users team (UpdateUsersTeam):", zap.Error(err))
		return err
	}
	return err
}

func (r *repository) GetMembersByTeamName(ctx context.Context, teamName string) ([]*entity.TeamMember, error) {
	logger := loggerPkg.LoggerFromContext(ctx)
	rows, err := r.db.QueryContext(ctx, GetMembersByTeamNameQuery, teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrTeamNoMembersByTeam
		}
		logger.Error("failed to get team members (GetMembersByTeamName):", zap.Error(err))
		return nil, err
	}

	defer func() {
		closeErr := rows.Close()
		if closeErr != nil && err == nil {
			err = closeErr
			logger.Error("failed to close rows (GetMembersByTeamName):", zap.Error(err))
		}
	}()

	teamMembers := make([]*entity.TeamMember, 0)
	for rows.Next() {
		var member entity.TeamMember
		err := rows.Scan(&member.UserID, &member.Username, &member.IsActive)
		if err != nil {
			logger.Error("failed to scan (GetMembersByTeamName):", zap.Error(err))
			return nil, err
		}
		teamMembers = append(teamMembers, &member)
	}

	if err := rows.Err(); err != nil {
		logger.Error("failed to iterate through rows (GetMembersByTeamName):", zap.Error(err))
		return nil, err
	}

	return teamMembers, nil
}

func (r *repository) CheckUserExistById(ctx context.Context, userId string) (bool, error) {
	logger := loggerPkg.LoggerFromContext(ctx)
	var isExist bool
	err := r.db.QueryRowContext(ctx, CheckUserExistByIdQuery, userId).Scan(&isExist)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("user not found (CheckUserExistById)", zap.String("user_id", userId))
			return isExist, nil
		}
		logger.Error("failed to check user exist by id (CheckUserExistById)", zap.Error(err))
		return isExist, err
	}
	return isExist, nil
}

func (r *repository) SetIsActive(ctx context.Context, userId string, isActive bool) error {
	logger := loggerPkg.LoggerFromContext(ctx)
	_, err := r.db.ExecContext(ctx, UpdateUserActiveQuery, isActive, userId)
	if err != nil {
		logger.Error("failed to update user is_active (SetIsActive)", zap.Error(err))
		return err
	}
	return nil
}

func (r *repository) GetUserById(ctx context.Context, userId string) (*entity.User, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	var user entity.User

	err := r.db.QueryRowContext(ctx, GetUserByIdQuery, userId).Scan(&user.UserId, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		logger.Error("failed to get user by id (GetUserById)", zap.Error(err))
		return nil, err
	}

	return &user, nil
}

func (r *repository) FindReviewers(ctx context.Context, authorId string) ([]string, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	rows, err := r.db.QueryContext(ctx, FindReviewersQuery, authorId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("reviewers not found (FindReviewers)", "author_id", authorId)
			return []string{}, nil
		}
		logger.Error("failed to get reviewers (FindReviewers)", zap.Error(err))
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
			logger.Error("failed to scan reviewer id (FindReviewers)", zap.Error(err))
			return nil, err
		}
		reviewersIds = append(reviewersIds, reviewerId)
	}

	if err = rows.Err(); err != nil {
		logger.Error("failed while iterate through rows (FindReviewers)", zap.Error(err))
		return nil, err
	}

	return reviewersIds, nil
}

func (r *repository) FindNewReviewer(ctx context.Context, prId, authorId, oldReviewerId string) (string, error) {
	logger := loggerPkg.LoggerFromContext(ctx)

	var reviewerId string
	err := r.db.QueryRowContext(ctx, FindNewReviewerQuery, authorId, prId, oldReviewerId).Scan(&reviewerId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("no available reviewer (FindNewReviewer)", zap.String("pr_id", prId), zap.String("author_id", authorId), zap.String("exclude_user_id", oldReviewerId))
			return "", nil
		}
		logger.Error("failed to find new reviewer (FindNewReviewer)", zap.Error(err))
		return "", err
	}

	return reviewerId, nil
}
