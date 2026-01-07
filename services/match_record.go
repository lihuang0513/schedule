package services

import (
	config "app/conf"
	"app/data"
	"app/validate"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 赛程状态常量
const (
	StateGoing       = "2" // 进行中
	StateFinished    = "3" // 完赛
	StateDeferred    = "4" // 延期
	StateInterrupted = "5" // 中断
)

// 静态数据接口地址
const StaticRecordURL = "http://s.qiumibao.com/json/record/"

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

// GetMatchRecordList 获取完赛列表
// 解析用户兴趣标签 → ES查询有数据的日期列表 → 静态文件获取赛程 → 过滤用户兴趣
// 分页设计：找到第一天有数据的立即返回，通过 next_date 获取下一页
func GetMatchRecordList(req validate.MatchRecordListRequest) *validate.MatchRecordResponse {
	// 解析用户兴趣标签（"1,2,3" → ["足球", "篮球", "NBA"]）
	userSportsLabels := getUserSportsLabels(req.UserSports)
	if len(userSportsLabels) == 0 {
		return nil
	}

	// 从 ES 查询有数据的日期列表（过去1个月 ~ 指定日期，倒序）
	dateList := getScheduleDateList(req.NextDate, userSportsLabels)
	if len(dateList) == 0 {
		return nil
	}

	// 按日期获取赛程数据，找到第一天有数据的立即返回
	result := getDayScheduleListByDateList(dateList, userSportsLabels)
	if result == nil || len(result.List) == 0 {
		return nil
	}

	return result
}

// PgameLeagueData Redis 中联赛数据结构
type PgameLeagueData struct {
	Date    string                   `json:"date"`
	DateStr string                   `json:"date_str"`
	List    []map[string]interface{} `json:"list"`
}

// MergePgameLeagueData 合并全民联赛数据
// 从 Redis 获取指定联赛数据（key: pgame:league:schedule:日期）
// 根据 saishi_id 去重，按 start_time 倒序排序
func MergePgameLeagueData(result *validate.MatchRecordResponse, date string, leagueIds string) *validate.MatchRecordResponse {
	// 没有传联赛ID，直接返回原结果
	if leagueIds == "" {
		return result
	}

	// 确定查询日期，优先使用完赛结果的日期
	queryDate := date
	if result != nil && result.Date != "" {
		queryDate = result.Date
	}

	// 收集已有的 saishi_id 用于去重
	existingIds := make(map[string]bool)
	if result != nil {
		for _, item := range result.List {
			if m, ok := item.(map[string]interface{}); ok {
				if saishiId, ok := m["saishi_id"].(string); ok {
					existingIds[saishiId] = true
				}
			}
		}
	}

	// 从 Redis 获取联赛数据，自动过滤重复的 saishi_id
	leagueData := getPgameLeagueScheduleFromRedis(queryDate, leagueIds, existingIds)
	if len(leagueData) == 0 {
		return result
	}

	// 合并数据并排序，如果没有完赛数据但是有全民赛事，返回全民赛事
	if result == nil {
		// 完赛没数据，用 Redis 数据构造响应
		sortByStartTime(leagueData)
		return &validate.MatchRecordResponse{
			Date:     queryDate,
			DateStr:  formatDateStr(queryDate),
			NextDate: calcNextDate(queryDate),
			List:     leagueData,
		}
	}

	// 完赛有数据，合并后按 start_time 倒序排序
	result.List = append(result.List, leagueData...)
	sortByStartTime(result.List)
	return result
}

// sortByStartTime 按 start_time 倒序排序（时间大的在前）
func sortByStartTime(list []interface{}) {
	sort.Slice(list, func(i, j int) bool {
		ti := getStartTime(list[i])
		tj := getStartTime(list[j])
		return ti > tj // 倒序：时间大的在前
	})
}

// getStartTime 从 item 中提取 start_time
func getStartTime(item interface{}) int64 {
	if m, ok := item.(map[string]interface{}); ok {
		if st, ok := m["start_time"].(string); ok {
			if ts, err := strconv.ParseInt(st, 10, 64); err == nil {
				return ts
			}
		}
	}
	return 0
}

// getPgameLeagueScheduleFromRedis 从 Redis 获取指定联赛的赛程数据（自动去重）
func getPgameLeagueScheduleFromRedis(date string, leagueIds string, existingIds map[string]bool) []interface{} {
	if date == "" || leagueIds == "" {
		return nil
	}

	// 构建 Redis key
	redisKey := fmt.Sprintf("pgame:league:schedule:%s", date)

	// 从 Redis 获取数据
	jsonStr, err := data.Rdb.Get(redisKey).Result()
	if err != nil {
		return nil
	}

	// 解析 JSON
	var allLeagueData map[string]PgameLeagueData
	if err := json.Unmarshal([]byte(jsonStr), &allLeagueData); err != nil {
		return nil
	}

	// 解析联赛 ID 列表
	ids := strings.Split(leagueIds, ",")
	idSet := make(map[string]bool)
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			idSet[id] = true
		}
	}

	// 筛选指定联赛的数据（同时去重）
	var result []interface{}
	for leagueId, leagueData := range allLeagueData {
		if idSet[leagueId] {
			for _, item := range leagueData.List {
				// 去重：跳过已存在的 saishi_id
				if saishiId, ok := item["saishi_id"].(string); ok {
					if existingIds[saishiId] {
						continue
					}
				}
				result = append(result, item)
			}
		}
	}

	return result
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

