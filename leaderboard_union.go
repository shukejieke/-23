package model

import (
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm/clause"
)

type UnionLeaderboard struct {
	ID           int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Rank         int    `json:"rank" gorm:"column:rank;index"`
	UnionID      int64  `json:"union_id" gorm:"column:union_id;uniqueIndex"`
	Name         string `json:"name" gorm:"column:name;index"`
	Power        int64  `json:"power" gorm:"column:power;index"`
	TotalMember  int    `json:"total_member" gorm:"column:total_member"`
	TotalNPCCity int    `json:"total_npc_city" gorm:"column:total_npc_city"`
	Region       int    `json:"region" gorm:"column:region"`
	RefreshTime  int64  `json:"refresh_time" gorm:"column:refresh_time;index"`
	SourceCmd    int    `json:"source_cmd" gorm:"column:source_cmd"`
	CaptureTime  int64  `json:"capture_time" gorm:"column:capture_time;index"`
}

func (UnionLeaderboard) TableName() string {
	return "union_leaderboard"
}

func SaveUnionLeaderboardFromDecoded(decoded string, sourceCmd int) {
	if Conn == nil {
		log.Println("[union-lb] DB未初始化，跳过保存")
		return
	}
	var top []any
	if err := json.Unmarshal([]byte(decoded), &top); err != nil {
		log.Printf("[union-lb] JSON解析失败: %v", err)
		return
	}
	if len(top) < 5 {
		log.Printf("[union-lb] 数据结构不符: len=%d", len(top))
		return
	}
	rows, ok := top[4].([]any)
	if !ok || len(rows) == 0 {
		log.Printf("[union-lb] top[4]非数组或为空: ok=%v len=%d", ok, len(rows))
		return
	}
	now := time.Now().Unix()
	saved := 0
	for _, row := range rows {
		pair, ok := row.([]any)
		if !ok || len(pair) < 2 {
			continue
		}
		rank := toInt(pair[0])
		obj, ok := pair[1].(map[string]any)
		if !ok {
			continue
		}
		uid := toInt64(obj["union_id"])
		if uid <= 0 {
			continue
		}
		rec := UnionLeaderboard{
			Rank:         rank,
			UnionID:      uid,
			Name:         toString(obj["name"]),
			Power:        toInt64(obj["power"]),
			TotalMember:  toInt(obj["total_member"]),
			TotalNPCCity: toInt(obj["total_npc_city"]),
			Region:       toInt(obj["region"]),
			RefreshTime:  toInt64(obj["refresh_time"]),
			SourceCmd:    sourceCmd,
			CaptureTime:  now,
		}
		if err := Conn.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "union_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"rank", "name", "power", "total_member", "total_npc_city", "region", "refresh_time", "source_cmd", "capture_time",
			}),
		}).Create(&rec).Error; err != nil {
			log.Printf("[union-lb] 写库失败 rank=%d union_id=%d err=%v", rank, uid, err)
		} else {
			saved++
		}
	}
	log.Printf("[union-lb] cmd700 处理完成 saved=%d total=%d", saved, len(rows))
}

func toInt(v any) int {
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

func toInt64(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int:
		return int64(x)
	case int64:
		return x
	default:
		return 0
	}
}

func toString(v any) string {
	s, _ := v.(string)
	return s
}
