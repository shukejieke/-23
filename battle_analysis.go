package service

import (
	"fmt"
	"strings"

	"stzbHelper/internal/repo"
	"stzbHelper/model"
)

// HeroStat 单个武将出场/胜场统计
type HeroStat struct {
	HeroID   int64  `json:"hero_id"`
	HeroName string `json:"hero_name"`
	Side     string `json:"side"` // attack / defend
	Count    int64  `json:"count"`
	WinCount int64  `json:"win_count"`
	WinRate  string `json:"win_rate"`
}

// ComboStat 武将组合出场/胜场统计
type ComboStat struct {
	Combo    string   `json:"combo"`          // 原始 ID 组合，如 "100134+100442+100587"
	Names    []string `json:"names"`          // 每个武将的名字
	Side     string   `json:"side"`
	Count    int64    `json:"count"`
	WinCount int64    `json:"win_count"`
	WinRate  string   `json:"win_rate"`
}

// AllianceStat 对手同盟出场统计
type AllianceStat struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// HeroAppearStats 武将出场次数与胜率，side="attack"/"defend"
func HeroAppearStats(side string, limit int) ([]HeroStat, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	db := repo.DB().Model(&model.BattleReport{})

	type row struct {
		HeroID   int64 `gorm:"column:hero_id"`
		Count    int64 `gorm:"column:count"`
		WinCount int64 `gorm:"column:win_count"`
	}

	var winCond string
	var cols [3]string
	if side == "defend" {
		cols = [3]string{"defend_hero1_id", "defend_hero2_id", "defend_hero3_id"}
		winCond = "result NOT IN (1,2,6)"
	} else {
		side = "attack"
		cols = [3]string{"attack_hero1_id", "attack_hero2_id", "attack_hero3_id"}
		winCond = "result IN (1,2)"
	}

	// UNION three hero slots
	union := fmt.Sprintf(
		"(SELECT %[1]s AS hero_id FROM battle_report WHERE %[1]s != 0) UNION ALL "+
			"(SELECT %[2]s AS hero_id FROM battle_report WHERE %[2]s != 0) UNION ALL "+
			"(SELECT %[3]s AS hero_id FROM battle_report WHERE %[3]s != 0)",
		cols[0], cols[1], cols[2],
	)
	winUnion := fmt.Sprintf(
		"(SELECT %[1]s AS hero_id FROM battle_report WHERE %[1]s != 0 AND %[4]s) UNION ALL "+
			"(SELECT %[2]s AS hero_id FROM battle_report WHERE %[2]s != 0 AND %[4]s) UNION ALL "+
			"(SELECT %[3]s AS hero_id FROM battle_report WHERE %[3]s != 0 AND %[4]s)",
		cols[0], cols[1], cols[2], winCond,
	)

	// Total appearance per hero
	var totals []row
	if err := db.Raw(fmt.Sprintf(
		"SELECT hero_id, COUNT(*) AS count FROM (%s) t GROUP BY hero_id ORDER BY count DESC LIMIT ?",
		union,
	), limit).Scan(&totals).Error; err != nil {
		return nil, err
	}

	// Win counts per hero
	var wins []row
	_ = db.Raw(fmt.Sprintf(
		"SELECT hero_id, COUNT(*) AS win_count FROM (%s) t GROUP BY hero_id",
		winUnion,
	)).Scan(&wins)
	winMap := map[int64]int64{}
	for _, w := range wins {
		winMap[w.HeroID] = w.WinCount
	}

	result := make([]HeroStat, 0, len(totals))
	for _, t := range totals {
		winC := winMap[t.HeroID]
		rate := "0%"
		if t.Count > 0 {
			rate = fmt.Sprintf("%.1f%%", float64(winC)/float64(t.Count)*100)
		}
		result = append(result, HeroStat{
			HeroID:   t.HeroID,
			HeroName: HeroNameByID(t.HeroID),
			Side:     side,
			Count:    t.Count,
			WinCount: winC,
			WinRate:  rate,
		})
	}
	return result, nil
}

// HeroComboStats 武将组合出场次数与胜率，side="attack"/"defend"
func HeroComboStats(side string, limit int) ([]ComboStat, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	db := repo.DB().Model(&model.BattleReport{})

	var comboCol, winCond string
	if side == "defend" {
		comboCol = "defend_hero_combo"
		winCond = "result NOT IN (1,2,6)"
	} else {
		side = "attack"
		comboCol = "attack_hero_combo"
		winCond = "result IN (1,2)"
	}

	type comboRow struct {
		Combo    string `gorm:"column:combo"`
		Count    int64  `gorm:"column:count"`
		WinCount int64  `gorm:"column:win_count"`
	}
	var rows []comboRow
	if err := db.Raw(fmt.Sprintf(
		"SELECT %[1]s AS combo, COUNT(*) AS count, SUM(CASE WHEN %[2]s THEN 1 ELSE 0 END) AS win_count "+
			"FROM battle_report WHERE %[1]s != '' GROUP BY %[1]s ORDER BY count DESC LIMIT ?",
		comboCol, winCond,
	), limit).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]ComboStat, 0, len(rows))
	for _, r := range rows {
		names := comboNames(r.Combo)
		rate := "0%"
		if r.Count > 0 {
			rate = fmt.Sprintf("%.1f%%", float64(r.WinCount)/float64(r.Count)*100)
		}
		result = append(result, ComboStat{
			Combo:    r.Combo,
			Names:    names,
			Side:     side,
			Count:    r.Count,
			WinCount: r.WinCount,
			WinRate:  rate,
		})
	}
	return result, nil
}

// AllianceEncounterStats 与哪些同盟交战最多（对手视角）
func AllianceEncounterStats(limit int) ([]AllianceStat, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	db := repo.DB().Model(&model.BattleReport{})

	// Combine both sides' alliance names (opponents from any side)
	var rows []AllianceStat
	err := db.Raw(
		"SELECT name, SUM(cnt) AS count FROM ("+
			"SELECT attack_union_name AS name, COUNT(*) AS cnt FROM battle_report WHERE attack_union_name != '' GROUP BY attack_union_name UNION ALL "+
			"SELECT defend_union_name AS name, COUNT(*) AS cnt FROM battle_report WHERE defend_union_name != '' GROUP BY defend_union_name"+
			") t GROUP BY name ORDER BY count DESC LIMIT ?",
		limit,
	).Scan(&rows).Error
	return rows, err
}

func comboNames(combo string) []string {
	parts := strings.Split(combo, "+")
	names := make([]string, 0, len(parts))
	for _, p := range parts {
		var id int64
		fmt.Sscan(p, &id)
		name := HeroNameByID(id)
		if name == "" {
			name = p
		}
		names = append(names, name)
	}
	return names
}
