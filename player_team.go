package service

import "stzbHelper/internal/repo"

type PlayerTeamRow struct {
	PlayerName   string `json:"player_name" gorm:"player_name"`
	BattleID     int    `json:"battle_id" gorm:"battle_id"`
	Hero1ID      int    `json:"hero1_id" gorm:"hero1_id"`
	Hero2ID      int    `json:"hero2_id" gorm:"hero2_id"`
	Hero3ID      int    `json:"hero3_id" gorm:"hero3_id"`
	Hero1Name    string `json:"hero1_name" gorm:"hero1_name"`
	Hero2Name    string `json:"hero2_name" gorm:"hero2_name"`
	Hero3Name    string `json:"hero3_name" gorm:"hero3_name"`
	Hero1Level   int    `json:"hero1_level" gorm:"hero1_level"`
	Hero2Level   int    `json:"hero2_level" gorm:"hero2_level"`
	Hero3Level   int    `json:"hero3_level" gorm:"hero3_level"`
	Hero1Star    int    `json:"hero1_star" gorm:"hero1_star"`
	Hero2Star    int    `json:"hero2_star" gorm:"hero2_star"`
	Hero3Star    int    `json:"hero3_star" gorm:"hero3_star"`
	TotalStar    int    `json:"total_star" gorm:"total_star"`
	Hp           int    `json:"hp" gorm:"hp"`
	AllSkillInfo string `json:"all_skill_info" gorm:"all_skill_info"`
	AllSkillNames string `json:"all_skill_names" gorm:"all_skill_names"`
	Role         string `json:"role" gorm:"role"`
	Time         int    `json:"time" gorm:"time"`
	Gear         string `json:"gear" gorm:"gear"`
	GearNames    string `json:"gear_names" gorm:"gear_names"`
	HeroType     string `json:"hero_type" gorm:"hero_type"`
	Idu          string `json:"idu" gorm:"idu"`
}

func QueryPlayerTeams(name, unionName, idu string) ([]PlayerTeamRow, error) {
	var results []PlayerTeamRow
	likeName := "%" + name + "%"
	likeUnion := "%" + unionName + "%"
	likeIdu := "%" + idu + "%"

	query := `WITH ranked_data AS (
		SELECT
			attack_name AS player_name,
			attack_hero1_id AS hero1_id,
			attack_hero2_id AS hero2_id,
			attack_hero3_id AS hero3_id,
			attack_hero1_name AS hero1_name,
			attack_hero2_name AS hero2_name,
			attack_hero3_name AS hero3_name,
			attack_hero1_level AS hero1_level,
			attack_hero2_level AS hero2_level,
			attack_hero3_level AS hero3_level,
			attack_hero1_star AS hero1_star,
			attack_hero2_star AS hero2_star,
			attack_hero3_star AS hero3_star,
			attack_total_star AS total_star,
			attack_hp AS hp,
			attacker_gear_info AS gear,
			attacker_gear_names AS gear_names,
			attack_hero_type AS hero_type,
			attack_idu AS idu,
			time,
			all_skill_info,
			all_skill_names,
			battle_id,
			'attack' AS role,
			ROW_NUMBER() OVER (PARTITION BY attack_name, attack_hero1_id ORDER BY attack_hero1_level DESC, time DESC) AS rn
		FROM battle_report
		WHERE attack_hero1_id != 0 AND attack_hero2_id != 0 AND attack_hero3_id != 0
			AND attack_hero1_level >= 15 AND attack_hero2_level >= 15 AND attack_hero3_level >= 15
			AND attack_hp >= 10000
			AND attack_name LIKE ?
			AND attack_union_name LIKE ?
			AND attack_idu LIKE ?
			AND npc = 0
			AND all_skill_info != '' AND all_skill_info IS NOT NULL

		UNION ALL

		SELECT
			defend_name AS player_name,
			defend_hero1_id AS hero1_id,
			defend_hero2_id AS hero2_id,
			defend_hero3_id AS hero3_id,
			defend_hero1_name AS hero1_name,
			defend_hero2_name AS hero2_name,
			defend_hero3_name AS hero3_name,
			defend_hero1_level AS hero1_level,
			defend_hero2_level AS hero2_level,
			defend_hero3_level AS hero3_level,
			defend_hero1_star AS hero1_star,
			defend_hero2_star AS hero2_star,
			defend_hero3_star AS hero3_star,
			defend_total_star AS total_star,
			defend_hp AS hp,
			defender_gear_info AS gear,
			defender_gear_names AS gear_names,
			defend_hero_type AS hero_type,
			defend_idu AS idu,
			time,
			all_skill_info,
			all_skill_names,
			battle_id,
			'defend' AS role,
			ROW_NUMBER() OVER (PARTITION BY defend_name, defend_hero1_id ORDER BY defend_hero1_level DESC, time DESC) AS rn
		FROM battle_report
		WHERE defend_hero1_id != 0 AND defend_hero2_id != 0 AND defend_hero3_id != 0
			AND defend_hero1_level >= 15 AND defend_hero2_level >= 15 AND defend_hero3_level >= 15
			AND defend_hp >= 10000
			AND defend_name LIKE ?
			AND defend_union_name LIKE ?
			AND defend_idu LIKE ?
			AND npc = 0
			AND all_skill_info != '' AND all_skill_info IS NOT NULL
	), deduplicated_data AS (
		SELECT
			player_name, hero1_id, hero2_id, hero3_id,
			hero1_name, hero2_name, hero3_name,
			hero1_level, hero2_level, hero3_level,
			hero1_star, hero2_star, hero3_star,
			total_star, hp, gear, gear_names, hero_type, idu,
			time, all_skill_info, all_skill_names, battle_id, role,
			ROW_NUMBER() OVER (PARTITION BY player_name, hero1_id, hero2_id, hero3_id ORDER BY time DESC) AS dedup_rn
		FROM ranked_data
		WHERE rn = 1
	)
	SELECT
		player_name, hero1_id, hero2_id, hero3_id,
		hero1_name, hero2_name, hero3_name,
		hero1_level, hero2_level, hero3_level,
		hero1_star, hero2_star, hero3_star,
		total_star, hp, gear, gear_names, hero_type, idu,
		time, all_skill_info, all_skill_names, battle_id, role
	FROM deduplicated_data
	WHERE dedup_rn = 1
	ORDER BY player_name, time DESC`

	err := repo.DB().Raw(query,
		likeName, likeUnion, likeIdu,
		likeName, likeUnion, likeIdu,
	).Scan(&results).Error
	return results, err
}
