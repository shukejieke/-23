package service

import (
	"stzbHelper/internal/repo"
	"stzbHelper/model"
)

type NamedCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type BattleStats struct {
	Total           int64        `json:"total"`
	NonNPC          int64        `json:"non_npc"`
	AttackWin       int64        `json:"attack_win"`
	Draw            int64        `json:"draw"`
	DefendWin       int64        `json:"defend_win"`
	TopAttackers    []NamedCount `json:"top_attackers"`
	TopDefenders    []NamedCount `json:"top_defenders"`
	TopAttackUnions []NamedCount `json:"top_attack_unions"`
	TopLocations    []NamedCount `json:"top_locations"`
}

func QueryBattleStats(limit int) (BattleStats, error) {
	if limit <= 0 || limit > 20 {
		limit = 8
	}
	var stats BattleStats
	db := repo.DB().Model(&model.BattleReport{})

	if err := db.Count(&stats.Total).Error; err != nil {
		return stats, err
	}
	if err := db.Where("npc = 0").Count(&stats.NonNPC).Error; err != nil {
		return stats, err
	}
	if err := db.Where("result IN ?", []int{1, 2}).Count(&stats.AttackWin).Error; err != nil {
		return stats, err
	}
	if err := db.Where("result = ?", 6).Count(&stats.Draw).Error; err != nil {
		return stats, err
	}
	if err := db.Where("result NOT IN ?", []int{1, 2, 6}).Count(&stats.DefendWin).Error; err != nil {
		return stats, err
	}

	loadTop := func(dest *[]NamedCount, col string) error {
		return db.
			Select(col + " AS name, COUNT(*) AS count").
			Where(col + " <> ''").
			Group(col).
			Order("count DESC, " + col + " ASC").
			Limit(limit).
			Scan(dest).Error
	}

	if err := loadTop(&stats.TopAttackers, "attack_name"); err != nil {
		return stats, err
	}
	if err := loadTop(&stats.TopDefenders, "defend_name"); err != nil {
		return stats, err
	}
	if err := loadTop(&stats.TopAttackUnions, "attack_union_name"); err != nil {
		return stats, err
	}
	if err := loadTop(&stats.TopLocations, "wid_name"); err != nil {
		return stats, err
	}
	return stats, nil
}