// calcNextDate 计算下一页日期（前一天）
func calcNextDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return ""
	}
	return t.AddDate(0, 0, -1).Format("2006-01-02")
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

func getScheduleDateList(date string, userSportsLabels []string) []string {
	var endDate, startDate string
	if date == "" {
		endDate = time.Now().Format("2006-01-02")
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	} else {
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			endDate = time.Now().Format("2006-01-02")
		} else {
			endDate = t.Format("2006-01-02")
		}
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}

	// 将标签转换为小写
	lowerLabels := make([]interface{}, len(userSportsLabels))
	for i, label := range userSportsLabels {
		lowerLabels[i] = strings.ToLower(label)
	}

	// 构建ES查询
	query := map[string]interface{}{
		"_source": []string{"saishi_id", "type", "s_date"},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{
						"terms": map[string]interface{}{
							"state": []string{StateGoing, StateFinished, StateDeferred, StateInterrupted},
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"is_visible": 1,
						},
					},
					map[string]interface{}{
						"range": map[string]interface{}{
							"s_date": map[string]interface{}{
								"gte": startDate,
								"lte": endDate,
							},
						},
					},
					map[string]interface{}{
						"terms": map[string]interface{}{
							"label_has_rec.label_text": lowerLabels,
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"group_by_date": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "s_date",
					"size":  30,
					"order": map[string]interface{}{
						"_key": "desc",
					},
				},
			},
		},
		"size": 0,
	}

	// 发送ES查询请求
	resultJSON := searchSchedulesIndex(query)
	if resultJSON == "" {
		return nil
	}

	// 解析结果
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return nil
	}

	// 提取日期列表
	aggregations, ok := result["aggregations"].(map[string]interface{})
	if !ok {
		return nil
	}

	groupByDate, ok := aggregations["group_by_date"].(map[string]interface{})
	if !ok {
		return nil
	}

	buckets, ok := groupByDate["buckets"].([]interface{})
	if !ok {
		return nil
	}

	dateList := make([]string, 0, len(buckets))
	for _, bucket := range buckets {
		b, ok := bucket.(map[string]interface{})
		if !ok {
			continue
		}
		if keyAsString, ok := b["key_as_string"].(string); ok {
			dateList = append(dateList, keyAsString)
		} else if key, ok := b["key"].(string); ok {
			dateList = append(dateList, key)
		}
	}

	return dateList
}

// searchSchedulesIndex 发送ES搜索请求
func searchSchedulesIndex(query map[string]interface{}) string {
	esConfig := config.Config.ScheduleES
	url := fmt.Sprintf("http://%s:%s/schedules/_search?pretty", esConfig.Host, esConfig.Port)

	queryJSON, err := json.Marshal(query)
	if err != nil {
		data.Logger.Printf("ES查询JSON序列化失败: %v\n", err)
		return ""
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(queryJSON))
	if err != nil {
		data.Logger.Printf("ES请求创建失败: %v\n", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(esConfig.User, esConfig.Password)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		data.Logger.Printf("ES请求执行失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		data.Logger.Printf("ES响应读取失败: %v\n", err)
		return ""
	}

	return string(body)
}

func getDayScheduleListByDateList(dateList []string, userSportsLabels []string) *validate.MatchRecordResponse {
	if len(dateList) == 0 {
		return nil
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for idx, date := range dateList {
		// 获取静态JSON数据
		url := StaticRecordURL + date + ".htm"
		resp, err := client.Get(url)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		var dayData validate.DayScheduleData
		if err := json.Unmarshal(body, &dayData); err != nil {
			continue
		}

		if len(dayData.List) == 0 {
			continue
		}

		// 过滤符合用户兴趣的赛程
		filteredList := make([]interface{}, 0)
		for _, item := range dayData.List {
			if matchUserSportsLabels(item, userSportsLabels) {
				filteredList = append(filteredList, item)
			}
		}

		if len(filteredList) > 0 {
			// 计算 next_date
			var nextDate string
			if idx < len(dateList)-1 {
				nextDate = dateList[idx+1]
			} else {
				// 计算前一天
				t, err := time.Parse("2006-01-02", date)
				if err == nil {
					nextDate = t.AddDate(0, 0, -1).Format("2006-01-02")
				}
			}

			return &validate.MatchRecordResponse{
				Date:     dayData.Date,
				DateStr:  dayData.DateStr,
				NextDate: nextDate,
				List:     filteredList,
			}
		}
	}

	return nil
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
