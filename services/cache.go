package services

import (
	"fmt"
	"net/http"
	config "schedule-api/conf"
	"schedule-api/data"
	"schedule-api/validate"
	"time"
)

// RefreshCache 刷新指定类型的缓存
// cacheType: recommend（推荐）、finished（完赛）
// date: 完赛类型必传，格式 2006-01-02
func RefreshCache(cacheType, date string) validate.CacheRefreshResult {
	switch cacheType {
	case "recommend":
		RefreshPgameRecommendCache()
		return validate.CacheRefreshResult{Code: http.StatusOK, Success: true, Msg: "推荐缓存刷新成功"}
	case "finished":
		if date == "" {
			return validate.CacheRefreshResult{Code: http.StatusBadRequest, Success: false, Msg: "日期错误"}
		}
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return validate.CacheRefreshResult{Code: http.StatusBadRequest, Success: false, Msg: "日期错误."}
		}
		cacheData := FetchDayMatchRecord(date)
		data.SetMatchRecordCache(date, cacheData)
		return validate.CacheRefreshResult{Code: http.StatusOK, Success: true, Msg: "完赛缓存刷新成功", Date: date}
	default:
		return validate.CacheRefreshResult{Code: http.StatusBadRequest, Success: false, Msg: "参数错误"}
	}
}

// CleanupExpiredRedisKeys 清理6个月前的 Redis key 和内存缓存
// 清理 key：pgame:finished:{date}、pgame:finished:{date}:code、pgame:finished:{date}:hash
func CleanupExpiredRedisKeys() {
	expireDate := time.Now().AddDate(0, 0, -config.CacheDaysExpire)
	redisDeletedCount := 0
	memoryDeletedCount := 0

	// 往前清理1天的数据
	for i := 0; i < 1; i++ {
		date := expireDate.AddDate(0, 0, -i).Format("2006-01-02")

		// 清理 Redis key：数据、版本号、hash
		redisKeys := []string{
			config.PgameLeagueSchedulePrefix + date,
			fmt.Sprintf(config.PgameLeagueScheduleCodeKey, date),
			fmt.Sprintf(config.PgameLeagueScheduleHashKey, date),
		}

		for _, key := range redisKeys {
			if err := data.Rdb.Del(key).Err(); err == nil {
				redisDeletedCount++
			}
		}

		// 清理内存缓存（完赛数据 + 版本号）
		memoryCacheKey := config.MatchRecordCachePrefix + date
		data.Cache.Delete(memoryCacheKey)
		memoryDeletedCount++

		versionCacheKey := fmt.Sprintf(config.PgameLeagueScheduleCodeKey, date)
		data.Cache.Delete(versionCacheKey)
		memoryDeletedCount++
	}

	if redisDeletedCount > 0 || memoryDeletedCount > 0 {
		data.Logger.Printf("清理过期数据完成，Redis: %d 个 key，内存: %d 个 key\n", redisDeletedCount, memoryDeletedCount)
	}
}
