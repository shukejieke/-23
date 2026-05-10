package service

import (
	"stzbHelper/internal/repo"
	"stzbHelper/model"

	"gorm.io/gorm"
)

// ReportListFilter 是战报列表查询条件。
// 约定：字段为空表示不参与过滤。
type ReportListFilter struct {
	NextID       int
	AtkName      string
	AtkUnionName string
	AtkHP        string
	AtkLevel     string
	AtkStar      string
	Type         string
	NoNPC        bool
}

// QueryReportList 查询战报分页列表 + 总条数。
// 说明：
// - 默认 Type=1（攻守都查）
// - 每页固定 30 条
func QueryReportList(f ReportListFilter) ([]model.BattleReport, int64, error) {
	var reportList []model.BattleReport
	query := repo.DB().Model(&model.BattleReport{}).Limit(30).Order("time DESC")
	if f.NextID > 0 {
		query = query.Where("id < ?", f.NextID)
	}

	sType := f.Type
	if sType == "" {
		sType = "1"
	}

	switch sType {
	case "1":
		if f.AtkName != "" {
			query = query.Where("attack_name LIKE ? OR defend_name LIKE ?", "%"+f.AtkName+"%", "%"+f.AtkName+"%")
		}
		if f.AtkUnionName != "" {
			query = query.Where("attack_union_name LIKE ? OR defend_union_name LIKE ?", "%"+f.AtkUnionName+"%", "%"+f.AtkUnionName+"%")
		}
		if f.AtkHP != "" {
			query = query.Where("attack_hp >= ? OR defend_hp >= ?", f.AtkHP, f.AtkHP)
		}
		if f.AtkLevel != "" {
			query = query.Where("(attack_hero1_level >= ? AND attack_hero2_level >= ? AND attack_hero3_level >= ?) OR (defend_hero1_level >= ? AND defend_hero2_level >= ? AND defend_hero3_level >= ?)", f.AtkLevel, f.AtkLevel, f.AtkLevel, f.AtkLevel, f.AtkLevel, f.AtkLevel)
		}
		if f.AtkStar != "" {
			query = query.Where("attack_total_star >= ? OR defend_total_star >= ?", f.AtkStar, f.AtkStar)
		}
	case "2":
		if f.AtkName != "" {
			query = query.Where("attack_name LIKE ?", "%"+f.AtkName+"%")
		}
		if f.AtkUnionName != "" {
			query = query.Where("attack_union_name LIKE ?", "%"+f.AtkUnionName+"%")
		}
		if f.AtkHP != "" {
			query = query.Where("attack_hp >= ?", f.AtkHP)
		}
		if f.AtkLevel != "" {
			query = query.Where("attack_hero1_level >= ? AND attack_hero2_level >= ? AND attack_hero3_level >= ?", f.AtkLevel, f.AtkLevel, f.AtkLevel)
		}
		if f.AtkStar != "" {
			query = query.Where("attack_total_star >= ?", f.AtkStar)
		}
	case "3":
		if f.AtkName != "" {
			query = query.Where("defend_name LIKE ?", "%"+f.AtkName+"%")
		}
		if f.AtkUnionName != "" {
			query = query.Where("defend_union_name LIKE ?", "%"+f.AtkUnionName+"%")
		}
		if f.AtkHP != "" {
			query = query.Where("defend_hp >= ?", f.AtkHP)
		}
		if f.AtkLevel != "" {
			query = query.Where("defend_hero1_level >= ? AND defend_hero2_level >= ? AND defend_hero3_level >= ?", f.AtkLevel, f.AtkLevel, f.AtkLevel)
		}
		if f.AtkStar != "" {
			query = query.Where("defend_total_star >= ?", f.AtkStar)
		}
	case "4":
		if f.AtkName != "" {
			query = query.Where("attack_name LIKE ? OR defend_name LIKE ?", "%"+f.AtkName+"%", "%"+f.AtkName+"%")
		}
		if f.AtkUnionName != "" {
			query = query.Where("attack_union_name LIKE ? OR defend_union_name LIKE ?", "%"+f.AtkUnionName+"%", "%"+f.AtkUnionName+"%")
		}
		if f.AtkHP != "" {
			query = query.Where("attack_hp >= ? AND defend_hp >= ?", f.AtkHP, f.AtkHP)
		}
		if f.AtkLevel != "" {
			query = query.Where("(attack_hero1_level >= ? AND attack_hero2_level >= ? AND attack_hero3_level >= ?) AND (defend_hero1_level >= ? AND defend_hero2_level >= ? AND defend_hero3_level >= ?)", f.AtkLevel, f.AtkLevel, f.AtkLevel, f.AtkLevel, f.AtkLevel, f.AtkLevel)
		}
		if f.AtkStar != "" {
			query = query.Where("attack_total_star >= ? AND defend_total_star >= ?", f.AtkStar, f.AtkStar)
		}
	}

	if f.NoNPC {
		query = query.Where("npc = 0")
	}

	countQuery := query.Session(&gorm.Session{})
	if f.NextID > 0 {
		countQuery = countQuery.Where("id < ?", f.NextID)
	}

	if err := query.Find(&reportList).Error; err != nil {
		return nil, 0, err
	}

	var count int64
	if err := countQuery.Offset(-1).Limit(-1).Order("").Count(&count).Error; err != nil {
		return nil, 0, err
	}
	return reportList, count, nil
}
