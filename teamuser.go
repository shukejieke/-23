package model

import "gorm.io/gorm/clause"

// UpsertTeamUserFromCmd949 仅更新 cmd 949 能确定的字段（name/wu/group），
// 避免把 cmd 103 写入的 power/contribute/pos 等字段清零。
func UpsertTeamUserFromCmd949(id int, name, group string, wu, joinTime int) {
	if Conn == nil {
		return
	}
	_ = Conn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "wu", "group"}),
	}).Create(&TeamUser{
		Id:       id,
		Name:     name,
		Wu:       wu,
		Group:    group,
		JoinTime: joinTime,
	}).Error
}

type TeamUser struct {
	Id              int    `json:"id" gorm:"column:id"`
	Name            string `json:"name" gorm:"column:name"`
	ContributeTotal int    `json:"contribute_total" gorm:"column:contribute_total"`
	ContributeWeek  int    `json:"contribute_week" gorm:"column:contribute_week"`
	Pos             int    `json:"pos" gorm:"column:pos"`
	Power           int    `json:"power" gorm:"column:power"`
	Wu              int    `json:"wu" gorm:"column:wu"`
	Group           string `json:"group" gorm:"column:group"`
	JoinTime        int    `json:"join_time" gorm:"column:join_time"`
}

func (TeamUser) TableName() string {
	return "team_user"
}

/*
		同盟成员信息索引
		[
			//0 id?
	        //1 名称
			//2 总贡献
			//6 坐标
			//7 本周贡献
			//8 势力
			//10 武勋
			//13 分组
			//30 应该是加入时间
		]
*/

func ToTeamUser(data []any) TeamUser {
	if len(data) < 31 {
		return TeamUser{}
	}

	group := teamUserToString(data[13])
	if group == "" {
		group = "未分组"
	}

	teamUser := TeamUser{
		Id:              teamUserToInt(data[0]),
		Name:            teamUserToString(data[1]),
		ContributeTotal: teamUserToInt(data[2]),
		ContributeWeek:  teamUserToInt(data[7]),
		Pos:             teamUserToInt(data[6]),
		Power:           teamUserToInt(data[8]),
		Wu:              teamUserToInt(data[10]),
		Group:           group,
		JoinTime:        teamUserToInt(data[30]),
	}

	return teamUser
}

func teamUserToInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	default:
		return 0
	}
}

func teamUserToString(v any) string {
	s, _ := v.(string)
	return s
}
