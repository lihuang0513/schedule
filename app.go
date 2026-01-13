package main

import (
	// 导入配置模块，用于初始化和读取配置信息
	config "schedule-api/conf"
	// 导入定时任务模块，用于初始化缓存定时任务
	"schedule-api/crontab"
	// 导入数据模块，可能用于初始化数据结构或数据库连接
	"schedule-api/data"
	// 导入路由模块，用于设置HTTP路由
	"schedule-api/router"
	// 导入endless库，用于创建可以平滑重启的HTTP服务器
	"github.com/fvbock/endless"
	// 导入日志库，用于记录日志信息
	"log"
	// 导入时间库，用于处理时间相关操作
	"time"
)

// init函数在程序启动时初始化必要的环境配置
func init() {
	// 1. 初始化时区
	// 设置东八区时区，确保时间处理的一致性
	loc := time.FixedZone("UTC", 8*3600)
	time.Local = loc
	// 2. 初始化配置
	// 加载配置文件，确保全局配置可用
	config.InitConfig()
	// 3. 初始化扩展库
	// 初始化数据模块，可能是数据库连接或其他数据结构
	data.InitData()
	// 4. 初始化定时任务（完赛数据缓存，每10秒刷新近10天数据）
	crontab.InitTask()
}

// main函数是程序的入口点
func main() {
	// 1，注册路由
	// 设置HTTP路由，处理不同的URL请求
	r := router.SetRouters()
	// 2，配置(使用endless进行重启服务)
	// 创建一个支持平滑重启的HTTP服务器
	s := endless.NewServer(config.Config.Server.Host+":"+config.Config.Server.Port, r)
	// 监听并服务，如果发生错误则记录错误日志
	err := s.ListenAndServe()
	if err != nil {
		log.Printf("server err: %v", err)
		return
	}
}
