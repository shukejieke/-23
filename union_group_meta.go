package model

import "gorm.io/gorm/clause"

type UnionGroupMeta struct {
	ID          int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	GroupID     int    `json:"group_id" gorm:"column:group_id;uniqueIndex"`
	GroupName   string `json:"group_name" gorm:"column:group_name;index"`
	GroupCode   int    `json:"group_code" gorm:"column:group_code;index"`
	LeaderName  string `json:"leader_name" gorm:"column:leader_name"`
	MemberCount int    `json:"member_count" gorm:"column:member_count"`
	Power       int64  `json:"power" gorm:"column:power"`
	CaptureTime int64  `json:"capture_time" gorm:"column:capture_time;index"`
}

func (UnionGroupMeta) TableName() string { return "union_group_meta" }

func UpsertUnionGroupMeta(rec UnionGroupMeta) {
	if Conn == nil {
		return
	}
	_ = Conn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"group_name", "group_code", "leader_name", "member_count", "power", "capture_time"}),
	}).Create(&rec).Error
}

func LoadGroupCodeNameMap() map[int]string {
	m := map[int]string{}
	if Conn == nil {
		return m
	}
	var rows []UnionGroupMeta
	if err := Conn.Find(&rows).Error; err != nil {
		return m
	}
	for _, r := range rows {
		if r.GroupCode != 0 && r.GroupName != "" {
			m[r.GroupCode] = r.GroupName
		}
	}
	return m
}
