package repo

import "stzbHelper/model"

func CreateTask(task *model.Task) error {
	return DB().Create(task).Error
}

func ListTasks() ([]model.Task, error) {
	var taskList []model.Task
	err := DB().Omit("user_list").Order("id DESC").Find(&taskList).Error
	return taskList, err
}

func GetTaskByID(id int) (*model.Task, error) {
	var task model.Task
	if err := DB().Last(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func DeleteTaskByID(id int) (int64, error) {
	action := DB().Delete(&model.Task{}, id)
	return action.RowsAffected, action.Error
}

func SaveTask(task *model.Task) (int64, error) {
	action := DB().Save(task)
	return action.RowsAffected, action.Error
}

func DeleteReportByWid(wid int) (int64, error) {
	action := DB().Delete(&model.Report{}, "wid = ?", wid)
	return action.RowsAffected, action.Error
}
