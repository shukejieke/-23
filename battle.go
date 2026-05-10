package api

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"

	"stzbHelper/http/common"
	"stzbHelper/internal/service"
	"stzbHelper/model"
)

func ReportList(c *gin.Context) {
	nextid := c.Query("nextid")
	if nextid == "" {
		common.Response{Message: "参数错误"}.Error(c)
		return
	}
	nextID, _ := strconv.Atoi(nextid)
	log.Printf("[api] report/list 请求 nextid=%s atkname=%s atkunionname=%s type=%s nonpc=%s", nextid, c.Query("atkname"), c.Query("atkunionname"), c.Query("type"), c.Query("nonpc"))

	reports, total, err := service.QueryReportList(service.ReportListFilter{
		NextID:       nextID,
		AtkName:      c.Query("atkname"),
		AtkUnionName: c.Query("atkunionname"),
		AtkHP:        c.Query("atkhp"),
		AtkLevel:     c.Query("atklevel"),
		AtkStar:      c.Query("atkstar"),
		Type:         c.Query("type"),
		NoNPC:        c.Query("nonpc") == "1",
	})
	if err != nil {
		common.Response{Message: "查询失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: gin.H{"report": reports, "total": total}}.Success(c)
}

// ReportListCN 返回中文字段的战报列表，便于前端直接展示。
func ReportListCN(c *gin.Context) {
	nextid := c.Query("nextid")
	if nextid == "" {
		common.Response{Message: "参数错误"}.Error(c)
		return
	}
	nextID, _ := strconv.Atoi(nextid)
	log.Printf("[api] report/list/cn 请求 nextid=%s atkname=%s atkunionname=%s type=%s nonpc=%s", nextid, c.Query("atkname"), c.Query("atkunionname"), c.Query("type"), c.Query("nonpc"))

	reports, total, err := service.QueryReportList(service.ReportListFilter{
		NextID:       nextID,
		AtkName:      c.Query("atkname"),
		AtkUnionName: c.Query("atkunionname"),
		AtkHP:        c.Query("atkhp"),
		AtkLevel:     c.Query("atklevel"),
		AtkStar:      c.Query("atkstar"),
		Type:         c.Query("type"),
		NoNPC:        c.Query("nonpc") == "1",
	})
	if err != nil {
		common.Response{Message: "查询失败: " + err.Error()}.Error(c)
		return
	}

	cn := make([]gin.H, 0, len(reports))
	for _, r := range reports {
		cn = append(cn, reportToCN(r))
	}
	common.Response{Data: gin.H{"战报": cn, "总数": total}}.Success(c)
}

func GetPlayerTeam(c *gin.Context) {
	name := c.Query("atkname")
	uname := c.Query("atkunionname")
	idu := c.Query("idu")
	log.Printf("[api] player/team/get 请求 atkname=%s atkunionname=%s idu=%s", name, uname, idu)

	rows, err := service.QueryPlayerTeams(name, uname, idu)
	if err != nil {
		common.Response{Message: "查询失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: rows}.Success(c)
}

func BattleStats(c *gin.Context) {
	limit := 8
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 20 {
			limit = n
		}
	}
	stats, err := service.QueryBattleStats(limit)
	if err != nil {
		common.Response{Message: "查询战报统计失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: stats}.Success(c)
}

func reportToCN(r model.BattleReport) gin.H {
	resultText := "平局"
	if r.Result > 0 {
		resultText = "进攻胜"
	} else if r.Result < 0 {
		resultText = "防守胜"
	}

	atkHeroNames := []string{
		service.HeroNameByID(r.AttackHero1Id),
		service.HeroNameByID(r.AttackHero2Id),
		service.HeroNameByID(r.AttackHero3Id),
	}
	defHeroNames := []string{
		service.HeroNameByID(r.DefendHero1Id),
		service.HeroNameByID(r.DefendHero2Id),
		service.HeroNameByID(r.DefendHero3Id),
	}

	return gin.H{
		"ID":     r.ID,
		"战报ID":   r.BattleId,
		"发生时间":   r.Time,
		"战斗坐标":   r.Wid,
		"地块名称":   r.WidName,
		"进攻方":    r.AttackName,
		"进攻方同盟":  r.AttackUnionName,
		"防守方":    r.DefendName,
		"防守方同盟":  r.DefendUnionName,
		"进攻兵力":   r.AttackHp,
		"防守兵力":   r.DefendHp,
		"是否NPC":  r.Npc,
		"战斗结果":   r.Result,
		"战斗结果文本": resultText,
		"进攻队伍标识": r.AttackIdu,
		"防守队伍标识": r.DefendIdu,
		"进攻武将ID": []int64{r.AttackHero1Id, r.AttackHero2Id, r.AttackHero3Id},
		"防守武将ID": []int64{r.DefendHero1Id, r.DefendHero2Id, r.DefendHero3Id},
		"进攻武将名称": atkHeroNames,
		"防守武将名称": defHeroNames,
		"进攻总红度":  r.AttackTotalStar,
		"防守总红度":  r.DefendTotalStar,
		"技能信息":   r.AllSkillInfo,
	}
}

// BattleHeroStats 武将出场/胜率统计
func BattleHeroStats(c *gin.Context) {
	side := c.DefaultQuery("side", "attack")
	limit := 20
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}
	rows, err := service.HeroAppearStats(side, limit)
	if err != nil {
		common.Response{Message: "查询失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: rows}.Success(c)
}

// BattleHeroComboStats 武将组合出场/胜率统计
func BattleHeroComboStats(c *gin.Context) {
	side := c.DefaultQuery("side", "attack")
	limit := 20
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}
	rows, err := service.HeroComboStats(side, limit)
	if err != nil {
		common.Response{Message: "查询失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: rows}.Success(c)
}

// BattleAllianceStats 对手同盟出场统计
func BattleAllianceStats(c *gin.Context) {
	limit := 20
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}
	rows, err := service.AllianceEncounterStats(limit)
	if err != nil {
		common.Response{Message: "查询失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: rows}.Success(c)
}

func Example(c *gin.Context) {
	common.Response{Message: "This is example func"}.Success(c)
}
