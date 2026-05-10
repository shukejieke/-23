package api

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"stzbHelper/http/common"
	"stzbHelper/internal/capture"
	"stzbHelper/internal/parser"
	"stzbHelper/internal/runtime"
	"stzbHelper/model"
)

// RuntimeStatus 用于排查“抓不到数据”问题。
func RuntimeStatus(c *gin.Context) {
	seen, parsed, selected := capture.Stats()
	cmd103, cmd92, teamUsers, reports, battles := parser.Stats()
	src, dst := runtime.GetBoundIPs()
	cmdStats := capture.CmdStats()
	freshness := runtime.FreshnessSnapshot()
	topN := 20
	if len(cmdStats) < topN {
		topN = len(cmdStats)
	}

	common.Response{Data: gin.H{
		"packet_seen":         seen,
		"parse_triggered":     parsed,
		"database_selected":   selected,
		"bound_src":           src,
		"bound_dst":           dst,
		"cmd103_count":        cmd103,
		"cmd92_count":         cmd92,
		"teamuser_save_count": teamUsers,
		"report_save_count":   reports,
		"battle_save_count":   battles,
		"cmd_observed_types":  len(cmdStats),
		"cmd_observed_top":    cmdStats[:topN],
		"dataset_freshness":   freshness,
	}}.Success(c)
}

// RuntimeCmdList 返回抓包阶段观测到的全量协议号列表。
func RuntimeCmdList(c *gin.Context) {
	stats := capture.CmdStats()
	common.Response{Data: gin.H{
		"total_types": len(stats),
		"items":       stats,
	}}.Success(c)
}

// RuntimeFreshness 返回各数据集新鲜度与游戏内触发建议。
func RuntimeFreshness(c *gin.Context) {
	common.Response{Data: gin.H{
		"items": runtime.FreshnessSnapshot(),
	}}.Success(c)
}

// RuntimeRankTrace 返回排行榜场景抓包诊断留档（700/514 触发前后窗口）。
func RuntimeRankTrace(c *gin.Context) {
	limit := 20
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 120 {
			limit = n
		}
	}
	items := capture.RankTraceLatest(limit)
	common.Response{Data: gin.H{
		"items": items,
		"count": len(items),
		"limit": limit,
	}}.Success(c)
}

type apiExample struct {
	Method string `json:"method"`
	Route  string `json:"route"`
	Req    string `json:"req"`
	Resp   string `json:"resp"`
}

type cmdIgnoreReq struct {
	CmdID  int  `json:"cmd_id"`
	Ignore bool `json:"ignore"`
}

type cmdAnalysisItem struct {
	CmdID             int          `json:"cmd_id"`
	CmdName           string       `json:"cmd_name"`
	Mapped            bool         `json:"mapped"`
	DataType          int          `json:"data_type"`
	SeenCount         int64        `json:"seen_count"`
	FirstSeen         int64        `json:"first_seen"`
	UpdatedAt         int64        `json:"updated_at"`
	SamplePreview     string       `json:"sample_preview"`
	InferredSemantics string       `json:"inferred_semantics"`
	SuggestedName     string       `json:"suggested_name"`
	RequestType       string       `json:"request_type"`
	Ignored           bool         `json:"ignored"`
	APIs              []string     `json:"apis"`
	APIExamples       []apiExample `json:"api_examples"`
	Confidence        float64      `json:"confidence"`
	Basis             []string     `json:"basis"`
}

type trafficAPIItem struct {
	CmdID          int          `json:"cmd_id"`
	CmdName        string       `json:"cmd_name"`
	Domain         string       `json:"domain"`
	InterfacePath  string       `json:"interface_path"`
	Method         string       `json:"method"`
	Remark         string       `json:"remark"`
	RequestType    string       `json:"request_type"`
	ReqExample     string       `json:"req_example"`
	RawRespExample string       `json:"raw_resp_example"`
	RespExample    string       `json:"resp_example"`
	SeenCount      int64        `json:"seen_count"`
	UpdatedAt      int64        `json:"updated_at"`
	Confidence     float64      `json:"confidence"`
	Basis          []string     `json:"basis"`
	APIExamples    []apiExample `json:"api_examples"`
}

