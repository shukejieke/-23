package repo

import (
	"sort"
	"strconv"
	"strings"

	"stzbHelper/model"
)

func ListTeamUsers(group string) ([]model.TeamUser, error) {
	users, err := loadNormalizedTeamUsers()
	if err != nil {
		return nil, err
	}
	if group == "" {
		return users, nil
	}
	filtered := make([]model.TeamUser, 0, len(users))
	for _, u := range users {
		if u.Group == group {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}

func ListTeamGroups() ([]string, error) {
	metas, err := loadUnionGroupMetas()
	if err != nil {
		return nil, err
	}
	if len(metas) > 0 {
		groups := make([]string, 0, len(metas))
		for _, meta := range metas {
			if meta.GroupName != "" {
				groups = append(groups, meta.GroupName)
			}
		}
		return groups, nil
	}

	users, err := loadNormalizedTeamUsers()
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	groups := make([]string, 0)
	for _, u := range users {
		if u.Group == "" {
			continue
		}
		if _, ok := seen[u.Group]; ok {
			continue
		}
		seen[u.Group] = struct{}{}
		groups = append(groups, u.Group)
	}
	sort.Strings(groups)
	return groups, nil
}

type GroupWuStats struct {
	Group               string `json:"group"`
	MemberCount         int    `json:"member_count"`
	CapturedMemberCount int    `json:"captured_member_count"`
	TotalWu             int    `json:"total_wu"`
	AverageWu           int    `json:"average_wu"`
	ZeroWuCount         int    `json:"zero_wu_count"`
	GroupPower          int64  `json:"group_power"`
	LeaderName          string `json:"leader_name"`
	CaptureTime         int64  `json:"capture_time"`
	Complete            bool   `json:"complete"`
}

func QueryGroupWuStats() ([]GroupWuStats, error) {
	users, err := loadNormalizedTeamUsers()
	if err != nil {
		return nil, err
	}
	metas, err := loadUnionGroupMetas()
	if err != nil {
		return nil, err
	}

	type agg struct {
		count int
		total int
		zero  int
	}
	aggMap := map[string]*agg{}
	for _, u := range users {
		group := strings.TrimSpace(u.Group)
		if group == "" {
			group = "未分组"
		}
		if _, ok := aggMap[group]; !ok {
			aggMap[group] = &agg{}
		}
		aggMap[group].count++
		aggMap[group].total += u.Wu
		if u.Wu == 0 {
			aggMap[group].zero++
		}
	}

	stats := make([]GroupWuStats, 0, len(metas)+len(aggMap))
	seen := map[string]struct{}{}
	for _, meta := range metas {
		group := strings.TrimSpace(meta.GroupName)
		if group == "" {
			continue
		}
		seen[group] = struct{}{}
		a := aggMap[group]
		row := GroupWuStats{
			Group:               group,
			MemberCount:         meta.MemberCount,
			GroupPower:          meta.Power,
			LeaderName:          meta.LeaderName,
			CaptureTime:         meta.CaptureTime,
			CapturedMemberCount: 0,
			Complete:            false,
		}
		if a != nil {
			row.CapturedMemberCount = a.count
			row.TotalWu = a.total
			row.ZeroWuCount = a.zero
			if a.count > 0 {
				row.AverageWu = a.total / a.count
			}
		}
		row.Complete = meta.MemberCount > 0 && row.CapturedMemberCount >= meta.MemberCount
		stats = append(stats, row)
	}

	for group, a := range aggMap {
		if _, ok := seen[group]; ok {
			continue
		}
		row := GroupWuStats{
			Group:               group,
			MemberCount:         a.count,
			CapturedMemberCount: a.count,
			TotalWu:             a.total,
			ZeroWuCount:         a.zero,
			Complete:            true,
		}
		if a.count > 0 {
			row.AverageWu = a.total / a.count
		}
		stats = append(stats, row)
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Complete != stats[j].Complete {
			return stats[i].Complete
		}
		if stats[i].CapturedMemberCount != stats[j].CapturedMemberCount {
			return stats[i].CapturedMemberCount > stats[j].CapturedMemberCount
		}
		return stats[i].Group < stats[j].Group
	})

	return stats, nil
}

func FindUsersByGroups(groups []string) ([]model.TeamUser, error) {
	users, err := loadNormalizedTeamUsers()
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return users, nil
	}
	groupSet := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		groupSet[group] = struct{}{}
	}
	filtered := make([]model.TeamUser, 0, len(users))
	for _, u := range users {
		if _, ok := groupSet[u.Group]; ok {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}

func loadNormalizedTeamUsers() ([]model.TeamUser, error) {
	var users []model.TeamUser
	if err := DB().Where("id > 0").Find(&users).Error; err != nil {
		return nil, err
	}
	codeMap := model.LoadGroupCodeNameMap()
	for i := range users {
		users[i].Group = normalizeGroupName(users[i].Group, codeMap)
	}
	sort.Slice(users, func(i, j int) bool {
		if users[i].Wu != users[j].Wu {
			return users[i].Wu > users[j].Wu
		}
		if users[i].Power != users[j].Power {
			return users[i].Power > users[j].Power
		}
		return users[i].Id > users[j].Id
	})
	return users, nil
}

func loadUnionGroupMetas() ([]model.UnionGroupMeta, error) {
	var metas []model.UnionGroupMeta
	err := DB().Order("group_id ASC").Find(&metas).Error
	return metas, err
}

func normalizeGroupName(group string, codeMap map[int]string) string {
	group = strings.TrimSpace(group)
	if group == "" {
		return "未分组"
	}
	if mapped, ok := codeMap[toInt(group)]; ok && mapped != "" {
		return mapped
	}
	return group
}

func toInt(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
