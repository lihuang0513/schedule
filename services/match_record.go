package services

import (
	config "app/conf"
	"app/data"
	"app/validate"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// httpClient 复用 HTTP 客户端（避免每次请求创建新连接）
var httpClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,              // 最大空闲连接
		MaxIdleConnsPerHost: 20,               // 单主机最大空闲连接
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时
	},
}

// sportsLabelMap 用户运动标签映射
var sportsLabelMap = map[string][]string{
	"1":  {"足球"},
	"2":  {"篮球"},
	"3":  {"NBA"},
	"4":  {"电竞"},
	"41": {"英雄联盟"},
	"42": {"DOTA2"},
	"43": {"绝地求生"},
	"44": {"王者荣耀"},
	"45": {"无畏契约"},
	"51": {"F1"},
	"52": {"网球"},
	"53": {"斯诺克"},
	"54": {"NFL"},
	"55": {"MLB"},
	"56": {"NHL"},
	"57": {"拳击"},
	"58": {"UFC"},
	"59": {"高尔夫"},
	"60": {"田径"},
	"61": {"排球"},
	"62": {"羽毛球"},
	"63": {"乒乓球"},
}

// ========== 完赛缓存 ==========

// RefreshMatchRecordCache 刷新指定天数范围的完赛缓存（并行获取，统一写入）
// startDay: 起始偏移天数（0=今天）
// endDay: 结束偏移天数（不包含）
// 例：RefreshMatchRecordCache(0, 10) 刷新近10天，RefreshMatchRecordCache(10, 30) 刷新10-30天
// 增量更新：先检查 Redis 版本号，版本号有变化才拉取数据
func RefreshMatchRecordCache(startDay, endDay int) {
	now := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 临时存储结果，避免并发写入
	results := make(map[string]*validate.DayMatchRecordCache)
	updatedCount := 0

	// 并行获取数据
	for i := startDay; i < endDay; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			date := now.AddDate(0, 0, -offset).Format("2006-01-02")

			// 检查版本号是否变化（通用方法）
			versionKey := fmt.Sprintf(config.PgameLeagueScheduleCodeKey, date)
			if !data.CheckVersionChanged(versionKey) {
				return // 版本号未变化，跳过
			}

			cacheData := FetchDayMatchRecord(date)

			// 加锁写入临时 map
			mu.Lock()
			results[date] = cacheData
			updatedCount++
			mu.Unlock()
		}(i)
	}

	// 等待所有协程完成
	wg.Wait()

	// 统一写入缓存
	for date, cacheData := range results {
		data.SetMatchRecordCache(date, cacheData)
	}

	if updatedCount > 0 {
		data.Logger.Printf("完赛缓存刷新完成，范围 %d-%d 天，更新 %d 天\n", startDay, endDay, updatedCount)
	}
}

// FetchDayMatchRecord 获取某一天的完赛数据（静态文件 + 全民赛程）
func FetchDayMatchRecord(date string) *validate.DayMatchRecordCache {
	var allList []interface{}

	// 1. 从静态文件获取完赛数据
	staticList := fetchStaticMatchRecord(date)
	if len(staticList) > 0 {
		allList = append(allList, staticList...)
	}

	// 2. 从 Redis 获取全民赛程数据
	pgameList := fetchPgameLeagueRecord(date)
	if len(pgameList) > 0 {
		// 去重：用 saishi_id
		existingIds := make(map[string]bool)
		for _, item := range allList {
			if m, ok := item.(map[string]interface{}); ok {
				if saishiId, ok := m["saishi_id"].(string); ok {
					existingIds[saishiId] = true
				}
			}
		}

		for _, item := range pgameList {
			if m, ok := item.(map[string]interface{}); ok {
				if saishiId, ok := m["saishi_id"].(string); ok {
					if !existingIds[saishiId] {
						allList = append(allList, item)
					}
				}
			}
		}
	}

	// 3. 按 start_time 倒序排序
	if len(allList) > 0 {
		sortByStartTime(allList)
	}

	return &validate.DayMatchRecordCache{
		Date:     date,
		DateStr:  formatDateStr(date),
		List:     allList,
		UpdateAt: time.Now(),
	}
}

