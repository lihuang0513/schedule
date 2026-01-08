package data

import (
	config "app/conf"
	"app/validate"
	"encoding/json"
	"reflect"
	"runtime"
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

// CacheStats 缓存统计信息
type CacheStats struct {
	ItemCount   int              `json:"item_count"`    // 缓存项数量
	CacheSizeMB float64          `json:"cache_size_mb"` // 缓存估算大小（MB）
	Keys        []string         `json:"keys"`          // 所有 key
	KeyDetails  []CacheKeyDetail `json:"key_details"`   // 每个 key 的详情
	MemStats    RuntimeMemStats  `json:"mem_stats"`     // 运行时内存统计
}

// CacheKeyDetail 单个缓存项详情
type CacheKeyDetail struct {
	Key       string  `json:"key"`
	SizeMB    float64 `json:"size_mb"`    // 估算大小（MB）
	ListCount int     `json:"list_count"` // List 数量（如果有）
}

// RuntimeMemStats 运行时内存统计
type RuntimeMemStats struct {
	AllocMB      float64 `json:"alloc_mb"`       // 当前分配的内存（MB）
	TotalAllocMB float64 `json:"total_alloc_mb"` // 累计分配的内存（MB）
	SysMB        float64 `json:"sys_mb"`         // 从系统获取的内存（MB）
	NumGC        uint32  `json:"num_gc"`         // GC 次数
}

// GetCacheStats 获取缓存统计信息
func GetCacheStats() CacheStats {
	items := Cache.Items()
	keys := make([]string, 0, len(items))
	keyDetails := make([]CacheKeyDetail, 0, len(items))
	var totalSize int64

	for k, item := range items {
		keys = append(keys, k)

		// 估算每个 key 的大小
		jsonBytes, _ := json.Marshal(item.Object)
		size := len(jsonBytes)
		totalSize += int64(size)

		detail := CacheKeyDetail{
			Key:    k,
			SizeMB: float64(size) / 1024 / 1024,
		}

		// 获取 List 数量
		if cacheData, ok := item.Object.(*validate.DayMatchRecordCache); ok {
			detail.ListCount = len(cacheData.List)
		} else if cacheData, ok := item.Object.(*validate.PgameRecommendCache); ok {
			detail.ListCount = len(cacheData.Data)
		}

		keyDetails = append(keyDetails, detail)
	}

	// 获取运行时内存统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return CacheStats{
		ItemCount:   Cache.ItemCount(),
		CacheSizeMB: float64(totalSize) / 1024 / 1024,
		Keys:        keys,
		KeyDetails:  keyDetails,
		MemStats: RuntimeMemStats{
			AllocMB:      float64(m.Alloc) / 1024 / 1024,
			TotalAllocMB: float64(m.TotalAlloc) / 1024 / 1024,
			SysMB:        float64(m.Sys) / 1024 / 1024,
			NumGC:        m.NumGC,
		},
	}
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
// 过期策略：10天内不过期，10-30天过期1小时，30天前过期1天
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

	// 计算过期时间
	expiration := cache.NoExpiration
	if t, err := time.Parse("2006-01-02", date); err == nil {
		daysSince := int(time.Since(t).Hours() / 24)
		if daysSince > config.CacheDaysExtended {
			// 30天前：1天过期
			expiration = 24 * time.Hour
		} else if daysSince > config.CacheDaysRecent {
			// 10-30天：1小时过期
			expiration = time.Hour
		}
		// 10天内：不过期定时器更新
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
