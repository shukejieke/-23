package capture

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"stzbHelper/global"
	"stzbHelper/internal/parser"
	"stzbHelper/internal/repo"
	"stzbHelper/internal/runtime"
	"stzbHelper/model"
)

var databaseSelected bool

var packetSeen uint64
var parseTriggered uint64

var cmdMu sync.RWMutex
var cmdSeenCounter = map[int]uint64{}

var flowMu sync.Mutex
var flowStates = map[string]*flowState{}

var rankTraceMu sync.RWMutex
var recentWireEvents []WireEvent
var rankTraceRecords []RankTraceRecord

const (
	recentWireEventsMax = 600
	rankTraceMax        = 120
	rankTraceWindowSec  = int64(2)
	flowIdleTTL         = 20 * time.Second
	packetLossTTL       = 5 * time.Second
)

type CmdSeenStat struct {
	CmdID int    `json:"cmd_id"`
	Count uint64 `json:"count"`
}

type flowState struct {
	fullbuf      []byte
	waitbuf      bool
	packetLoss   bool
	lossBytes    []byte
	lossCmdID    int
	needBufSize  int
	lossUpdated  time.Time
	lastActivity time.Time
}

type WireEvent struct {
	TsUnix       int64  `json:"ts_unix"`
	TsText       string `json:"ts_text"`
	Kind         string `json:"kind"` // valid / invalid
	CmdID        int    `json:"cmd_id"`
	CmdName      string `json:"cmd_name"`
	DataType     int    `json:"data_type"`
	PayloadLen   int    `json:"payload_len"`
	PacketBufLen int    `json:"packet_buf_len"`
	Offset       int    `json:"offset"`
	Src          string `json:"src"`
	Dst          string `json:"dst"`
	PSH          bool   `json:"psh"`
	Preview      string `json:"preview"`
}

type RankTraceRecord struct {
	TsUnix       int64       `json:"ts_unix"`
	TsText       string      `json:"ts_text"`
	TriggerCmdID int         `json:"trigger_cmd_id"`
	TriggerName  string      `json:"trigger_name"`
	TriggerKind  string      `json:"trigger_kind"`
	PayloadLen   int         `json:"payload_len"`
	PacketBufLen int         `json:"packet_buf_len"`
	Src          string      `json:"src"`
	Dst          string      `json:"dst"`
	Preview      string      `json:"preview"`
	Around       []WireEvent `json:"around"`
}

func cmdNameCN(cmdID int) string {
	switch cmdID {
	case 3686:
		return "主公簿激活包"
	case 103:
		return "同盟成员数据"
	case 100:
		return "同盟总览信息"
	case 143:
		return "成员分组映射"
	case 142:
		return "同盟分组元数据"
	case 949:
		return "成员状态增量"
	case 92:
		return "战报数据"
	case 514:
		return "个人积分事件快照"
	case 947:
		return "玩家外观增量"
	case 6314:
		return "个人领地排行榜"
	case 6317:
		return "领地排名位置引用"
	case 700:
		return "同盟排行榜"
	case 2100:
		return "同盟发言"
	case 3758:
		return "同盟通告"
	default:
		return "未映射"
	}
}

var ignoreMu sync.RWMutex
var ignoredCmdIDs = map[int]bool{
	2100:  true, // 同盟发言（用户已确认过滤）
	90008: true, // 高频心跳/确认包
	694:   true, // 时间戳同步包
	2200:  true, // 名称/状态广播包
	90006: true, // 高频短包（当前无业务价值）
}

func shouldFilterCmd(cmdID int) bool {
	ignoreMu.RLock()
	defer ignoreMu.RUnlock()
	return ignoredCmdIDs[cmdID]
}

