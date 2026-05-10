package api

import (
	"github.com/gin-gonic/gin"
	"stzbHelper/http/handle/api"
)

func Register(r *gin.RouterGroup) {
	// 获取同盟成员列表
	r.Any("getTeamUser", api.GetTeamUser)
	// 获取同盟成员分组列表
	r.Any("getTeamGroup", api.GetTeamGroup)
	// 获取任务列表
	r.Any("getTaskList", api.GetTaskList)
	// 获取任务详情
	r.Any("getTask/:tid", api.GetTask)
	// 创建任务
	r.POST("createTask", api.CreateTask)
	// 删除任务
	r.Any("deleteTask/:tid", api.DelTask)
	// 开启获取战报
	r.POST("enable/getReport", api.EnableGetReport)
	// 关闭获取战报
	r.Any("disable/getReport", api.DisableGetReport)
	// 获取战报数据
	r.Any("getReportNumByTaskId/:tid", api.GetReportNumByTaskId)
	// 统计考勤数据
	r.Any("statisticsReport/:tid", api.StatisticsReport)
	r.Any("getGroupWu", api.GetGroupWu)
	// 删除任务战报
	r.Any("deleteTaskReport/:tid", api.DelTaskReport)
	r.GET("stzb/report/list", api.ReportList)
	r.GET("stzb/report/list/cn", api.ReportListCN)
	r.GET("stzb/report/stats", api.BattleStats)
	r.GET("stzb/player/team/get", api.GetPlayerTeam)
	// 开启获取战报详情
	r.Any("enable/getBattleReport", api.EnableGetBattleData)
	// 关闭获取战报详情
	r.Any("disable/getBattleReport", api.DisableGetBattleData)
	// 运行时状态（抓包/解析/绑定IP）
	r.GET("stzb/runtime/status", api.RuntimeStatus)
	// 抓包阶段观测到的全量协议号列表
	r.GET("stzb/runtime/cmd/list", api.RuntimeCmdList)
	// cmd_id 自动语义分析（用于人工确认）
	r.GET("stzb/runtime/cmd/analysis", api.RuntimeCmdAnalysis)
	// 网络流量协议目录（文档视图）
	r.GET("stzb/runtime/traffic/apis", api.RuntimeTrafficAPIs)
	// 业务数据可得性面板
	r.GET("stzb/runtime/business/availability", api.RuntimeBusinessAvailability)
	// 103/92 原始包归档
	r.GET("stzb/runtime/raw-capture/list", api.RuntimeRawCaptureList)
	// 同盟成员 / 分组抓取诊断
	r.GET("stzb/runtime/member/diagnosis", api.RuntimeMemberDiagnosis)
	// cmd_id 人工确认语义写回
	r.POST("stzb/runtime/cmd/confirm", api.RuntimeCmdConfirm)
	// cmd_id 忽略列表
	r.GET("stzb/runtime/cmd/ignore/list", api.RuntimeCmdIgnoreList)
	// cmd_id 忽略开关
	r.POST("stzb/runtime/cmd/ignore/set", api.RuntimeCmdIgnoreSet)
	// 数据集新鲜度 + 游戏内触发建议
	r.GET("stzb/runtime/freshness", api.RuntimeFreshness)
	// 排行榜场景抓包留档（700/514 前后窗口）
	r.GET("stzb/runtime/rank/trace", api.RuntimeRankTrace)
	// 同盟排行榜看板数据（cmd_id=700）
	r.GET("stzb/leaderboard/union", api.UnionLeaderboardList)
	// 玩家外观批量快照看板数据（cmd_id=514，历史沿用 personal_leaderboard 表）
	r.GET("stzb/leaderboard/personal", api.PersonalLeaderboardList)
	// 个人领地排行榜（cmd_id=6314）
	r.GET("stzb/leaderboard/territory", api.PlayerTerritoryRankList)
	// 战报多维聚合分析
	r.GET("stzb/battle/analysis/hero-stats", api.BattleHeroStats)
	r.GET("stzb/battle/analysis/hero-combo", api.BattleHeroComboStats)
	r.GET("stzb/battle/analysis/alliance-stats", api.BattleAllianceStats)
}
