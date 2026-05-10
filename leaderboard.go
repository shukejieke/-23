package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"stzbHelper/http/common"
	"stzbHelper/model"
)

func UnionLeaderboardList(c *gin.Context) {
	limit := 50
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	var rows []model.UnionLeaderboard
	q := model.Conn.Order("`rank` asc").Limit(limit)
	if name := c.Query("name"); name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if err := q.Find(&rows).Error; err != nil {
		common.Response{Message: "查询排行榜失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: gin.H{"items": rows, "count": len(rows)}}.Success(c)
}

func PlayerTerritoryRankList(c *gin.Context) {
	limit := 100
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	type row struct {
		model.PlayerTerritoryRank
		PlayerName string `json:"player_name" gorm:"column:player_name"`
	}

	q := model.Conn.Table("player_territory_rank ptr").
		Select("ptr.*, tu.name AS player_name").
		Joins("LEFT JOIN team_user tu ON tu.pos = ptr.player_pos").
		Order("ptr.rank asc").
		Limit(limit)
	var rows []row
	if err := q.Find(&rows).Error; err != nil {
		common.Response{Message: "查询失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: gin.H{"items": rows, "count": len(rows)}}.Success(c)
}

func PersonalLeaderboardList(c *gin.Context) {
	limit := 200
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}

	type row struct {
		model.PersonalLeaderboard
		PlayerName string `json:"player_name" gorm:"column:player_name"`
	}

	q := model.Conn.Table("personal_leaderboard pl").
		Select("pl.*, tu.name AS player_name").
		Joins("LEFT JOIN team_user tu ON tu.pos = pl.object_id").
		Order("pl.capture_time desc, pl.event_id asc").
		Limit(limit)
	if eventID := c.Query("event_id"); eventID != "" {
		q = q.Where("pl.event_id = ?", eventID)
	}
	if objID := c.Query("object_id"); objID != "" {
		q = q.Where("pl.object_id = ?", objID)
	}
	var rows []row
	if err := q.Find(&rows).Error; err != nil {
		common.Response{Message: "查询失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: gin.H{"items": rows, "count": len(rows)}}.Success(c)
}
