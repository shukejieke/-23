package service

import "stzbHelper/model"

func GetReportCountByTaskID(taskID int) (int64, error) {
	task, err := GetTask(taskID)
	if err != nil {
		return 0, err
	}
	var taskNum int64
	err = model.Conn.Model(model.Report{}).Where(model.Report{Wid: task.Pos}).Count(&taskNum).Error
	return taskNum, err
}

func StatisticsTaskReport(taskID int) (int64, error) {
	task, err := GetTask(taskID)
	if err != nil {
		return 0, err
	}

	task.CompleteUserNum = 0
	for id, t := range task.UserList {
		var atkNum, disNum, atkTeamNum, disTeamNum int64
		model.Conn.Model(model.Report{}).Where(model.Report{Wid: task.Pos, AttackName: t.Name}).Where("garrison = ?", 0).Count(&atkNum)
		model.Conn.Model(model.Report{}).Where(model.Report{Wid: task.Pos, AttackName: t.Name, Garrison: 1}).Count(&disNum)
		model.Conn.Model(model.Report{}).Where(model.Report{Wid: task.Pos, AttackName: t.Name}).Where("garrison = ?", 0).Group("attack_base_heroid").Count(&atkTeamNum)
		model.Conn.Model(model.Report{}).Where(model.Report{Wid: task.Pos, AttackName: t.Name, Garrison: 1}).Group("attack_base_heroid").Count(&disTeamNum)

		task.UserList[id].AtkNum = int(atkNum)
		task.UserList[id].DisNum = int(disNum)
		task.UserList[id].AtkTeamNum = int(atkTeamNum)
		task.UserList[id].DisTeamNum = int(disTeamNum)
		if atkNum != 0 || disNum != 0 {
			task.CompleteUserNum++
		}
	}
	return SaveTask(task)
}

func DeleteTaskReports(taskID int) (int64, error) {
	task, err := GetTask(taskID)
	if err != nil {
		return 0, err
	}
	return DeleteTaskReportsByWid(task.Pos)
}