func ListIgnoredCmdIDs() []int {
	ignoreMu.RLock()
	defer ignoreMu.RUnlock()
	ids := make([]int, 0, len(ignoredCmdIDs))
	for id, on := range ignoredCmdIDs {
		if on {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	return ids
}

func SetCmdIgnore(cmdID int, ignore bool) {
	ignoreMu.Lock()
	defer ignoreMu.Unlock()
	if ignore {
		ignoredCmdIDs[cmdID] = true
		return
	}
	delete(ignoredCmdIDs, cmdID)
}

func IsCmdIgnored(cmdID int) bool {
	ignoreMu.RLock()
	defer ignoreMu.RUnlock()
	return ignoredCmdIDs[cmdID]
}

func Stats() (seen uint64, parsed uint64, selected bool) {
	return atomic.LoadUint64(&packetSeen), atomic.LoadUint64(&parseTriggered), databaseSelected
}

func frameMeta(buf []byte, offset int) (bufsize, cmdID, dataType, totalLen int, ok bool) {
	if offset < 0 || len(buf)-offset < 8 {
		return 0, 0, -1, 0, false
	}
	remain := buf[offset:]
	bufsize = int(binary.BigEndian.Uint32(buf[offset : offset+4]))
	cmdID = int(binary.BigEndian.Uint32(buf[offset+4 : offset+8]))
	dataType = -1
	if len(remain) > 12 {
		dataType = int(buf[offset+12])
	}
	totalLen = bufsize + 4
	if bufsize <= 0 || totalLen <= 0 || totalLen > len(remain) {
		return bufsize, cmdID, dataType, totalLen, false
	}
	if dataType != 3 && dataType != 5 {
		return bufsize, cmdID, dataType, totalLen, false
	}
	return bufsize, cmdID, dataType, totalLen, true
}

func pickFrame(buf []byte) (offset int, frame []byte, remain []byte, bufsize, cmdID, dataType int, ok bool) {
	if len(buf) < 13 {
		return 0, nil, nil, 0, 0, -1, false
	}
	if bs, cmd, dt, totalLen, valid := frameMeta(buf, 0); valid {
		return 0, buf[:totalLen], buf[totalLen:], bs, cmd, dt, true
	}
	maxOffset := 128
	if len(buf)-13 < maxOffset {
		maxOffset = len(buf) - 13
	}
	for off := 1; off <= maxOffset; off++ {
		bs, cmd, dt, totalLen, valid := frameMeta(buf, off)
		if !valid {
			continue
		}
		slice := buf[off:]
		return off, slice[:totalLen], slice[totalLen:], bs, cmd, dt, true
	}
	return 0, nil, nil, 0, 0, -1, false
}

func decodeResponse(dataType int, frame []byte) (string, bool) {
	if len(frame) <= 17 {
		return "", false
	}
	switch dataType {
	case 3:
		raw := parseZlibData(frame[17:])
		if len(raw) == 0 {
			return "", false
		}
		return string(raw), true
	case 5:
		decoded := parser.DecodeType5(frame[12:])
		if decoded == "" {
			return "", false
		}
		return decoded, true
	default:
		return "", false
	}
}

func responsePreview(dataType int, frame []byte) (string, bool) {
	decoded, ok := decodeResponse(dataType, frame)
	if !ok {
		return "", false
	}
	preview, _, _ := safePreview(decoded, 200)
	return preview, true
}

func responseJSONPreview(decoded string, max int) (preview string, total int, truncated bool, ok bool) {
	var obj any
	if err := json.Unmarshal([]byte(decoded), &obj); err != nil {
		return "", 0, false, false
	}
	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", 0, false, false
	}
	p, total, truncated := safePreview(string(pretty), max)
	return p, total, truncated, true
}

func responseKVPreview(cmdID int, decoded string, maxItems int) (string, bool) {
	switch cmdID {
	case 514:
		var arr []any
		if err := json.Unmarshal([]byte(decoded), &arr); err != nil {
			return "", false
		}
		if len(arr) == 0 {
			return "", false
		}
		parts := make([]string, 0, maxItems)
		for i := 0; i+1 < len(arr) && len(parts) < maxItems; i += 2 {
			k := fmt.Sprintf("%v", arr[i])
			vBytes, _ := json.Marshal(arr[i+1])
			parts = append(parts, fmt.Sprintf("事件%s=%s", k, string(vBytes)))
		}
		if len(parts) == 0 {
			return "", false
		}
		result := strings.Join(parts, " | ")
		if len(arr)/2 > maxItems {
			result += " | ..."
		}
		preview, _, _ := safePreview(result, 1200)
		return preview, true
	case 700:
		var top []any
		if err := json.Unmarshal([]byte(decoded), &top); err != nil {
			return "", false
		}
		if len(top) < 5 {
			return "", false
		}
		rows, ok := top[4].([]any)
		if !ok || len(rows) == 0 {
			return "", false
		}
		parts := make([]string, 0, maxItems)
		for i := 0; i < len(rows) && len(parts) < maxItems; i++ {
			pair, ok := rows[i].([]any)
			if !ok || len(pair) < 2 {
				continue
			}
			rank := fmt.Sprintf("%v", pair[0])
			obj, ok := pair[1].(map[string]any)
			if !ok {
				continue
			}
			name := fmt.Sprintf("%v", obj["name"])
			power := fmt.Sprintf("%v", obj["power"])
			member := fmt.Sprintf("%v", obj["total_member"])
			city := fmt.Sprintf("%v", obj["total_npc_city"])
			uid := fmt.Sprintf("%v", obj["union_id"])
			parts = append(parts, fmt.Sprintf("排名%s 联盟=%s 势力=%s 成员=%s 城池=%s union_id=%s", rank, name, power, member, city, uid))
		}
		if len(parts) == 0 {
			return "", false
		}
		result := strings.Join(parts, " | ")
		if len(rows) > maxItems {
			result += " | ..."
		}
		preview, _, _ := safePreview(result, 1200)
		return preview, true
	case 3758:
		var rows []any
		if err := json.Unmarshal([]byte(decoded), &rows); err != nil {
			return "", false
		}
		if len(rows) == 0 {
			return "", false
		}
		parts := make([]string, 0, maxItems)
		for i := 0; i < len(rows) && len(parts) < maxItems; i++ {
			item, ok := rows[i].([]any)
			if !ok || len(item) < 4 {
				continue
			}
			id := fmt.Sprintf("%v", item[0])
			author := fmt.Sprintf("%v", item[1])
			title := fmt.Sprintf("%v", item[2])
			ts := fmt.Sprintf("%v", item[3])
			parts = append(parts, fmt.Sprintf("通告#%s 发布者=%s 内容=%s 时间=%s", id, author, title, ts))
		}
		if len(parts) == 0 {
			return "", false
		}
		result := strings.Join(parts, " | ")
		if len(rows) > maxItems {
			result += " | ..."
		}
		preview, _, _ := safePreview(result, 1200)
		return preview, true
	default:
		return "", false
	}
}

func safePreview(s string, max int) (preview string, total int, truncated bool) {
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	total = len(s)
	if len(s) > max {
		return s[:max] + "...", total, true
	}
	return s, total, false
}

func recordCmd(cmdID int) {
	cmdMu.Lock()
	cmdSeenCounter[cmdID]++
	cmdMu.Unlock()
}

func recordWireEvent(ev WireEvent) {
	rankTraceMu.Lock()
	recentWireEvents = append(recentWireEvents, ev)
	if len(recentWireEvents) > recentWireEventsMax {
		recentWireEvents = recentWireEvents[len(recentWireEvents)-recentWireEventsMax:]
	}
	rankTraceMu.Unlock()
}

func isRankTriggerCmd(cmdID int) bool {
	return cmdID == 700 || cmdID == 514
}

func appendRankTrace(trigger WireEvent) {
	triggerTs := trigger.TsUnix
	rankTraceMu.Lock()
	around := make([]WireEvent, 0, 64)
	for _, ev := range recentWireEvents {
		delta := ev.TsUnix - triggerTs
		if delta < 0 {
			delta = -delta
		}
		if delta <= rankTraceWindowSec {
			around = append(around, ev)
		}
	}
	rec := RankTraceRecord{
		TsUnix:       trigger.TsUnix,
		TsText:       trigger.TsText,
		TriggerCmdID: trigger.CmdID,
		TriggerName:  trigger.CmdName,
		TriggerKind:  trigger.Kind,
		PayloadLen:   trigger.PayloadLen,
		PacketBufLen: trigger.PacketBufLen,
		Src:          trigger.Src,
		Dst:          trigger.Dst,
		Preview:      trigger.Preview,
		Around:       around,
	}
	rankTraceRecords = append(rankTraceRecords, rec)
	if len(rankTraceRecords) > rankTraceMax {
		rankTraceRecords = rankTraceRecords[len(rankTraceRecords)-rankTraceMax:]
	}
	rankTraceMu.Unlock()
	writeRankTraceLine(rec)
}

func writeRankTraceLine(rec RankTraceRecord) {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		return
	}
	f, err := os.OpenFile("logs/ranktrace.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	b, err := json.Marshal(rec)
	if err != nil {
		return
	}
	_, _ = f.Write(append(b, '\n'))
}

func invalidPreview(buf []byte, maxBytes int) string {
	if len(buf) == 0 {
		return ""
	}
	if len(buf) > maxBytes {
		return hex.EncodeToString(buf[:maxBytes]) + "..."
	}
	return hex.EncodeToString(buf)
}

func RankTraceLatest(limit int) []RankTraceRecord {
	if limit <= 0 {
		limit = 20
	}
	if limit > rankTraceMax {
		limit = rankTraceMax
	}
	rankTraceMu.RLock()
	defer rankTraceMu.RUnlock()
	if len(rankTraceRecords) == 0 {
		return []RankTraceRecord{}
	}
	start := len(rankTraceRecords) - limit
	if start < 0 {
		start = 0
	}
	out := make([]RankTraceRecord, 0, len(rankTraceRecords)-start)
	for i := len(rankTraceRecords) - 1; i >= start; i-- {
		out = append(out, rankTraceRecords[i])
	}
	return out
}

func shouldArchiveRawCapture(cmdID int) bool {
	return cmdID == 100 || cmdID == 103 || cmdID == 92
}

// CmdStats 返回抓包阶段观察到的所有协议号计数（按 count desc, cmd_id asc 排序）。
func CmdStats() []CmdSeenStat {
	cmdMu.RLock()
	stats := make([]CmdSeenStat, 0, len(cmdSeenCounter))
	for cmdID, count := range cmdSeenCounter {
		stats = append(stats, CmdSeenStat{CmdID: cmdID, Count: count})
	}
	cmdMu.RUnlock()

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count == stats[j].Count {
			return stats[i].CmdID < stats[j].CmdID
		}
		return stats[i].Count > stats[j].Count
	})
	return stats
}

