package service

import (
	"stzbHelper/internal/repo"
	"stzbHelper/model"
)

func GetTeamUsers(group string) ([]model.TeamUser, error) {
	return repo.ListTeamUsers(group)
}

func GetTeamGroups() ([]string, error) {
	return repo.ListTeamGroups()
}

func GetGroupWuStats() ([]repo.GroupWuStats, error) {
	return repo.QueryGroupWuStats()
}
