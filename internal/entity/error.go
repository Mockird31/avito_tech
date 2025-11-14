package entity

import "errors"

var (
	ErrTeamNameExist       = errors.New("team_name already exists")
	ErrTeamNameNotFound    = errors.New("resource not found")
	ErrTeamNoMembersByTeam = errors.New("no members found by team name")
)