func flowKey(srcIP, dstIP string) string {
	return srcIP + "->" + dstIP
}

func getFlowState(srcIP, dstIP string) *flowState {
	now := time.Now()
	key := flowKey(srcIP, dstIP)

	flowMu.Lock()
	defer flowMu.Unlock()

	for k, st := range flowStates {
		if now.Sub(st.lastActivity) > flowIdleTTL {
			delete(flowStates, k)
		}
	}

	st, ok := flowStates[key]
	if !ok {
		st = &flowState{}
		flowStates[key] = st
	}
	st.lastActivity = now
	if st.packetLoss && now.Sub(st.lossUpdated) > packetLossTTL {
		st.packetLoss = false
		st.lossBytes = nil
		st.lossCmdID = 0
		st.needBufSize = 0
	}
	return st
}

func safeGo(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[panic-recovered] task=%s panic=%v", name, r)
			}
		}()
		fn()
	}()
}

// Start 在指定网卡上启动抓包循环。
// 过滤条件：仅处理 src port=8001 的 TCP 包。
func Start(deviceName string, wg *sync.WaitGroup) {
	defer wg.Done()
	ensureDBReady()

	handle, err := pcap.OpenLive(deviceName, 65535, true, pcap.BlockForever)
	if err != nil {
		log.Printf("无法打开接口 %s: %v\n", deviceName, err)
		return
	}
	defer handle.Close()

	if err = handle.SetBPFFilter("tcp and src port 8001"); err != nil {
		log.Printf("无法在接口 %s 上设置过滤器: %v\n", deviceName, err)
		return
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		handlePacket(packet)
	}
}