type businessAvailabilityItem struct {
	Key          string   `json:"key"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	Reason       string   `json:"reason"`
	Ready        bool     `json:"ready"`
	Counts       gin.H    `json:"counts"`
	Dependencies []string `json:"dependencies"`
}

type memberDiagnosisCapture struct {
	ID          int64  `json:"id"`
	CmdID       int    `json:"cmd_id"`
	CmdName     string `json:"cmd_name"`
	DataType    int    `json:"data_type"`
	Preview     string `json:"preview"`
	DecodedText string `json:"decoded_text"`
	RawHex      string `json:"raw_hex"`
	CaptureTime int64  `json:"capture_time"`
}

// RuntimeCmdAnalysis 返回 cmd_id 自动分析结果（用于后续人工确认语义）。
func RuntimeCmdAnalysis(c *gin.Context) {
	onlyUnknown := c.DefaultQuery("only_unknown", "1") == "1"
	onlyNotIgnored := c.DefaultQuery("only_not_ignored", "0") == "1"
	onlyIgnored := c.DefaultQuery("only_ignored", "0") == "1"
	cmdID := 0
	if s := strings.TrimSpace(c.Query("cmd_id")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n <= 0 {
			common.Response{Message: "参数错误: cmd_id 必须为正整数"}.Error(c)
			return
		}
		cmdID = n
	}
	limit := 100
	if s := c.Query("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n <= 0 || n > 500 {
			limit = 100
		} else {
			limit = n
		}
	}

	var rows []model.CmdSchema
	q := model.Conn.Order("seen_count DESC, cmd_id ASC").Limit(limit)
	if onlyUnknown {
		q = q.Where("cmd_name = ?", "未映射")
	}
	if cmdID > 0 {
		q = q.Where("cmd_id = ?", cmdID)
	}
	if err := q.Find(&rows).Error; err != nil {
		common.Response{Message: "查询cmd分析失败: " + err.Error()}.Error(c)
		return
	}

	items := make([]cmdAnalysisItem, 0, len(rows))
	for _, r := range rows {
		ignored := capture.IsCmdIgnored(r.CmdID)
		if onlyNotIgnored && ignored {
			continue
		}
		if onlyIgnored && !ignored {
			continue
		}
		guess, suggestName, reqType, confidence, basis := inferCmdSemantics(r)
		items = append(items, cmdAnalysisItem{
			CmdID:             r.CmdID,
			CmdName:           r.CmdName,
			Mapped:            r.CmdName != "未映射",
			DataType:          r.DataType,
			SeenCount:         r.SeenCount,
			FirstSeen:         r.FirstSeen,
			UpdatedAt:         r.UpdatedAt,
			SamplePreview:     previewText(r.SampleText, 400),
			InferredSemantics: guess,
			SuggestedName:     suggestName,
			RequestType:       reqType,
			Ignored:           ignored,
			APIs:              inferCmdAPIs(r.CmdID),
			APIExamples:       inferCmdAPIExamples(r.CmdID),
			Confidence:        confidence,
			Basis:             basis,
		})
	}

	common.Response{Data: gin.H{
		"items":            items,
		"count":            len(items),
		"only_unknown":     onlyUnknown,
		"only_not_ignored": onlyNotIgnored,
		"only_ignored":     onlyIgnored,
	}}.Success(c)
}

// RuntimeTrafficAPIs 返回已观测网络协议的“接口文档视图”。
func RuntimeTrafficAPIs(c *gin.Context) {
	limit := 200
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}
	onlyUnknown := c.DefaultQuery("only_unknown", "0") == "1"
	cmdID := 0
	if s := strings.TrimSpace(c.Query("cmd_id")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			cmdID = n
		}
	}

	src, _ := runtime.GetBoundIPs()
	domain := hostOnly(src)
	if domain == "" {
		domain = "game-tcp-8001"
	}

	var rows []model.CmdSchema
	q := model.Conn.Order("updated_at DESC, seen_count DESC, cmd_id ASC").Limit(limit)
	if onlyUnknown {
		q = q.Where("cmd_name = ?", "未映射")
	}
	if cmdID > 0 {
		q = q.Where("cmd_id = ?", cmdID)
	}
	if err := q.Find(&rows).Error; err != nil {
		common.Response{Message: "查询流量接口目录失败: " + err.Error()}.Error(c)
		return
	}

	items := make([]trafficAPIItem, 0, len(rows))
	for _, row := range rows {
		guess, suggestName, reqType, confidence, basis := inferCmdSemantics(row)
		apiExamples := inferCmdAPIExamples(row.CmdID)
		method := "TCP"
		reqExample := "需在游戏内触发，当前抓包链路无法直接构造二次请求。"
		interfacePath := "/game/cmd/" + strconv.Itoa(row.CmdID)
		if len(apiExamples) > 0 {
			method = apiExamples[0].Method
			interfacePath = apiExamples[0].Route
			reqExample = apiExamples[0].Req
		}
		remark := guess
		if remark == "" {
			remark = suggestName
		}
		if remark == "" {
			remark = row.CmdName
		}
		items = append(items, trafficAPIItem{
			CmdID:          row.CmdID,
			CmdName:        row.CmdName,
			Domain:         domain,
			InterfacePath:  interfacePath,
			Method:         method,
			Remark:         remark,
			RequestType:    reqType,
			ReqExample:     reqExample,
			RawRespExample: previewText(row.RawSampleHex, 4000),
			RespExample:    previewText(row.SampleText, 4000),
			SeenCount:      row.SeenCount,
			UpdatedAt:      row.UpdatedAt,
			Confidence:     confidence,
			Basis:          basis,
			APIExamples:    apiExamples,
		})
	}

	common.Response{Data: gin.H{
		"items":        items,
		"count":        len(items),
		"domain":       domain,
		"only_unknown": onlyUnknown,
	}}.Success(c)
}

// RuntimeRawCaptureList 返回 103/92 原始包归档，便于定位为什么业务表没出来。
func RuntimeRawCaptureList(c *gin.Context) {
	limit := 50
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	var rows []model.RawCapture
	q := model.Conn.Order("capture_time DESC, id DESC").Limit(limit)
	if s := strings.TrimSpace(c.Query("cmd_id")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			q = q.Where("cmd_id = ?", n)
		}
	}
	if err := q.Find(&rows).Error; err != nil {
		common.Response{Message: "查询原始包归档失败: " + err.Error()}.Error(c)
		return
	}
	common.Response{Data: gin.H{
		"items": rows,
		"count": len(rows),
		"limit": limit,
	}}.Success(c)
}

// RuntimeMemberDiagnosis 专门用于定位“同盟成员 / 分组武勋为什么还是不对”。
// 聚焦 100 / 103 / 142 三类包，以及 team_user / union_group_meta 当前表状态。
func RuntimeMemberDiagnosis(c *gin.Context) {
	limit := 30
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	countTable := func(table string) int64 {
		var n int64
		_ = model.Conn.Table(table).Count(&n).Error
		return n
	}
	countRaw := func(cmdID int) int64 {
		var n int64
		_ = model.Conn.Model(&model.RawCapture{}).Where("cmd_id = ?", cmdID).Count(&n).Error
		return n
	}

	var rawRows []model.RawCapture
	if err := model.Conn.
		Where("cmd_id IN ?", []int{100, 103, 142}).
		Order("capture_time DESC, id DESC").
		Limit(limit).
		Find(&rawRows).Error; err != nil {
		common.Response{Message: "查询成员诊断失败: " + err.Error()}.Error(c)
		return
	}

	items := make([]memberDiagnosisCapture, 0, len(rawRows))
	for _, row := range rawRows {
		items = append(items, memberDiagnosisCapture{
			ID:          row.ID,
			CmdID:       row.CmdID,
			CmdName:     row.CmdName,
			DataType:    row.DataType,
			Preview:     previewText(row.Preview, 400),
			DecodedText: previewText(row.DecodedText, 4000),
			RawHex:      previewText(row.RawHex, 3000),
			CaptureTime: row.CaptureTime,
		})
	}

	cmd103, cmd92, teamUsersSaved, _, _ := parser.Stats()
	common.Response{Data: gin.H{
		"summary": gin.H{
			"team_user_count":        countTable("team_user"),
			"union_group_meta_count": countTable("union_group_meta"),
			"raw_100_count":          countRaw(100),
			"raw_103_count":          countRaw(103),
			"raw_142_count":          countRaw(142),
			"cmd103_parse_triggered": cmd103,
			"teamuser_save_count":    teamUsersSaved,
			"cmd92_parse_triggered":  cmd92,
		},
		"items": items,
		"count": len(items),
		"limit": limit,
	}}.Success(c)
}

// RuntimeBusinessAvailability 返回功能复现所需业务数据的可得性状态。
func RuntimeBusinessAvailability(c *gin.Context) {
	type countResult struct {
		N int64
	}
	getCount := func(table string) int64 {
		var x countResult
		_ = model.Conn.Table(table).Count(&x.N).Error
		return x.N
	}

	raw103 := int64(0)
	raw92 := int64(0)
	_ = model.Conn.Model(&model.RawCapture{}).Where("cmd_id = ?", 103).Count(&raw103).Error
	_ = model.Conn.Model(&model.RawCapture{}).Where("cmd_id = ?", 92).Count(&raw92).Error
	cmd103, cmd92, _, _, _ := parser.Stats()

	teamUserCount := getCount("team_user")
	groupCount := getCount("union_group_meta")
	reportCount := getCount("reports")
	battleCount := getCount("battle_report")
	unionBoardCount := getCount("union_leaderboard")
	personalBoardCount := getCount("personal_leaderboard")

	items := []businessAvailabilityItem{
		{
			Key:          "team_member",
			Title:        "同盟成员",
			Status:       statusFor(teamUserCount >= 50 && raw103 > 0),
			Ready:        teamUserCount >= 50 && raw103 > 0,
			Reason:       reasonFor(teamUserCount >= 50 && raw103 > 0, "已具备成员主表", "未稳定抓到 103 成员主协议，当前成员表不足 50 条"),
			Counts:       gin.H{"team_user": teamUserCount, "raw_103": raw103, "cmd103_parse": cmd103},
			Dependencies: []string{"cmd=103", "table=team_user"},
		},
		{
			Key:          "group_meta",
			Title:        "分组元数据",
			Status:       statusFor(groupCount > 0),
			Ready:        groupCount > 0,
			Reason:       reasonFor(groupCount > 0, "已抓到分组元数据", "尚未抓到 142 分组元数据"),
			Counts:       gin.H{"union_group_meta": groupCount},
			Dependencies: []string{"cmd=142", "table=union_group_meta"},
		},
		{
			Key:          "report",
			Title:        "攻城战报",
			Status:       statusFor(reportCount > 0 && raw92 > 0),
			Ready:        reportCount > 0 && raw92 > 0,
			Reason:       reasonFor(reportCount > 0 && raw92 > 0, "已产出普通战报", "未稳定抓到 92 普通战报或未成功落库"),
			Counts:       gin.H{"reports": reportCount, "raw_92": raw92, "cmd92_parse": cmd92},
			Dependencies: []string{"cmd=92", "table=reports"},
		},
		{
			Key:          "battle_report",
			Title:        "详细战报",
			Status:       statusFor(battleCount > 0 && raw92 > 0),
			Ready:        battleCount > 0 && raw92 > 0,
			Reason:       reasonFor(battleCount > 0 && raw92 > 0, "已产出详细战报", "未开启或未成功抓到详细战报链路"),
			Counts:       gin.H{"battle_report": battleCount, "raw_92": raw92, "cmd92_parse": cmd92},
			Dependencies: []string{"cmd=92", "table=battle_report", "enable=getBattleReport"},
		},
		{
			Key:          "union_leaderboard",
			Title:        "同盟排行榜",
			Status:       statusFor(unionBoardCount > 0),
			Ready:        unionBoardCount > 0,
			Reason:       reasonFor(unionBoardCount > 0, "已抓到同盟排行榜", "尚未抓到 700"),
			Counts:       gin.H{"union_leaderboard": unionBoardCount},
			Dependencies: []string{"cmd=700", "table=union_leaderboard"},
		},
		{
			Key:          "personal_leaderboard",
			Title:        "个人排行榜附属数据",
			Status:       statusFor(personalBoardCount > 0),
			Ready:        personalBoardCount > 0,
			Reason:       reasonFor(personalBoardCount > 0, "已抓到 514 外观附属数据", "尚未抓到 514"),
			Counts:       gin.H{"personal_leaderboard": personalBoardCount},
			Dependencies: []string{"cmd=514", "table=personal_leaderboard"},
		},
	}

	common.Response{Data: gin.H{
		"items": items,
		"summary": gin.H{
			"team_user_count":        teamUserCount,
			"union_group_count":      groupCount,
			"report_count":           reportCount,
			"battle_count":           battleCount,
			"union_board_count":      unionBoardCount,
			"personal_board_count":   personalBoardCount,
			"raw_103_count":          raw103,
			"raw_92_count":           raw92,
			"cmd103_parse_triggered": cmd103,
			"cmd92_parse_triggered":  cmd92,
		},
	}}.Success(c)
}

func previewText(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func statusFor(ok bool) string {
	if ok {
		return "ready"
	}
	return "blocked"
}

func reasonFor(ok bool, yes string, no string) string {
	if ok {
		return yes
	}
	return no
}

func hostOnly(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	if i := strings.LastIndex(addr, ":"); i > 0 {
		return addr[:i]
	}
	return addr
}

func inferCmdAPIs(cmdID int) []string {
	examples := inferCmdAPIExamples(cmdID)
	apis := make([]string, 0, len(examples))
	for _, ex := range examples {
		apis = append(apis, ex.Method+" "+ex.Route)
	}
	return apis
}

func inferCmdAPIExamples(cmdID int) []apiExample {
	switch cmdID {
	case 92:
		return []apiExample{
			{Method: "GET", Route: "/v1/getReportNumByTaskId/:tid", Req: "path: tid=123", Resp: "{code:200,data:{count:88}}"},
			{Method: "GET", Route: "/v1/statisticsReport/:tid", Req: "path: tid=123", Resp: "{code:200,data:{arrive:66,absent:22}}"},
		}
	case 100:
		return []apiExample{{Method: "GET", Route: "/v1/getTeamGroup", Req: "无", Resp: "{code:200,data:[{group:\"百里\",member_count:50}]}"}}
	case 103:
		return []apiExample{
			{Method: "GET", Route: "/v1/getTeamUser", Req: "query: group=百里(可选)", Resp: "{code:200,data:[{id:101,name:\"xxx\",group:\"百里\",wu:1234}]}"},
			{Method: "GET", Route: "/v1/getGroupWu", Req: "无", Resp: "{code:200,data:[{group:\"百里\",total_wu:99999,member_count:50}]}"},
		}
	case 142, 143, 949:
		return []apiExample{{Method: "GET", Route: "/v1/getTeamUser", Req: "query: group=百里(新链路开发中)", Resp: "{code:200,data:[...]}(新链路开发中)"}}
	case 514:
		return []apiExample{{Method: "GET", Route: "/v1/stzb/leaderboard/personal", Req: "query: limit=5", Resp: "{code:200,data:[{rank:1,player_name:\"xxx\",metric_1:...}]}"}}
	case 700:
		return []apiExample{{Method: "GET", Route: "/v1/stzb/leaderboard/union", Req: "query: limit=5", Resp: "{code:200,data:[{rank:1,union_name:\"xxx\",power:...}]}"}}
	case 3758:
		return []apiExample{{Method: "GET", Route: "/v1/stzb/runtime/cmd/analysis", Req: "query: only_unknown=1&limit=200&cmd_id=700", Resp: "{code:200,data:{items:[...]}}"}}
	default:
		return []apiExample{}
	}
}

// RuntimeCmdIgnoreList 返回当前忽略列表（包含默认+人工配置）。
func RuntimeCmdIgnoreList(c *gin.Context) {
	common.Response{Data: gin.H{"items": capture.ListIgnoredCmdIDs()}}.Success(c)
}

// RuntimeCmdIgnoreSet 配置单个 cmd_id 是否忽略（影响抓包统计与日志打印）。
func RuntimeCmdIgnoreSet(c *gin.Context) {
	var req cmdIgnoreReq
	_ = c.ShouldBindJSON(&req)
	if req.CmdID <= 0 {
		if v := c.PostForm("cmd_id"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				req.CmdID = n
			}
		}
	}
	if req.CmdID <= 0 {
		common.Response{Message: "参数错误: cmd_id 必须为正整数"}.Error(c)
		return
	}
	capture.SetCmdIgnore(req.CmdID, req.Ignore)
	common.Response{Data: gin.H{"cmd_id": req.CmdID, "ignore": req.Ignore, "items": capture.ListIgnoredCmdIDs()}}.Success(c)
}

func inferCmdSemantics(row model.CmdSchema) (guess, suggestName, requestType string, confidence float64, basis []string) {
	switch row.CmdID {
	case 100:
		return "同盟总览信息（联盟基础状态）", "同盟总览信息", "state_pull", 0.95, []string{"样本含 union 级状态字段", "与同盟页打开强相关"}
	case 142:
		return "同盟分组元数据", "同盟分组元数据", "meta_pull", 0.98, []string{"行结构为 group_id/group_name/leader/member_count/power"}
	case 143:
		return "成员分组映射", "成员分组映射", "mapping_pull", 0.98, []string{"行结构为 member_id/group_id/group_name"}
	case 204:
		return "邮件/通知详情", "邮件详情", "detail_pull", 0.93, []string{"样本包含具体通知文本与 mail_id"}
	case 3758:
		return "同盟通告列表", "同盟通告列表", "list_pull", 0.98, []string{"行结构为 notice_id/author/title/time"}
	case 3759:
		return "同盟通告详情", "同盟通告详情", "detail_pull", 0.97, []string{"样本包含通告正文、目标、发布时间"}
	case 514:
		return "玩家外观批量快照", "玩家外观批量快照", "snapshot_pull", 0.9, []string{"形态为 [wid,[avatar,title,flag,extra]] 批量映射", "与947单条外观推送字段一致"}
	case 947:
		return "玩家外观增量", "玩家外观增量", "delta_push", 0.9, []string{"样本形态 [type,wid,value,flag,name,time,avatar,title,extra]"}
	case 700:
		return "同盟排行榜数据", "同盟排行榜", "board_pull", 0.99, []string{"含 union_id/name/power/member/city"}
	case 90005:
		return "通用数据变更推送", "通用数据变更推送", "delta_push", 0.96, []string{"样本由 Tb_* 表增量组成"}
	case 90006:
		return "心跳/短计数包", "心跳短包", "system_heartbeat", 0.95, []string{"高频且仅单数值"}
	case 90008:
		return "心跳ACK包", "心跳ACK", "system_heartbeat", 0.95, []string{"高频极短包"}
	case 694:
		return "服务器时间同步", "时间同步", "system_sync", 0.98, []string{"样本为13位时间戳"}
	case 2100:
		return "同盟聊天消息", "同盟聊天", "chat_stream", 0.98, []string{"样本含昵称、内容、身份字段"}
	case 2121:
		return "同盟计划/排表列表", "同盟计划列表", "plan_pull", 0.95, []string{"样本含 plan_id/title/time/wid/target_group"}
	case 2200:
		return "玩家名称/队伍标题广播", "名称状态广播", "state_broadcast", 0.8, []string{"样本形态 [类型,昵称,标题/状态]"}
	case 6314:
		return "团队对抗关系快照", "团队对抗快照", "state_snapshot", 0.78, []string{"样本为 [节点ID,值,关联列表] 持续快照"}
	case 6317:
		return "团队对抗变更事件", "团队对抗变更事件", "delta_event", 0.82, []string{"样本为节点ID列表，与6314联动"}
	case 949:
		return "同盟成员状态增量", "同盟成员状态增量", "delta_push", 0.75, []string{"样本含 uid/group/union/time/name 等增量字段"}
	}

	s := strings.ToLower(row.SampleText)
	matched := 0
	basis = make([]string, 0, 3)

	appendBasis := func(hit, label string) {
		if strings.Contains(s, hit) {
			matched++
			if len(basis) < 3 {
				basis = append(basis, label)
			}
		}
	}

	appendBasis("\"attack_name\"", "包含 attack_name")
	appendBasis("\"defend_name\"", "包含 defend_name")
	appendBasis("\"battle_id\"", "包含 battle_id")
	if matched >= 2 {
		return "战报/战斗相关", "战报相关", "detail_pull", 0.85, basis
	}

	matched = 0
	basis = basis[:0]
	appendBasis("\"union_id\"", "包含 union_id")
	appendBasis("\"total_member\"", "包含 total_member")
	appendBasis("\"total_npc_city\"", "包含 total_npc_city")
	if matched >= 2 {
		return "同盟排行榜相关", "同盟排行榜相关", "board_pull", 0.8, basis
	}

	matched = 0
	basis = basis[:0]
	appendBasis("\"contribute_week\"", "包含 contribute_week")
	appendBasis("\"join_time\"", "包含 join_time")
	appendBasis("\"wu\"", "包含 wu")
	if matched >= 2 {
		return "同盟成员/武勋相关", "同盟成员武勋相关", "state_pull", 0.75, basis
	}

	matched = 0
	basis = basis[:0]
	appendBasis("\"group_name\"", "包含 group_name")
	appendBasis("\"leader\"", "包含 leader")
	appendBasis("\"member_count\"", "包含 member_count")
	if matched >= 2 {
		return "同盟分组元数据相关", "同盟分组元数据", "meta_pull", 0.7, basis
	}

	if strings.Contains(s, "\"server\"") && strings.Contains(s, "\"role_name\"") {
		return "主公簿/角色激活相关", "主公簿激活", "state_pull", 0.7, []string{"包含 server", "包含 role_name"}
	}

	return "待人工确认", "待确认", "unknown", 0.3, []string{"样本信息不足或特征不明显"}
}

type cmdConfirmReq struct {
	CmdID   int             `json:"cmd_id"`
	CmdName string          `json:"cmd_name"`
	Items   []cmdConfirmReq `json:"items"`
}

// RuntimeCmdConfirm 将人工确认语义写入 cmd_schema.cmd_name。
func RuntimeCmdConfirm(c *gin.Context) {
	var req cmdConfirmReq
	_ = c.ShouldBindJSON(&req)

	if req.CmdID == 0 {
		if v := c.PostForm("cmd_id"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				req.CmdID = n
			}
		}
	}
	if strings.TrimSpace(req.CmdName) == "" {
		req.CmdName = strings.TrimSpace(c.PostForm("cmd_name"))
	}

	items := req.Items
	if len(items) == 0 && req.CmdID > 0 && strings.TrimSpace(req.CmdName) != "" {
		items = []cmdConfirmReq{{CmdID: req.CmdID, CmdName: req.CmdName}}
	}
	if len(items) == 0 {
		common.Response{Message: "参数错误: 需要 cmd_id/cmd_name 或 items[]"}.
			Error(c)
		return
	}
	if len(items) > 200 {
		common.Response{Message: "参数错误: 单次最多确认200条"}.
			Error(c)
		return
	}

	updated := 0
	failed := make([]gin.H, 0)
	for _, it := range items {
		name := strings.TrimSpace(it.CmdName)
		if it.CmdID <= 0 || name == "" || name == "未映射" {
			failed = append(failed, gin.H{"cmd_id": it.CmdID, "cmd_name": it.CmdName, "reason": "非法参数"})
			continue
		}
		if err := model.ConfirmCmdName(it.CmdID, name); err != nil {
			failed = append(failed, gin.H{"cmd_id": it.CmdID, "cmd_name": it.CmdName, "reason": err.Error()})
			continue
		}
		updated++
	}

	common.Response{Data: gin.H{
		"updated": updated,
		"failed":  failed,
	}}.Success(c)
}
