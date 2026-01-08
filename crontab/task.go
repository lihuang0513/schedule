package crontab

import (
	config "app/conf"
	"app/services"
	"sync"
	"time"
)

// 缓存刷新间隔
const (
	RecentRefreshInterval   = 10 * time.Second // 近10天刷新间隔
	ExtendedRefreshInterval = 1 * time.Hour    // 10-30天刷新间隔
)

var once sync.Once

// InitTask 初始化定时任务（只执行一次）
func InitTask() {
	once.Do(func() {
		// 首次加载所有缓存
		services.RefreshMatchRecordCache(0, config.CacheDaysRecent)                        // 近10天
		services.RefreshMatchRecordCache(config.CacheDaysRecent, config.CacheDaysExtended) // 10-30天
		services.RefreshPgameRecommendCache()

		// 启动定时刷新
		go startRecentCacheRefreshTimer()   // 近10天，每10秒
		go startExtendedCacheRefreshTimer() // 10-30天，每1小时
	})
}

// startRecentCacheRefreshTimer 近10天缓存刷新（每10秒）
func startRecentCacheRefreshTimer() {
	ticker := time.NewTicker(RecentRefreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		services.RefreshMatchRecordCache(0, config.CacheDaysRecent)
		services.RefreshPgameRecommendCache()
	}
}

// startExtendedCacheRefreshTimer 10-30天缓存刷新（每30分钟）
func startExtendedCacheRefreshTimer() {
	ticker := time.NewTicker(ExtendedRefreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		services.RefreshMatchRecordCache(config.CacheDaysRecent, config.CacheDaysExtended)
	}
}
