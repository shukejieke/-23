package runtime

import (
	"sync"

	"stzbHelper/global"
)

// runtime 状态层：集中管理运行期可变状态，
// 避免在 handler/capture/parser 各处散写全局变量。
var mu sync.RWMutex

func EnableReport(pos int) {
	mu.Lock()
	defer mu.Unlock()
	global.ExVar.NeededReportPos = pos
	global.ExVar.NeedGetReport = true
}

func DisableReport() {
	mu.Lock()
	defer mu.Unlock()
	global.ExVar.NeededReportPos = 0
	global.ExVar.NeedGetReport = false
}

func EnableBattleReport() {
	mu.Lock()
	defer mu.Unlock()
	global.ExVar.NeedGetBattleData = true
}

func DisableBattleReport() {
	mu.Lock()
	defer mu.Unlock()
	global.ExVar.NeedGetBattleData = false
}

// BindIPs 在首次识别主公簿后，固化本地与目标服务器 IP。
func BindIPs(src, dst string) {
	mu.Lock()
	defer mu.Unlock()
	global.OnlySrcIp = src
	global.OnlyDstIp = dst
}

// GetBoundIPs 供抓包链路过滤非目标连接使用。
func GetBoundIPs() (string, string) {
	mu.RLock()
	defer mu.RUnlock()
	return global.OnlySrcIp, global.OnlyDstIp
}
