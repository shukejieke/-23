package model

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm/clause"
)

type PersonalLeaderboard struct {
	ID         int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	EventID    int64  `json:"event_id" gorm:"column:event_id;index:idx_personal_event_obj,unique"`
	ObjectID   int64  `json:"object_id" gorm:"column:object_id;index:idx_personal_event_obj,unique"`
	ParamRaw   string `json:"param_raw" gorm:"column:param_raw"`
	ParamA     string `json:"param_a" gorm:"column:param_a;index"`
	ParamB     string `json:"param_b" gorm:"column:param_b;index"`
	ExtraRaw   string `json:"extra_raw" gorm:"column:extra_raw"`
	Flag       int    `json:"flag" gorm:"column:flag"`
	SourceCmd  int    `json:"source_cmd" gorm:"column:source_cmd"`
	CaptureTime int64 `json:"capture_time" gorm:"column:capture_time;index"`
}

func (PersonalLeaderboard) TableName() string { return "personal_leaderboard" }

func SavePersonalLeaderboardFromDecoded(decoded string, sourceCmd int) {
	if Conn == nil {
		return
	}
	var arr []any
	if err := json.Unmarshal([]byte(decoded), &arr); err != nil {
		return
	}
	if len(arr) < 2 {
		return
	}
	now := time.Now().Unix()
	saved := 0
	skipped := 0
	for i := 0; i+1 < len(arr); i += 2 {
		eid := toInt64(arr[i])
		payload, ok := arr[i+1].([]any)
		if !ok || len(payload) < 4 {
			skipped++
			continue
		}
		objID := toInt64(payload[0])
		paramRaw := fmt.Sprintf("%v", payload[1])
		flag := toInt(payload[2])
		extra := fmt.Sprintf("%v", payload[3])
		a, b := split2(paramRaw)
		rec := PersonalLeaderboard{
			EventID: eid, ObjectID: objID, ParamRaw: paramRaw, ParamA: a, ParamB: b,
			ExtraRaw: extra, Flag: flag, SourceCmd: sourceCmd, CaptureTime: now,
		}
		if err := Conn.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "event_id"}, {Name: "object_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"param_raw", "param_a", "param_b", "extra_raw", "flag", "source_cmd", "capture_time"}),
		}).Create(&rec).Error; err != nil {
			log.Printf("[personal-lb] 写库失败 event_id=%d object_id=%d err=%v", eid, objID, err)
		} else {
			saved++
		}
	}
	log.Printf("[personal-lb] cmd514 处理完成 saved=%d skipped=%d total_pairs=%d", saved, skipped, len(arr)/2)
}

func split2(s string) (string, string) {
	parts := strings.SplitN(s, ",", 2)
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}
