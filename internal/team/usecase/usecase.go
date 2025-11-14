package usecase

import (
	"context"

	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/team"
	"github.com/Mockird31/avito_tech/internal/user"
)

type Usecase struct {
	TeamRepository team.IRepository
	UserRepository user.IRepository
}

func NewUsecase(teamRepository team.IRepository, userRepository user.IRepository) team.IUsecase {
	return &Usecase{
		TeamRepository: teamRepository,
		UserRepository: userRepository,
	}
}

func (u *Usecase) AddTeam(ctx context.Context, team *entity.Team) (*entity.Team, error) {
	isExist, err := u.TeamRepository.CheckTeamNameExist(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}

	if isExist {
		return nil, entity.ErrTeamNameExist
	}

	err = u.TeamRepository.CreateTeam(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}

	membersIds := make([]string, 0, len(team.Members))
	for _, member := range team.Members {
		membersIds = append(membersIds, member.UserID)
	}

	existentIds, err := u.UserRepository.GetExistentUsers(ctx, membersIds)
	if err != nil {
		return nil, err
	}

	var existentUsers []*entity.TeamMember
	var nonExistentUsers []*entity.TeamMember

	for _, member := range team.Members {
		if _, ok := existentIds[member.UserID]; ok {
			existentUsers = append(existentUsers, member)
		} else {
			nonExistentUsers = append(nonExistentUsers, member)
		}
	}

	if len(nonExistentUsers) > 0 {
		err = u.UserRepository.CreateUsers(ctx, nonExistentUsers, team.TeamName)
		if err != nil {
			return nil, err
		}
	}

	if len(existentUsers) > 0 {
		err = u.UserRepository.UpdateUsersTeam(ctx, existentUsers, team.TeamName)
		if err != nil {
			return nil, err
		}
	}

	return team, nil
}

func (u *Usecase) GetTeam(ctx context.Context, teamName string) (*entity.Team, error) {
	isExist, err := u.TeamRepository.CheckTeamNameExist(ctx, teamName)
	if err != nil {
		return nil, err
	}

	if !isExist {
		return nil, entity.ErrTeamNameNotFound
	}

	members, err := u.UserRepository.GetMembersByTeamName(ctx, teamName)
	if err != nil {
		return nil, err
	}

	collectedTeam := &entity.Team{
		TeamName: teamName,
		Members:  members,
	}
	return collectedTeam, nil
}
