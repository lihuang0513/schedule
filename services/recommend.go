package services

import (
	config "app/conf"
	"app/data"
	"app/tool"
	"app/validate"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	errorZbbDidEmpty      = "设备参数错误"
	errorDeviceParamEmpty = "设备信息异常"
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

			if existing := dateMap[formatDate]; existing != nil {
				if len(list) > 0 {
					existing.List = append(existing.List, list...)
				}
			} else {
				date, _ := dateObj["date"].(string)
				dateMap[formatDate] = &validate.PgameRecommendDateItem{
					FormatDate: formatDate,
					Date:       date,
					List:       list,
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

func getPGameListParams(c *gin.Context) (request validate.PGameListRequestParams) {
	// 解析直播吧设备参数
	zbbDid := c.Request.FormValue("zbb_did")        // 获取 zbb_did 参数
	userSports := c.Request.FormValue("usersports") // 获取 usersports 参数
	isDebug := c.Request.FormValue("_debug")        // 获取 _debug 参数

	if zbbDid != "" { // 如果 zbb_did 不为空
		// 获取设备号
		deviceParam, err := tool.GetDeviceParam(zbbDid)
		if err != nil {
			data.Logger.Println(errorDeviceParamEmpty) // 记录日志：设备参数为空
		} else {
			// 根据设备平台获取对应的设备ID
			platform := deviceParam["os"] // 获取设备平台
			versionCode := ""             // 初始化版本号为空字符串

			if platform == "android" { // 如果是 Android 平台
				if deviceId, ok := deviceParam["android_id"].(string); ok {
					request.DeviceId = deviceId // 设置设备ID为 android_id
				}
				versionCode = c.Request.FormValue("version_name") // 获取 Android 版本号

			} else if platform == "harmony" {
				if deviceId, ok := deviceParam["udid"].(string); ok {
					request.DeviceId = deviceId
				}
				versionCode = c.Request.FormValue("version_name") // 获取 Android 版本号

			} else { // 如果是 iOS 或其他平台
				if deviceId, ok := deviceParam["udid"].(string); ok {
					request.DeviceId = deviceId // 设置设备ID为 udid
				}
				versionCode = c.Request.FormValue("version_code") // 获取 iOS 版本号
			}

			if os, ok := platform.(string); ok {
				request.Platform = os // 设置平台为字符串类型
			}
			request.Platform = strings.ToLower(request.Platform)        // 将平台名称转换为小写
			request.VersionCode = tool.ConvertVersionToInt(versionCode) // 将版本号转换为整数
		}
	} else {
		data.Logger.Println(errorZbbDidEmpty) // 记录日志：zbb_did 为空
	}

	// 运动兴趣
	request.UserSports = ""
	if userSports != "" {
		request.UserSports = userSports // 设置 UserSports 为用户提供的值
	}

	// 调试模式
	request.IsDebug = false
	if isDebug != "" {
		request.IsDebug = true // 设置 IsDebug 为 true
	}

	// 返回请求结果
	return request
}
