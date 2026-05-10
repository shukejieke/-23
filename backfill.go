package service

import (
	"log"
	"strconv"
	"strings"

	"stzbHelper/model"
)

// BackfillBattleReportNames 回填 battle_report 表中名称字段为空的历史数据。
// 在 extra 表导入完成后调用。
func BackfillBattleReportNames() {
	db := model.Conn
	if db == nil {
		return
	}

	var rows []model.BattleReport
	if err := db.Where(
		"(attack_hero1_name = '' OR attack_hero1_name IS NULL) AND attack_hero1_id != 0",
	).Find(&rows).Error; err != nil {
		log.Printf("[backfill] 查询失败: %v", err)
		return
	}
	if len(rows) == 0 {
		log.Println("[backfill] 无需回填，所有记录名称字段已有值")
		return
	}
	log.Printf("[backfill] 开始回填 %d 条记录", len(rows))
	ok := 0
	for _, r := range rows {
		updates := map[string]any{
			"attack_hero1_name":  HeroNameByID(r.AttackHero1Id),
			"attack_hero2_name":  HeroNameByID(r.AttackHero2Id),
			"attack_hero3_name":  HeroNameByID(r.AttackHero3Id),
			"defend_hero1_name":  HeroNameByID(r.DefendHero1Id),
			"defend_hero2_name":  HeroNameByID(r.DefendHero2Id),
			"defend_hero3_name":  HeroNameByID(r.DefendHero3Id),
			"attacker_gear_names": backfillGearNames(r.AttackerGearInfo),
			"defender_gear_names": backfillGearNames(r.DefenderGearInfo),
			"all_skill_names":    backfillSkillNames(r.AllSkillInfo),
		}
		if err := db.Model(&model.BattleReport{}).Where("battle_id = ?", r.BattleId).Updates(updates).Error; err != nil {
			log.Printf("[backfill] 更新失败 battle_id=%d: %v", r.BattleId, err)
		} else {
			ok++
		}
	}
	log.Printf("[backfill] 回填完成 ok=%d total=%d", ok, len(rows))
}

func backfillGearNames(raw string) string {
	if raw == "" {
		return ""
	}
	var names []string
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.SplitN(part, ",", 2)
		id, err := strconv.ParseInt(strings.TrimSpace(fields[0]), 10, 64)
		if err != nil || id == 0 {
			names = append(names, "-")
			continue
		}
		name := GearNameByID(id)
		if name == "" {
			name = fields[0]
		}
		names = append(names, name)
	}
	return strings.Join(names, ";")
}

func backfillSkillNames(raw string) string {
	if raw == "" {
		return ""
	}
	var slotResults []string
	for _, slot := range strings.Split(raw, ";") {
		slot = strings.TrimSpace(slot)
		if slot == "" {
			continue
		}
		parts := strings.Split(slot, ",")
		var names []string
		for _, idx := range []int{1, 3, 5} {
			if idx >= len(parts) {
				break
			}
			id, err := strconv.ParseInt(strings.TrimSpace(parts[idx]), 10, 64)
			if err != nil || id == 0 {
				continue
			}
			name := SkillNameByID(id)
			if name == "" {
				name = parts[idx]
			}
			names = append(names, name)
		}
		if len(names) > 0 {
			slotResults = append(slotResults, strings.Join(names, "/"))
		}
	}
	return strings.Join(slotResults, ";")
}