// handlePacket 负责包重组、协议头读取、丢包拼接、分发给 parser。
func handlePacket(packet gopacket.Packet) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return
	}
	appLayer := packet.ApplicationLayer()
	if appLayer == nil {
		return
	}

	payload := appLayer.Payload()
	if len(payload) < 8 {
		return
	}
	seen := atomic.AddUint64(&packetSeen, 1)
	if seen%5000 == 0 {
		log.Printf("[capture] 已捕获数据包: %d, 已触发解析: %d", seen, atomic.LoadUint64(&parseTriggered))
	}

	PSH := tcpLayer.(*layers.TCP).PSH
	srcIP, dstIP := packetIP(packet, tcpLayer)
	if srcIP == "" || dstIP == "" {
		return
	}
	log.Printf("[packet-hit] payload_len=%d src=%s dst=%s psh=%v", len(payload), srcIP, dstIP, PSH)

	if global.ExVar.BindIpInfo {
		boundSrc, boundDst := runtime.GetBoundIPs()
		if boundSrc != "" && boundDst != "" {
			if boundSrc != srcIP || boundDst != dstIP {
				return
			}
		}
	}

	state := getFlowState(srcIP, dstIP)

	var buf []byte
	if !PSH {
		state.fullbuf = append(state.fullbuf, payload...)
		return
	}
	if len(state.fullbuf) > 0 {
		buf = append(state.fullbuf, payload...)
		state.fullbuf = nil
	} else {
		buf = payload
	}
	state.waitbuf = false

	work := buf
	for len(work) >= 13 {
		offset, frame, remain, bufsize, cmdID, dataType, ok := pickFrame(work)
		if !ok {
			log.Printf("[signal-noise] reason=invalid-frame payload_len=%d packet_buf_len=%d src=%s dst=%s psh=%v", len(payload), len(work), srcIP, dstIP, PSH)
			now := time.Now()
			recordWireEvent(WireEvent{
				TsUnix:       now.Unix(),
				TsText:       now.Format("2006-01-02 15:04:05"),
				Kind:         "invalid",
				CmdID:        0,
				CmdName:      "invalid-frame",
				DataType:     -1,
				PayloadLen:   len(payload),
				PacketBufLen: len(work),
				Offset:       0,
				Src:          srcIP,
				Dst:          dstIP,
				PSH:          PSH,
				Preview:      invalidPreview(work, 48),
			})
			break
		}
		if shouldFilterCmd(cmdID) {
			work = remain
			continue
		}

		if offset > 0 {
			log.Printf("[signal-resync] offset=%d cmd_id=%d cmd_name=%s data_type=%d payload_len=%d packet_buf_len=%d src=%s dst=%s psh=%v", offset, cmdID, cmdNameCN(cmdID), dataType, len(payload), len(work), srcIP, dstIP, PSH)
		} else {
			log.Printf("[signal-valid] cmd_id=%d cmd_name=%s data_type=%d payload_len=%d packet_buf_len=%d src=%s dst=%s psh=%v", cmdID, cmdNameCN(cmdID), dataType, len(payload), len(work), srcIP, dstIP, PSH)
		}
		recordCmd(cmdID)
		now := time.Now()
		ev := WireEvent{
			TsUnix:       now.Unix(),
			TsText:       now.Format("2006-01-02 15:04:05"),
			Kind:         "valid",
			CmdID:        cmdID,
			CmdName:      cmdNameCN(cmdID),
			DataType:     dataType,
			PayloadLen:   len(payload),
			PacketBufLen: len(work),
			Offset:       offset,
			Src:          srcIP,
			Dst:          dstIP,
			PSH:          PSH,
		}

		if preview, ok := responsePreview(dataType, frame); ok {
			log.Printf("[response-preview] cmd_id=%d data_type=%d preview=%s", cmdID, dataType, preview)
			ev.Preview = preview
		}
		decoded, decodedOK := decodeResponse(dataType, frame)
		if shouldArchiveRawCapture(cmdID) {
			rawHex := hex.EncodeToString(frame)
			safeGo("SaveRawCapture", func() {
				model.SaveRawCapture(model.RawCapture{
					CmdID:        cmdID,
					CmdName:      cmdNameCN(cmdID),
					DataType:     dataType,
					Src:          srcIP,
					Dst:          dstIP,
					PayloadLen:   len(payload),
					PacketBufLen: len(work),
					Preview:      ev.Preview,
					RawHex:       rawHex,
					DecodedText:  decoded,
				})
			})
		}
		if decodedOK {
			safeGo("SaveCmdSchema", func() {
				model.SaveCmdSchema(cmdID, cmdNameCN(cmdID), dataType, decoded, hex.EncodeToString(frame))
			})
			if cmdID == 700 {
				runtime.TouchDataset(runtime.DatasetUnionLeaderboard)
				safeGo("SaveUnionLeaderboard", func() {
					model.SaveUnionLeaderboardFromDecoded(decoded, cmdID)
				})
			}
			if cmdID == 514 {
				runtime.TouchDataset(runtime.DatasetPersonalLeaderboard)
				safeGo("SavePersonalLeaderboard", func() {
					model.SavePersonalLeaderboardFromDecoded(decoded, cmdID)
				})
			}
			if cmdID == 6314 {
				runtime.TouchDataset(runtime.DatasetPersonalLeaderboard)
				safeGo("SavePlayerTerritoryRank", func() {
					model.SavePlayerTerritoryRankFromDecoded(decoded, cmdID)
				})
			}
			jsonMax := 1200
			if cmdID == 514 || cmdID == 6314 || cmdID == 6317 || cmdID == 700 {
				jsonMax = 5000
			}
			if jsonPreview, total, truncated, ok := responseJSONPreview(decoded, jsonMax); ok {
				log.Printf("[response-json] cmd_id=%d data_type=%d json_chars=%d truncated=%v json=%s", cmdID, dataType, total, truncated, jsonPreview)
			}
			if kvPreview, ok := responseKVPreview(cmdID, decoded, 8); ok {
				log.Printf("[response-kv] cmd_id=%d data_type=%d kv=%s", cmdID, dataType, kvPreview)
			}
		}
		recordWireEvent(ev)
		if isRankTriggerCmd(cmdID) {
			appendRankTrace(ev)
		}

		switch dataType {
		case 3:
			if len(frame)-bufsize != 4 && (cmdID == 103 || cmdID == 92) {
				state.lossCmdID = cmdID
				state.lossBytes = append([]byte(nil), frame...)
				state.packetLoss = true
				state.needBufSize = bufsize
				state.lossUpdated = time.Now()
			} else {
				if cmdID == 100 || cmdID == 103 || cmdID == 92 || cmdID == 142 || cmdID == 949 {
					atomic.AddUint64(&parseTriggered, 1)
					safeGo("ParseData/type3", func() {
						parser.ParseData(cmdID, frame[17:])
					})
				}
			}
		case 5:
			decodedType5 := parser.DecodeType5(frame[12:])
			// 兼容协议变更：部分业务包（如同盟成员/战报）可能从 data_type=3 切到 data_type=5。
			// 这些包必须继续走 ParseData 才能落库，否则前端会出现"有抓包、无数据"。
			if (cmdID == 100 || cmdID == 103 || cmdID == 92 || cmdID == 142 || cmdID == 949) && decodedType5 != "" {
				atomic.AddUint64(&parseTriggered, 1)
				safeGo("ParseData/type5", func() {
					parser.ParseData(cmdID, []byte(decodedType5))
				})
			} else if cmdID == 100 || cmdID == 103 || cmdID == 92 || cmdID == 142 || cmdID == 949 {
				log.Printf("[capture] unhandled data_type=5 empty decoded cmd_id=%d packet_buf_len=%d src=%s dst=%s", cmdID, len(work), srcIP, dstIP)
			}
		default:
			log.Printf("[capture] unhandled data_type=%d cmd_id=%d packet_buf_len=%d src=%s dst=%s", dataType, cmdID, len(work), srcIP, dstIP)
		}

		if cmdID > 99999 && state.packetLoss && (state.lossCmdID == 103 || state.lossCmdID == 92) {
			result := make([]byte, len(frame)+len(state.lossBytes))
			copy(result, state.lossBytes)
			copy(result[len(state.lossBytes):], frame)
			if len(frame)+len(state.lossBytes)-state.needBufSize != 4 {
				state.lossBytes = result
				state.lossUpdated = time.Now()
			} else {
				state.packetLoss = false
				atomic.AddUint64(&parseTriggered, 1)
				lossCmdID := state.lossCmdID
				state.lossBytes = nil
				state.lossCmdID = 0
				state.needBufSize = 0
				safeGo("ParseData/loss-recovery", func() {
					parser.ParseData(lossCmdID, result[17:])
				})
			}
		}

		if cmdID == 3686 && !databaseSelected {
			initDatabaseFromPacket(frame, srcIP, dstIP)
		}
		work = remain
	}

	// 大帧跨 PSH 边界时，末尾剩余字节是下一帧的头部开始——保存到 fullbuf 供下包拼接。
	// 不保存的话，invalid-frame 后流永远失去同步。
	if len(work) >= 8 {
		bs := int(binary.BigEndian.Uint32(work[0:4]))
		totalLen := bs + 4
		if bs > 0 && totalLen > len(work) && totalLen <= 1<<20 {
			state.fullbuf = append([]byte(nil), work...)
		}
	} else if len(work) > 0 {
		state.fullbuf = append([]byte(nil), work...)
	}
}

