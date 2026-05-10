package api

import (
	"github.com/gin-gonic/gin"

	"stzbHelper/http/common"
	"stzbHelper/internal/service"
)

func GetTeamUser(c *gin.Context) {
	teamUsers, err := service.GetTeamUsers(c.Query("group"))
	if err != nil {
		common.Response{Message: "查询失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: teamUsers}.Success(c)
}

func GetTeamGroup(c *gin.Context) {
	groups, err := service.GetTeamGroups()
	if err != nil {
		common.Response{Message: "查询失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: groups}.Success(c)
}

func GetGroupWu(c *gin.Context) {
	stats, err := service.GetGroupWuStats()
	if err != nil {
		common.Response{Message: "查询失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: stats}.Success(c)
}
