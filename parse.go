package parser

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"stzbHelper/global"
	"stzbHelper/internal/runtime"
	"stzbHelper/internal/service"
	"stzbHelper/model"
)

// ParseData 协议分发入口。
// 目前主要处理：
// - 100: 同盟总览兜底成员数据
// - 103: 同盟成员
// - 92 : 同盟战报/详细战报
func ParseData(cmdId int, data []byte) {
	if global.IsDebug {
		log.Println("收到[" + strconv.Itoa(cmdId) + "]消息:" + string(parseZlibData(data)))
	}

	if cmdId == 100 {
		parseTeamUserFromCmd100(data)
	} else if cmdId == 103 {
		atomic.AddUint64(&cmd103Count, 1)
		parseTeamUser(data)
	} else if cmdId == 92 {
		atomic.AddUint64(&cmd92Count, 1)
		parseReport(data)
		parseBattleData(data)
	} else if cmdId == 142 {
		parseUnionGroupMeta142(data)
	} else if cmdId == 949 {
		parseTeamUserFromCmd949(data)
	}
}

func DecodeType5(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	if data[0] == 5 {
		result := make([]byte, len(data)-1)
		for index, value := range data[1:] {
			result[index] = value ^ 152
		}
		return string(result)
	}
	return ""
}

var cmd103Count uint64
var cmd92Count uint64
var teamUserSaveCount uint64
var reportSaveCount uint64
var battleSaveCount uint64

// Stats 返回解析链路计数器，便于排障“抓到包但没有入库”的问题。
func Stats() (cmd103, cmd92, teamUsers, reports, battles uint64) {
	return atomic.LoadUint64(&cmd103Count), atomic.LoadUint64(&cmd92Count), atomic.LoadUint64(&teamUserSaveCount), atomic.LoadUint64(&reportSaveCount), atomic.LoadUint64(&battleSaveCount)
}

type RawData []interface{}

type BattleData struct {
	BattleId              int64       `json:"battle_id"`
	AttackHelpId          string      `json:"attack_help_id"`
	Time                  int64       `json:"time"`
	Wid                   interface{} `json:"wid"`
	WidName               string      `json:"wid_name"`
	AttackName            string      `json:"attack_name"`
	AttackUnionName       string      `json:"attack_union_name"`
	AttackClanName        string      `json:"attack_clan_name"`
	DefendClanName        string      `json:"defend_clan_name"`
	DefendName            string      `json:"defend_name"`
	DefendUnionName       string      `json:"defend_union_name"`
	AttackAdvance         string      `json:"attack_advance"`
	AttackAllHeroInfo     string      `json:"attack_all_hero_info"`
	AttackerGearInfo      string      `json:"attacker_gear_info"`
	DefendAdvance         string      `json:"defend_advance"`
	DefendAllHeroInfo     string      `json:"defend_all_hero_info"`
	DefenderGearInfo      string      `json:"defender_gear_info"`
	AttackHeroType        string      `json:"attack_hero_type"`
	AttackHeroTypeAdvance string      `json:"attack_hero_type_advance"`
	DefendHeroType        string      `json:"defend_hero_type"`
	DefendHeroTypeAdvance string      `json:"defend_hero_type_advance"`
	AttackHp              int64       `json:"attack_hp"`
	DefendHp              int64       `json:"defend_hp"`
	Npc                   int64       `json:"npc"`
	AllSkillInfo          string      `json:"all_skill_info"`
	Result                int64       `json:"result"`
	AttackIdu             string      `json:"attack_idu"`
	DefendIdu             string      `json:"defend_idu"`
}

