package data

import (
	config "app/conf"
	"app/validate"
	"reflect"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// Cache 全局内存缓存
var Cache *cache.Cache

// initCache 初始化缓存
func initCache(ws *sync.WaitGroup) {
	defer ws.Done()
	// 设置超时时间和清理时间
	Cache = cache.New(5*time.Minute, 10*time.Minute)
}

// ========== 完赛数据缓存 ==========

// getMatchRecordCacheKey 生成完赛缓存 key
func getMatchRecordCacheKey(date string) string {
	return config.MatchRecordCachePrefix + date
}

// GetMatchRecordCache 从缓存获取完赛数据
// 返回值：缓存数据, 是否需要重新加载（缓存不存在时需要重新加载）
func GetMatchRecordCache(date string) (*validate.DayMatchRecordCache, bool) {
	key := getMatchRecordCacheKey(date)
	if cached, found := Cache.Get(key); found {
		if cacheData, ok := cached.(*validate.DayMatchRecordCache); ok {
			return cacheData, false // 缓存存在，不需要重新加载
		}
	}
	return nil, true // 缓存不存在，需要重新加载
}

// SetMatchRecordCache 设置缓存
// 10天内数据不过期（由定时任务刷新），10天前数据过期时间1分钟
// 先读取现有缓存，数据无变化则跳过写入
func SetMatchRecordCache(date string, cacheData *validate.DayMatchRecordCache) {
	key := getMatchRecordCacheKey(date)
	oldData, found := Cache.Get(key)
	if found {
		if old, ok := oldData.(*validate.DayMatchRecordCache); ok {
			if reflect.DeepEqual(old.List, cacheData.List) {
				return // 数据无变化，跳过写入
			}
		}
	}

	// 计算过期时间：10天内不过期（有定时任务10s刷新），10天前过期5分钟
	expiration := cache.NoExpiration
	if t, err := time.Parse("2006-01-02", date); err == nil {
		if time.Since(t) > time.Duration(config.CacheDays)*24*time.Hour {
			expiration = time.Minute * 5
		}
	}

	Cache.Set(key, cacheData, expiration)
}

// ========== 推荐数据缓存 ==========

// GetPgameRecommendCache 从缓存获取推荐数据
// 返回值：缓存数据, 是否需要重新加载（缓存不存在时需要重新加载）
func GetPgameRecommendCache() (*validate.PgameRecommendCache, bool) {
	if cached, found := Cache.Get(config.PgameRecommendCacheKey); found {
		if cacheData, ok := cached.(*validate.PgameRecommendCache); ok {
			return cacheData, false // 缓存存在，不需要重新加载
		}
	}
	return nil, true // 缓存不存在，需要重新加载
}

// SetPgameRecommendCache 设置推荐数据缓存
// 先读取现有缓存，数据无变化则跳过写入
func SetPgameRecommendCache(cacheData *validate.PgameRecommendCache) {
	oldData, found := Cache.Get(config.PgameRecommendCacheKey)
	if found {
		if old, ok := oldData.(*validate.PgameRecommendCache); ok {
			if reflect.DeepEqual(old.Data, cacheData.Data) {
				return // 数据无变化，跳过写入
			}
		}
	}
	Cache.Set(config.PgameRecommendCacheKey, cacheData, cache.NoExpiration)
}
