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
func GetMatchRecordList(req validate.MatchRecordListRequest) validate.MatchRecordResponse {
	// 获取用户兴趣运动标签
	userSportsLabels := getUserSportsLabels(req.UserSports)
	if len(userSportsLabels) == 0 {
		return emptyResponse()
	}

	// 获取日期参数
	nextDate := req.NextDate
	if nextDate == "" || nextDate == "__DATE-__" {
		nextDate = time.Now().Format("2006-01-02")
	}

	// 查询存在兴趣赛程的日期列表
	dateList := getScheduleDateList(nextDate, userSportsLabels)
	if len(dateList) == 0 {
		return emptyResponse()
	}

	// 根据日期获取赛程数据
	result := getDayScheduleListByDateList(dateList, userSportsLabels)
	if len(result.List) == 0 {
		return emptyResponse()
	}

	return result
}

// emptyResponse 返回空数据响应
func emptyResponse() validate.MatchRecordResponse {
	return validate.MatchRecordResponse{
		Date:     "",
		DateStr:  "",
		NoData:   1,
		NextDate: "",
		List:     []interface{}{},
	}
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

// getScheduleDateList 从ES查询存在用户兴趣赛程的日期列表
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

// getDayScheduleListByDateList 根据日期列表获取赛程数据
func getDayScheduleListByDateList(dateList []string, userSportsLabels []string) validate.MatchRecordResponse {
	if len(dateList) == 0 {
		return emptyResponse()
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

			return validate.MatchRecordResponse{
				Date:     dayData.Date,
				DateStr:  dayData.DateStr,
				NextDate: nextDate,
				List:     filteredList,
			}
		}
	}

	return emptyResponse()
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