// parseBattleData 解析详细战报并落库 battle_report。
func parseBattleData(data []byte) {
	msgdata := parseZlibData(data)
	if len(msgdata) == 0 {
		return
	}

	var rawData RawData
	if err := json.Unmarshal(msgdata, &rawData); err != nil {
		log.Printf("解析JSON失败: %v", err)
		return
	}

	for _, item := range rawData {
		battleArray, ok := item.([]interface{})
		if !ok || len(battleArray) == 0 {
			continue
		}
		battleMap, ok := battleArray[0].(map[string]interface{})
		if !ok {
			continue
		}

		var battleData BattleData
		jsonData, err := json.Marshal(battleMap)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(jsonData, &battleData); err != nil {
			continue
		}

		widStr := ""
		switch v := battleData.Wid.(type) {
		case string:
			widStr = v
		case float64:
			widStr = strconv.FormatInt(int64(v), 10)
		case int64:
			widStr = strconv.FormatInt(v, 10)
		case int:
			widStr = strconv.Itoa(v)
		default:
			widStr = fmt.Sprintf("%v", v)
		}

		report := model.BattleReport{
			BattleId:              battleData.BattleId,
			AttackHelpId:          battleData.AttackHelpId,
			Time:                  battleData.Time,
			Wid:                   widStr,
			WidName:               battleData.WidName,
			AttackName:            battleData.AttackName,
			AttackUnionName:       battleData.AttackUnionName,
			AttackClanName:        battleData.AttackClanName,
			DefendClanName:        battleData.DefendClanName,
			DefendName:            battleData.DefendName,
			DefendUnionName:       battleData.DefendUnionName,
			AttackAdvance:         battleData.AttackAdvance,
			AttackAllHeroInfo:     battleData.AttackAllHeroInfo,
			AttackerGearInfo:      battleData.AttackerGearInfo,
			DefendAdvance:         battleData.DefendAdvance,
			DefendAllHeroInfo:     battleData.DefendAllHeroInfo,
			DefenderGearInfo:      battleData.DefenderGearInfo,
			AttackHeroType:        battleData.AttackHeroType,
			AttackHeroTypeAdvance: battleData.AttackHeroTypeAdvance,
			DefendHeroType:        battleData.DefendHeroType,
			DefendHeroTypeAdvance: battleData.DefendHeroTypeAdvance,
			AttackHp:              battleData.AttackHp,
			DefendHp:              battleData.DefendHp,
			Npc:                   battleData.Npc,
			AllSkillInfo:          battleData.AllSkillInfo,
			Result:                battleData.Result,
			AttackIdu:             battleData.AttackIdu,
			DefendIdu:             battleData.DefendIdu,
		}

		report = parseHeroInfo(report)
		if result := model.Conn.Save(&report); result.Error == nil {
			atomic.AddUint64(&battleSaveCount, 1)
			runtime.TouchDataset(runtime.DatasetBattleReport)
		} else {
			log.Printf("[battle] 写库失败 battle_id=%d err=%v", report.BattleId, result.Error)
		}
	}
	log.Printf("[battle] parseBattleData 完成 saved=%d", atomic.LoadUint64(&battleSaveCount))
}

