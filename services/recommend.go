package services

import (
	"encoding/json"
	config "schedule-api/conf"
	"schedule-api/data"
	"schedule-api/validate"
	"sort"
	"strings"
	"time"
)

// ========== 推荐数据缓存 ==========

// RefreshPgameRecommendCache 刷新推荐数据缓存
// 增量更新：先检查 Redis 版本号，版本号有变化才拉取数据
func RefreshPgameRecommendCache() {
	if !data.CheckVersionChanged(config.PgameRecommendCodeKey) {
		return // 版本号未变化，跳过
	}

	cacheData := FetchPgameRecommend()
	data.SetPgameRecommendCache(cacheData)
	data.Logger.Println("推荐缓存刷新完成")
}

// FetchPgameRecommend 从 Redis 获取推荐数据
func FetchPgameRecommend() *validate.PgameRecommendCache {
	jsonStr, err := data.Rdb.Get(config.PgameRecommendRedisKey).Result()
	if err != nil {
		return &validate.PgameRecommendCache{
			Data:     nil,
			UpdateAt: time.Now(),
		}
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &validate.PgameRecommendCache{
			Data:     nil,
			UpdateAt: time.Now(),
		}
	}

	return &validate.PgameRecommendCache{
		Data:     result,
		UpdateAt: time.Now(),
	}
}

// GetPgameRecommend 从缓存获取全民赛事推荐数据
func GetPgameRecommend() map[string]interface{} {
	cacheData, needReload := data.GetPgameRecommendCache()

	// 缓存未加载，重新获取
	if needReload {
		cacheData = FetchPgameRecommend()
		data.SetPgameRecommendCache(cacheData)
	}

	if cacheData == nil {
		return nil
	}
	return cacheData.Data
}

// GetPgameRecommendFormatted 获取格式化后的推荐数据（按联赛ID过滤）
// pgameLeagueIds: 逗号分隔的联赛ID列表，如 "3729,1234"
func GetPgameRecommendFormatted(pgameLeagueIds string) []validate.PgameRecommendDateItem {
	rawData := GetPgameRecommend()
	if rawData == nil || len(rawData) == 0 {
		return nil
	}

	// 解析联赛ID过滤集合
	var leagueIdSet map[string]struct{}
	if pgameLeagueIds != "" {
		ids := strings.Split(pgameLeagueIds, ",")
		leagueIdSet = make(map[string]struct{}, len(ids))
		for _, id := range ids {
			if id = strings.TrimSpace(id); id != "" {
				leagueIdSet[id] = struct{}{}
			}
		}
	}

	dateMap := make(map[string]*validate.PgameRecommendDateItem)
	idSet := make(map[string]map[string]struct{}) // key: formatDate, value: 已存在的 id 集合

	for leagueId, leagueData := range rawData {
		// 联赛ID过滤
		if leagueIdSet != nil {
			if _, ok := leagueIdSet[leagueId]; !ok {
				continue
			}
		}

		dateList, ok := leagueData.([]interface{})
		if !ok {
			continue
		}

		for i := range dateList {
			dateObj, ok := dateList[i].(map[string]interface{})
			if !ok {
				continue
			}

			formatDate, _ := dateObj["formatDate"].(string)
			if formatDate == "" {
				continue
			}

			list, _ := dateObj["list"].([]interface{})

			// 初始化该日期的数据结构
			if dateMap[formatDate] == nil {
				date, _ := dateObj["date"].(string)
				dateMap[formatDate] = &validate.PgameRecommendDateItem{
					FormatDate: formatDate,
					Date:       date,
					List:       []interface{}{},
				}
				idSet[formatDate] = make(map[string]struct{})
			}

			// 按 id 去重添加
			for _, item := range list {
				match, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				id, _ := match["id"].(string)
				if id == "" {
					continue
				}
				if _, exists := idSet[formatDate][id]; !exists {
					idSet[formatDate][id] = struct{}{}
					dateMap[formatDate].List = append(dateMap[formatDate].List, item)
				}
			}
		}
	}

	if len(dateMap) == 0 {
		return nil
	}

	// 直接分配精确容量
	result := make([]validate.PgameRecommendDateItem, 0, len(dateMap))
	for _, item := range dateMap {
		result = append(result, *item)
	}

	// 按日期排序（直接对结果排序，避免额外的 slice）
	sort.Slice(result, func(i, j int) bool {
		return result[i].FormatDate < result[j].FormatDate
	})

	return result
}
