package entity

import "errors"

var (
	ErrTeamNameExist        = errors.New("team_name already exists")
	ErrTeamNameNotFound     = errors.New("resource not found")
	ErrTeamNoMembersByTeam  = errors.New("no members found by team name")
	ErrUserNotFound         = errors.New("resource not found")
	ErrPullRequestExist     = errors.New("PR id already exists")
	ErrAuthorOrTeamNotExist = errors.New("resource not found")
	ErrPullRequestNotExist  = errors.New("resource not found")
	ErrRequestAlreadyMerged = errors.New("cannot reassign on merged PR")
	ErrUsersNotSameTeam     = errors.New("users not in the same team")
)
