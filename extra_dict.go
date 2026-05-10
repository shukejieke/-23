package service

import (
	"log"

	"stzbHelper/model"
)

// HeroNameByID 从 hero_extra 表查武将名。
func HeroNameByID(id int64) string {
	return queryExtraName("hero_extra", id)
}

// SkillNameByID 从 skill_extra 表查技能名。
func SkillNameByID(id int64) string {
	return queryExtraName("skill_extra", id)
}

// GearNameByID 从 gear_extra 表查宝物名。
func GearNameByID(id int64) string {
	return queryExtraName("gear_extra", id)
}

// queryExtraName 通过主键从指定表查 name 字段，查不到或出错返回空字符串。
func queryExtraName(table string, id int64) string {
	if id == 0 {
		return ""
	}
	db := model.Conn
	if db == nil {
		return ""
	}
	type row struct {
		Name string `gorm:"column:name"`
	}
	var r row
	if err := db.Table(table).Select("name").Where("id = ?", id).First(&r).Error; err != nil {
		log.Printf("[extra-dict] 查询失败 table=%s id=%d err=%v", table, id, err)
		return ""
	}
	return r.Name
}
