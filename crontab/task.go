package crontab

import (
	config "app/conf"
	"app/services"
	"sync"
	"time"
)

var once sync.Once

// InitTask 初始化定时任务（只执行一次）
func InitTask() {
	once.Do(func() {

		// 首次加载所有缓存
		// 更新完赛近30天的缓存数据，写到内存
		services.RefreshMatchRecordCache(0, config.CacheDaysRecent)                        // 近10天
		services.RefreshMatchRecordCache(config.CacheDaysRecent, config.CacheDaysExtended) // 10-30天

		// 更新全民推荐赛程数据，写到内存
		services.RefreshPgameRecommendCache()

		// 清理完赛180天前的redis数据
		services.CleanupExpiredRedisKeys() // 清理过期 Redis key

		// 启动定时刷新
		go startRecentCacheRefreshTimer()   // 每10秒
		go startExtendedCacheRefreshTimer() // 每1小时
		go startCleanupExpiredKeysTimer()   // 每天
	})
}

// startRecentCacheRefreshTimer 近10天缓存刷新（每10秒）
func startRecentCacheRefreshTimer() {
	ticker := time.NewTicker(config.RecentRefreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		services.RefreshMatchRecordCache(0, config.CacheDaysRecent)
		services.RefreshPgameRecommendCache()
	}
}

// startExtendedCacheRefreshTimer 10-30天缓存刷新（每30分钟）
func startExtendedCacheRefreshTimer() {
	ticker := time.NewTicker(config.ExtendedRefreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		services.RefreshMatchRecordCache(config.CacheDaysRecent, config.CacheDaysExtended)
	}
}

// startCleanupExpiredKeysTimer 每天清理6个月前的 Redis key
func startCleanupExpiredKeysTimer() {
	ticker := time.NewTicker(config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		services.CleanupExpiredRedisKeys()
	}
}