// LoadAndCacheMatchRecord 加载数据并写入缓存
func LoadAndCacheMatchRecord(date string) *validate.DayMatchRecordCache {
	cacheData := FetchDayMatchRecord(date)
	data.SetMatchRecordCache(date, cacheData)
	return cacheData
}

// fetchStaticMatchRecord 从静态文件获取完赛数据
func fetchStaticMatchRecord(date string) []interface{} {
	url := config.StaticRecordURL + date + ".htm"

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var dayData validate.DayScheduleData
	if err := json.Unmarshal(body, &dayData); err != nil {
		return nil
	}

	// 转换为 []interface{}
	result := make([]interface{}, 0, len(dayData.List))
	for _, item := range dayData.List {
		result = append(result, item)
	}

	return result
}

// fetchPgameLeagueRecord 从 Redis 获取全民赛程数据（所有联赛）
func fetchPgameLeagueRecord(date string) []interface{} {
	redisKey := config.PgameLeagueSchedulePrefix + date

	jsonStr, err := data.Rdb.Get(redisKey).Result()
	if err != nil {
		return nil
	}

	var allLeagueData map[string]validate.PgameLeagueData
	if err := json.Unmarshal([]byte(jsonStr), &allLeagueData); err != nil {
		return nil
	}

	var result []interface{}
	for _, leagueData := range allLeagueData {
		for _, item := range leagueData.List {
			result = append(result, item)
		}
	}

	return result
}

// GetMatchRecordList 获取完赛列表（从内存缓存读取）
// 解析用户兴趣标签 → 从缓存读取近10天数据 → 过滤用户兴趣+区域
// 分页设计：找到第一天有数据的立即返回，通过 next_date 获取下一页
func GetMatchRecordList(req validate.MatchRecordListRequest) *validate.MatchRecordResponse {
	// 构建过滤参数
	filter := buildMatchListFilter(req)
	if len(filter.UserSportsLabels) == 0 {
		return nil
	}

	// 计算查询的起始日期
	startDate, err := time.Parse("2006-01-02", req.NextDate)
	if err != nil {
		startDate = time.Now()
	}

	// 限制最大可查询范围：6个月（防止用户无限往前翻导致内存增长）
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	maxDays := 180 // 6个月
	if today.Sub(startDate).Hours()/24 > float64(maxDays) {
		return nil // 超出范围，返回空
	}

	// 从 startDate 往前读取 30 天数据，找到第一天有符合用户兴趣的数据
	for i := 0; i < config.CacheDaysExtended; i++ {
		date := startDate.AddDate(0, 0, -i).Format("2006-01-02")

		// 从缓存获取当天数据
		cacheData, needReload := data.GetMatchRecordCache(date)

		// 缓存未加载，重新获取并写入缓存
		if needReload {
			cacheData = LoadAndCacheMatchRecord(date)
		}

		// 缓存已加载但没有数据，跳过（不需要重新获取）
		if cacheData == nil || len(cacheData.List) == 0 {
			continue
		}

		// 根据过滤条件过滤
		filteredList := filterMatchList(cacheData.List, filter)
		if len(filteredList) == 0 {
			continue
		}

		// 计算 next_date（下一天继续查）
		nextDate := startDate.AddDate(0, 0, -(i + 1)).Format("2006-01-02")

		return &validate.MatchRecordResponse{
			Date:     cacheData.Date,
			DateStr:  cacheData.DateStr,
			NextDate: nextDate,
			List:     filteredList,
		}
	}

	return nil
}

// buildMatchListFilter 构建过滤参数
func buildMatchListFilter(req validate.MatchRecordListRequest) *validate.MatchListFilter {
	filter := &validate.MatchListFilter{
		UserSportsLabels: getUserSportsLabels(req.UserSports),
	}

	// 解析联赛ID过滤集合
	if req.PgameLeagueIds != "" {
		ids := strings.Split(req.PgameLeagueIds, ",")
		filter.LeagueIdSet = make(map[string]struct{}, len(ids))
		for _, id := range ids {
			if id = strings.TrimSpace(id); id != "" {
				filter.LeagueIdSet[id] = struct{}{}
			}
		}
	}

	return filter
}