func parseHeroInfo(report model.BattleReport) model.BattleReport {
	attackAdvance := splitAndFilter(report.AttackAdvance, ";")
	attackTotal := int64(0)
	for i, advance := range attackAdvance {
		if i == 0 {
			continue
		}
		if len(advance) > 0 {
			star, err := strconv.ParseInt(advance[0], 10, 64)
			if err == nil {
				switch i {
				case 1:
					report.AttackHero1Star = star
				case 2:
					report.AttackHero2Star = star
				case 3:
					report.AttackHero3Star = star
				}
				attackTotal += star
			}
		}
	}
	report.AttackTotalStar = attackTotal

	defendAdvance := splitAndFilter(report.DefendAdvance, ";")
	defendTotal := int64(0)
	for i, advance := range defendAdvance {
		if i == 3 {
			continue
		}
		if len(advance) > 0 {
			star, err := strconv.ParseInt(advance[0], 10, 64)
			if err == nil {
				switch i {
				case 0:
					report.DefendHero3Star = star
				case 1:
					report.DefendHero2Star = star
				case 2:
					report.DefendHero1Star = star
				}
				defendTotal += star
			}
		}
	}
	report.DefendTotalStar = defendTotal

	for i, hero := range splitAndFilter(report.AttackAllHeroInfo, ";") {
		if len(hero) >= 2 {
			heroID, _ := strconv.ParseInt(hero[0], 10, 64)
			heroLevel, _ := strconv.ParseInt(hero[1], 10, 64)
			switch i {
			case 0:
				report.AttackHero1Id, report.AttackHero1Level = heroID, heroLevel
			case 1:
				report.AttackHero2Id, report.AttackHero2Level = heroID, heroLevel
			case 2:
				report.AttackHero3Id, report.AttackHero3Level = heroID, heroLevel
			}
		}
	}

	for i, hero := range splitAndFilter(report.DefendAllHeroInfo, ";") {
		if len(hero) >= 2 {
			heroID, _ := strconv.ParseInt(hero[0], 10, 64)
			heroLevel, _ := strconv.ParseInt(hero[1], 10, 64)
			switch i {
			case 0:
				report.DefendHero1Id, report.DefendHero1Level = heroID, heroLevel
			case 1:
				report.DefendHero2Id, report.DefendHero2Level = heroID, heroLevel
			case 2:
				report.DefendHero3Id, report.DefendHero3Level = heroID, heroLevel
			}
		}
	}

	// 武将名
	report.AttackHero1Name = service.HeroNameByID(report.AttackHero1Id)
	report.AttackHero2Name = service.HeroNameByID(report.AttackHero2Id)
	report.AttackHero3Name = service.HeroNameByID(report.AttackHero3Id)
	report.DefendHero1Name = service.HeroNameByID(report.DefendHero1Id)
	report.DefendHero2Name = service.HeroNameByID(report.DefendHero2Id)
	report.DefendHero3Name = service.HeroNameByID(report.DefendHero3Id)

	// 宝物名：格式 "gearId,level,enchant;..." 共4个slot，取第一个字段做ID
	report.AttackerGearNames = resolveGearNames(report.AttackerGearInfo)
	report.DefenderGearNames = resolveGearNames(report.DefenderGearInfo)

	// 技能名：格式 "slot,skillId1,lv,skillId2,lv,skillId3,lv;..."
	report.AllSkillNames = resolveSkillNames(report.AllSkillInfo)

	// 武将组合（排序后+号拼接，便于聚合）
	report.AttackHeroCombo = heroCombo(report.AttackHero1Id, report.AttackHero2Id, report.AttackHero3Id)
	report.DefendHeroCombo = heroCombo(report.DefendHero1Id, report.DefendHero2Id, report.DefendHero3Id)

	return report
}

