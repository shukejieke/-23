package runtime

import (
	"sync"
	"time"
)

const (
	DatasetTeamUser            = "team_user"
	DatasetReport              = "report"
	DatasetBattleReport        = "battle_report"
	DatasetUnionLeaderboard    = "union_leaderboard"
	DatasetPersonalLeaderboard = "personal_leaderboard"
	DatasetUnionGroupMeta      = "union_group_meta"
)

type FreshnessItem struct {
	Dataset          string `json:"dataset"`
	LastCaptureUnix  int64  `json:"last_capture_unix"`
	AgeSec           int64  `json:"age_sec"`
	TTLSeconds       int64  `json:"ttl_seconds"`
	Status           string `json:"status"` // missing|fresh|stale
	TriggerHint      string `json:"trigger_hint"`
	RecommendedScene string `json:"recommended_scene"`
}

type datasetRule struct {
	ttlSeconds       int64
	triggerHint      string
	recommendedScene string
}

var (
	freshMu         sync.RWMutex
	lastCaptureUnix = map[string]int64{}
	freshRules      = map[string]datasetRule{
		DatasetTeamUser: {
			ttlSeconds:       30 * 60,
			triggerHint:      "打开游戏同盟成员列表并停留数秒",
			recommendedScene: "同盟成员页面",
		},
		DatasetReport: {
			ttlSeconds:       10 * 60,
			triggerHint:      "开启攻城战报采集后，进入目标坐标战报并持续下拉",
			recommendedScene: "攻城战报页面",
		},
		DatasetBattleReport: {
			ttlSeconds:       10 * 60,
			triggerHint:      "开启详细战报采集后，进入战报详情并持续翻页",
			recommendedScene: "战报详情页面",
		},
		DatasetUnionLeaderboard: {
			ttlSeconds:       60 * 60,
			triggerHint:      "打开游戏排行榜中的同盟榜页签",
			recommendedScene: "排行榜-同盟榜",
		},
		DatasetPersonalLeaderboard: {
			ttlSeconds:       60 * 60,
			triggerHint:      "打开游戏排行榜中的个人榜页签",
			recommendedScene: "排行榜-个人榜",
		},
		DatasetUnionGroupMeta: {
			ttlSeconds:       60 * 60,
			triggerHint:      "打开同盟相关分组/成员页面，触发分组元数据下发",
			recommendedScene: "同盟分组页面",
		},
	}
)

func TouchDataset(dataset string) {
	freshMu.Lock()
	lastCaptureUnix[dataset] = time.Now().Unix()
	freshMu.Unlock()
}

func FreshnessSnapshot() []FreshnessItem {
	now := time.Now().Unix()
	items := make([]FreshnessItem, 0, len(freshRules))

	freshMu.RLock()
	defer freshMu.RUnlock()

	for dataset, rule := range freshRules {
		last := lastCaptureUnix[dataset]
		status := "missing"
		ageSec := int64(0)
		if last > 0 {
			ageSec = now - last
			if ageSec <= rule.ttlSeconds {
				status = "fresh"
			} else {
				status = "stale"
			}
		}
		items = append(items, FreshnessItem{
			Dataset:          dataset,
			LastCaptureUnix:  last,
			AgeSec:           ageSec,
			TTLSeconds:       rule.ttlSeconds,
			Status:           status,
			TriggerHint:      rule.triggerHint,
			RecommendedScene: rule.recommendedScene,
		})
	}
	return items
}