// filterMatchList 根据过滤条件过滤赛程列表
func filterMatchList(list []interface{}, filter *validate.MatchListFilter) []interface{} {
	filtered := make([]interface{}, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		// 用户兴趣过滤
		if len(filter.UserSportsLabels) > 0 && !matchUserSportsLabels(m, filter.UserSportsLabels) {
			continue
		}
		// 联赛ID过滤：不存在则保留，存在则判断
		if filter.LeagueIdSet != nil {
			if leagueId, exists := m["pgame_league_id"].(string); exists {
				if _, ok := filter.LeagueIdSet[leagueId]; !ok {
					continue
				}
			}
		}
		filtered = append(filtered, item)
	}
	return filtered
}

// sortByStartTime 按 start_time 倒序排序（时间大的在前）
func sortByStartTime(list []interface{}) {
	sort.Slice(list, func(i, j int) bool {
		ti := getStartTime(list[i])
		tj := getStartTime(list[j])
		return ti > tj // 倒序：时间大的在前
	})
}

// getStartTime 从 item 中提取 start_time（兼容 string 和 number 类型）
func getStartTime(item interface{}) int64 {
	if m, ok := item.(map[string]interface{}); ok {
		switch v := m["start_time"].(type) {
		case string:
			if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
				return ts
			}
		case float64:
			return int64(v)
		case int64:
			return v
		}
	}
	return 0
}

// formatDateStr 格式化日期字符串，如 "1月07日 星期二"
func formatDateStr(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return ""
	}
	weekdays := []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}
	return fmt.Sprintf("%d月%02d日 %s", t.Month(), t.Day(), weekdays[t.Weekday()])
}

// getUserSportsLabels 获取用户兴趣运动标签
func getUserSportsLabels(userSports string) []string {
	if userSports == "" {
		return nil
	}

	userSports = strings.Trim(userSports, ", ")
	sports := strings.Split(userSports, ",")

	labelArr := make([]string, 0)
	labelMap := make(map[string]bool) // 用于去重

	for _, sport := range sports {
		sport = strings.TrimSpace(sport)
		if labels, ok := sportsLabelMap[sport]; ok {
			for _, label := range labels {
				if !labelMap[label] {
					labelMap[label] = true
					labelArr = append(labelArr, label)
				}
			}
		} else {
			// 未知的运动类型，添加"其他"和"综合"
			if !labelMap["其他"] {
				labelMap["其他"] = true
				labelArr = append(labelArr, "其他")
			}
			if !labelMap["综合"] {
				labelMap["综合"] = true
				labelArr = append(labelArr, "综合")
			}
		}
	}

	return labelArr
}

// matchUserSportsLabels 检查赛程是否匹配用户兴趣标签
func matchUserSportsLabels(item map[string]interface{}, userSportsLabels []string) bool {
	// 获取赛程的标签
	labelList := extractLabels(item)

	// 检查赛程集合
	if dataType, ok := item["data_type"].(string); ok && dataType == "schedule_collection" {
		if list, ok := item["list"].([]interface{}); ok {
			for _, collectionItem := range list {
				if ci, ok := collectionItem.(map[string]interface{}); ok {
					collectionLabels := extractLabels(ci)
					labelList = append(labelList, collectionLabels...)
				}
			}
		}
	}

	// 去重
	labelList = uniqueStrings(labelList)

	// 检查是否有交集
	return hasIntersection(labelList, userSportsLabels)
}

// extractLabels 从赛程项中提取标签列表
func extractLabels(item map[string]interface{}) []string {
	labelStr, ok := item["label"].(string)
	if !ok || labelStr == "" {
		return nil
	}
	return strings.Split(labelStr, ",")
}

// uniqueStrings 字符串切片去重
func uniqueStrings(strs []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// hasIntersection 检查两个字符串切片是否有交集
func hasIntersection(a, b []string) bool {
	set := make(map[string]bool)
	for _, s := range b {
		set[s] = true
	}
	for _, s := range a {
		if set[s] {
			return true
		}
	}
	return false
}