// heroCombo 返回排序后的武将ID组合字符串，过滤零值
func heroCombo(ids ...int64) string {
	var nonZero []int64
	for _, id := range ids {
		if id != 0 {
			nonZero = append(nonZero, id)
		}
	}
	sort.Slice(nonZero, func(i, j int) bool { return nonZero[i] < nonZero[j] })
	parts := make([]string, len(nonZero))
	for i, id := range nonZero {
		parts[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, "+")
}

// resolveGearNames 把 "gearId,level,enchant;..." 转成 "名称;名称;..." （无宝物槽用-占位）
func resolveGearNames(raw string) string {
	slots := splitAndFilter(raw, ";")
	names := make([]string, 0, len(slots))
	for _, slot := range slots {
		if len(slot) == 0 {
			names = append(names, "-")
			continue
		}
		id, err := strconv.ParseInt(slot[0], 10, 64)
		if err != nil || id == 0 {
			names = append(names, "-")
			continue
		}
		name := service.GearNameByID(id)
		if name == "" {
			name = slot[0]
		}
		names = append(names, name)
	}
	return strings.Join(names, ";")
}

// resolveSkillNames 把 "slot,skillId1,lv,skillId2,lv,skillId3,lv;..." 转成 "名1/名2/名3;..." 每slot用/连接
func resolveSkillNames(raw string) string {
	slots := splitAndFilter(raw, ";")
	result := make([]string, 0, len(slots))
	for _, slot := range slots {
		// slot[0]=slotNum, slot[1]=skillId1, slot[2]=lv, slot[3]=skillId2, slot[4]=lv, slot[5]=skillId3, slot[6]=lv
		var names []string
		for _, idx := range []int{1, 3, 5} {
			if idx >= len(slot) {
				break
			}
			id, err := strconv.ParseInt(slot[idx], 10, 64)
			if err != nil || id == 0 {
				continue
			}
			name := service.SkillNameByID(id)
			if name == "" {
				name = slot[idx]
			}
			names = append(names, name)
		}
		if len(names) > 0 {
			result = append(result, strings.Join(names, "/"))
		}
	}
	return strings.Join(result, ";")
}

func splitAndFilter(input, delimiter string) [][]string {
	if input == "" {
		return [][]string{}
	}
	parts := strings.Split(input, delimiter)
	var result [][]string
	for _, part := range parts {
		if part == "" {
			continue
		}
		subParts := strings.Split(part, ",")
		var filtered []string
		for _, subPart := range subParts {
			if subPart != "" {
				filtered = append(filtered, subPart)
			}
		}
		if len(filtered) > 0 {
			result = append(result, filtered)
		}
	}
	return result
}

// parseReport 解析同盟战报（简版）并落库 reports。
// 这里不再依赖运行期开关和固定坐标，否则很容易出现“明明抓到 92，但没有任何战报入库”。
func parseReport(data []byte) {
	msgdata := parseZlibData(data)
	if len(msgdata) == 0 {
		return
	}

	var jsondata [][]any
	if err := json.Unmarshal(msgdata, &jsondata); err != nil {
		log.Printf("[report] JSON解析失败: %v", err)
		return
	}
	reports := make([]model.Report, 0, len(jsondata))

	for _, v := range jsondata {
		if len(v) == 0 {
			continue
		}
		reportJSON, err := json.Marshal(v[0])
		if err != nil {
			continue
		}
		var report model.Report
		if err = json.Unmarshal(reportJSON, &report); err != nil {
			continue
		}
		if report.BattleID == 0 {
			continue
		}
		reports = append(reports, report)
	}
	if len(reports) > 0 {
		if result := model.Conn.Save(&reports); result.Error == nil {
			atomic.AddUint64(&reportSaveCount, uint64(len(reports)))
			runtime.TouchDataset(runtime.DatasetReport)
		} else {
			log.Printf("[report] 写库失败: %v", result.Error)
		}
	}
}

func parseTeamUser(data []byte) {
	log.Println("收到同盟成员消息")
	if global.IsDebug {
		log.Println(string(parseZlibData(data)))
	}

	msgdata := parseZlibData(data)
	if len(msgdata) > 0 {
		var jsondata [][]any
		json.Unmarshal(msgdata, &jsondata)

		var ids []int
		var teamUsers []model.TeamUser
		for _, item := range jsondata {
			tu := model.ToTeamUser(item)
			if tu.Id == 0 {
				continue
			}
			teamUsers = append(teamUsers, tu)
			ids = append(ids, tu.Id)
		}

		log.Println("同盟成员消息解析成功！共" + strconv.Itoa(len(teamUsers)) + "人")
		model.Conn.Save(teamUsers)
		model.Conn.Not("id", ids).Delete(model.TeamUser{})
	} else {
		log.Println("解析同盟成员消息失败")
	}
}

func extractTeamUserRows(root any) [][]any {
	// 形态1：直接是 [][]any
	if arr, ok := root.([]any); ok {
		rows := make([][]any, 0)
		for _, v := range arr {
			if row, ok := v.([]any); ok {
				rows = append(rows, row)
			}
		}
		if len(rows) > 0 {
			return rows
		}
	}
	// 形态2：对象包裹，常见键 data/list/members/items
	if obj, ok := root.(map[string]any); ok {
		for _, key := range []string{"data", "list", "members", "items"} {
			if v, exists := obj[key]; exists {
				if rows := extractTeamUserRows(v); len(rows) > 0 {
					return rows
				}
			}
		}
	}
	return nil
}

// parseTeamUserFromCmd100 兜底解析：递归扫描 cmd100 JSON，提取形如同盟成员的二维数组行。
func parseTeamUserFromCmd100(data []byte) {
	msgdata := parseZlibData(data)
	if len(msgdata) == 0 {
		return
	}
	var root any
	if err := json.Unmarshal(msgdata, &root); err != nil {
		log.Printf("[teamuser-cmd100] JSON解析失败: %v", err)
		return
	}

	rows := make([][]any, 0)
	collectCandidateTeamUserRows(root, &rows)
	if len(rows) == 0 {
		log.Printf("[teamuser-cmd100] 未找到候选成员数组")
		return
	}

	ids := make([]int, 0, len(rows))
	teamUsers := make([]model.TeamUser, 0, len(rows))
	for _, row := range rows {
		tu := model.ToTeamUser(row)
		if tu.Id == 0 || tu.Name == "" {
			continue
		}
		teamUsers = append(teamUsers, tu)
		ids = append(ids, tu.Id)
	}
	if len(teamUsers) == 0 {
		log.Printf("[teamuser-cmd100] 候选数组存在，但无有效成员")
		return
	}
	if result := model.Conn.Save(teamUsers); result.Error != nil {
		log.Printf("[teamuser-cmd100] 写库失败: %v", result.Error)
		return
	}
	atomic.AddUint64(&teamUserSaveCount, uint64(len(teamUsers)))
	runtime.TouchDataset(runtime.DatasetTeamUser)
	log.Printf("[teamuser-cmd100] 保存成功 count=%d", len(teamUsers))
}

func collectCandidateTeamUserRows(node any, rows *[][]any) {
	switch v := node.(type) {
	case []any:
		// 先判断自己是否是“成员行列表”
		if len(v) > 0 {
			rowCount := 0
			for _, item := range v {
				row, ok := item.([]any)
				if !ok {
					continue
				}
				if looksLikeTeamUserRow(row) {
					*rows = append(*rows, row)
					rowCount++
				}
			}
			if rowCount > 0 {
				return
			}
		}
		for _, item := range v {
			collectCandidateTeamUserRows(item, rows)
		}
	case map[string]any:
		for _, vv := range v {
			collectCandidateTeamUserRows(vv, rows)
		}
	}
}

func looksLikeTeamUserRow(row []any) bool {
	if len(row) < 31 {
		return false
	}
	_, idOK := row[0].(float64)
	name, nameOK := row[1].(string)
	_, powerOK := row[8].(float64)
	return idOK && nameOK && name != "" && powerOK
}

func parseUnionGroupMeta142(data []byte) {
	msgdata := parseZlibData(data)
	if len(msgdata) == 0 {
		return
	}
	var rows [][]any
	if err := json.Unmarshal(msgdata, &rows); err != nil {
		log.Printf("[group142] JSON解析失败: %v", err)
		return
	}
	if len(rows) == 0 {
		return
	}
	for _, r := range rows {
		if len(r) < 6 {
			continue
		}
		rec := model.UnionGroupMeta{
			GroupID:     toIntSafe(r[0]),
			GroupName:   toStringSafe(r[1]),
			GroupCode:   toIntSafe(r[2]),
			LeaderName:  toStringSafe(r[3]),
			MemberCount: toIntSafe(r[4]),
			Power:       int64(toIntSafe(r[5])),
			CaptureTime: time.Now().Unix(),
		}
		model.UpsertUnionGroupMeta(rec)
		runtime.TouchDataset(runtime.DatasetUnionGroupMeta)
	}
}

func parseTeamUserFromCmd949(data []byte) {
	msgdata := parseZlibData(data)
	if len(msgdata) == 0 {
		return
	}
	var root []any
	if err := json.Unmarshal(msgdata, &root); err != nil {
		return
	}
	if len(root) < 3 {
		return
	}
	list, ok := root[2].([]any)
	if !ok || len(list) == 0 {
		return
	}
	codeMap := model.LoadGroupCodeNameMap()
	saved := 0
	for _, item := range list {
		row, ok := item.([]any)
		if !ok || len(row) < 15 {
			continue
		}
		uid := toIntSafe(row[0])
		if uid == 0 {
			continue
		}
		gcode := toIntSafe(row[1])
		name := toStringSafe(row[13])
		if name == "" {
			continue
		}
		group := codeMap[gcode]
		if group == "" {
			group = strconv.Itoa(gcode)
		}
		wu := toIntSafe(row[6])
		joinTime := toIntSafe(row[12])
		model.UpsertTeamUserFromCmd949(uid, name, group, wu, joinTime)
		saved++
	}
	if saved > 0 {
		atomic.AddUint64(&teamUserSaveCount, uint64(saved))
		log.Printf("[teamuser-cmd949] 保存成功 count=%d", saved)
	}
}

func toIntSafe(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	default:
		return 0
	}
}

func toStringSafe(v any) string {
	s, _ := v.(string)
	return s
}

func parseZlibData(data []byte) []byte {
	if len(data) >= 2 && data[0] == 120 && data[1] == 156 {
		compressedReader := bytes.NewReader(data)
		zlibReader, err := zlib.NewReader(compressedReader)
		if err != nil {
			return []byte{}
		}
		defer zlibReader.Close()

		uncompressedData, err := io.ReadAll(zlibReader)
		if err != nil {
			return []byte{}
		}
		return uncompressedData
	}
	return data
}
