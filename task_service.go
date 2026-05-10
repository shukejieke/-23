package service

import (
	"stzbHelper/internal/repo"
	"stzbHelper/model"
)

func CreateTask(task *model.Task) error {
	return repo.CreateTask(task)
}

func GetTaskList() ([]model.Task, error) {
	return repo.ListTasks()
}

func GetTask(id int) (*model.Task, error) {
	return repo.GetTaskByID(id)
}

func DeleteTask(id int) (int64, error) {
	return repo.DeleteTaskByID(id)
}

func GetUsersByGroups(groups []string) ([]model.TeamUser, error) {
	return repo.FindUsersByGroups(groups)
}

func SaveTask(task *model.Task) (int64, error) {
	return repo.SaveTask(task)
}

func DeleteTaskReportsByWid(wid int) (int64, error) {
	return repo.DeleteReportByWid(wid)
}
