package crontab

import (
	"app/services"
	"time"
)

// 缓存配置
const (
	CacheRefreshInterval = 10 * time.Second // 刷新间隔
)

// InitTask 初始化定时任务
func InitTask() {
	// 首次加载缓存
	services.RefreshMatchRecordCache()
	services.RefreshPgameRecommendCache()

	// 启动定时刷新
	go startCacheRefreshTimer()
}

// startCacheRefreshTimer 启动定时刷新任务
func startCacheRefreshTimer() {
	ticker := time.NewTicker(CacheRefreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		services.RefreshMatchRecordCache()
		services.RefreshPgameRecommendCache()
	}
}
