package model

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"gorm.io/gorm/clause"
)

// PlayerTerritoryRank 存储 cmd 6314 的个人领地排行数据。
// 数据结构：[[player_pos, alliance_id, "territory_ids_csv"], ...]
// 数组顺序即为排名（index+1）。
type PlayerTerritoryRank struct {
	ID             int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Rank           int    `json:"rank" gorm:"column:rank;index"`
	PlayerPos      int64  `json:"player_pos" gorm:"column:player_pos;uniqueIndex:idx_ptr_pos_cmd"`
	AllianceID     int64  `json:"alliance_id" gorm:"column:alliance_id;index"`
	TerritoryIDs   string `json:"territory_ids" gorm:"column:territory_ids;type:text"`
	TerritoryCount int    `json:"territory_count" gorm:"column:territory_count;index"`
	SourceCmd      int    `json:"source_cmd" gorm:"column:source_cmd;uniqueIndex:idx_ptr_pos_cmd"`
	CaptureTime    int64  `json:"capture_time" gorm:"column:capture_time;index"`
}

func (PlayerTerritoryRank) TableName() string { return "player_territory_rank" }

func SavePlayerTerritoryRankFromDecoded(decoded string, sourceCmd int) {
	if Conn == nil {
		return
	}
	var arr []any
	if err := json.Unmarshal([]byte(decoded), &arr); err != nil {
		log.Printf("[territory-rank] JSON解析失败: %v", err)
		return
	}
	if len(arr) == 0 {
		return
	}
	now := time.Now().Unix()
	saved := 0
	for i, item := range arr {
		row, ok := item.([]any)
		if !ok || len(row) < 2 {
			continue
		}
		playerPos := toInt64(row[0])
		if playerPos == 0 {
			continue
		}
		allianceID := toInt64(row[1])
		var territoryIDs string
		if len(row) >= 3 {
			territoryIDs = strings.TrimSpace(toString(row[2]))
		}
		count := 0
		if territoryIDs != "" {
			count = strings.Count(territoryIDs, ",")+1
		}
		rec := PlayerTerritoryRank{
			Rank:           i + 1,
			PlayerPos:      playerPos,
			AllianceID:     allianceID,
			TerritoryIDs:   territoryIDs,
			TerritoryCount: count,
			SourceCmd:      sourceCmd,
			CaptureTime:    now,
		}
		if err := Conn.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "player_pos"}, {Name: "source_cmd"}},
			DoUpdates: clause.AssignmentColumns([]string{"rank", "alliance_id", "territory_ids", "territory_count", "capture_time"}),
		}).Create(&rec).Error; err != nil {
			log.Printf("[territory-rank] 写库失败 rank=%d pos=%d err=%v", i+1, playerPos, err)
		} else {
			saved++
		}
	}
	log.Printf("[territory-rank] cmd6314 处理完成 saved=%d total=%d", saved, len(arr))
}
