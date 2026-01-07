package data

import (
	"sync"
)

// once 确保 InitData 函数中的初始化逻辑仅被执行一次
var once sync.Once

// ws 用于等待所有初始化任务完成
var ws sync.WaitGroup

// InitData 初始化项目所需的数据和服务
// 该函数使用 once 来确保初始化逻辑在并发环境下只执行一次
// 使用 ws 来等待所有初始化任务完成后再继续执行其他逻辑
func InitData() {
	once.Do(func() {
		// 为所有初始化任务添加到 ws 中
		ws.Add(3)
		// 初始化日志模块
		go initLogger(&ws)
		// 初始化 redis
		go initRedis(&ws)
		// 初始化内存缓存
		go initCache(&ws)
		// 等待所有初始化任务完成
		ws.Wait()
	})
}