func ensureDBReady() {
	if model.Conn != nil {
		databaseSelected = true
		return
	}
	// MySQL-only 模式下，数据库名参数已忽略；这里直接确保连接可用，避免等待 3686 导致页面一直空数据。
	model.InitDB("")
	repo.SetDB(model.Conn)
	databaseSelected = true
	log.Printf("[capture] 已初始化默认数据库连接（mysql-only）")
}

func packetIP(packet gopacket.Packet, tcpLayer gopacket.Layer) (srcIP, dstIP string) {
	ipLayer := packet.NetworkLayer()
	if ipLayer == nil {
		return "", ""
	}
	srcPort := int(tcpLayer.(*layers.TCP).SrcPort)
	dstPort := int(tcpLayer.(*layers.TCP).DstPort)

	switch ip := ipLayer.(type) {
	case *layers.IPv4:
		return ip.SrcIP.String() + ":" + strconv.Itoa(srcPort), ip.DstIP.String() + ":" + strconv.Itoa(dstPort)
	case *layers.IPv6:
		return ip.SrcIP.String() + ":" + strconv.Itoa(srcPort), ip.DstIP.String() + ":" + strconv.Itoa(dstPort)
	default:
		return "", ""
	}
}

// initDatabaseFromPacket 从主公簿包提取角色/区服信息并初始化 DB。
func initDatabaseFromPacket(buf []byte, srcIP, dstIP string) {
	var data []byte
	if buf[12] == 5 {
		data = []byte(parser.DecodeType5(buf[12:]))
	} else if buf[12] == 3 {
		data = parseZlibData(buf[17:])
	}

	var raw []interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Printf("解析主公簿失败: %v", err)
		return
	}
	if len(raw) < 2 {
		return
	}
	dataMap, ok := raw[1].(map[string]interface{})
	if !ok {
		return
	}
	server, ok := dataMap["server"].([]interface{})
	if !ok || len(server) == 0 {
		log.Println("[capture] 主公簿缺少 server 字段，跳过绑定")
		return
	}
	logData, ok := dataMap["log"].(map[string]interface{})
	if !ok {
		log.Println("[capture] 主公簿缺少 log 字段，跳过绑定")
		return
	}
	roleName, _ := logData["role_name"].(string)
	serverName := fmt.Sprintf("%v", server[0])
	if roleName == "" || serverName == "" {
		log.Println("[capture] 角色名或服务器名为空，跳过绑定")
		return
	}

	runtime.BindIPs(srcIP, dstIP)
	databaseName := roleName + "_" + serverName
	model.InitDB(databaseName)
	repo.SetDB(model.Conn)
	log.Printf("[capture] 已绑定IP src=%s dst=%s, 数据库=%s", srcIP, dstIP, databaseName)
	databaseSelected = true
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

type Buffer struct {
	Byte   []byte
	offset int
}

func NewBufferFrom(b []byte) *Buffer { return &Buffer{Byte: b} }

func (bb *Buffer) ReadInt() int {
	if bb.offset+4 > len(bb.Byte) {
		return 0
	}
	value := binary.BigEndian.Uint32(bb.Byte[bb.offset : bb.offset+4])
	bb.offset += 4
	return int(value)
}
