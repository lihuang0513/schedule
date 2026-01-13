package data

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"log"
	"os"
	c "schedule-api/conf"
	"schedule-api/tool"
	"sync"
	"time"
)

// Logger 日志客户端
var Logger *log.Logger

func initLogger(ws *sync.WaitGroup) {
	defer ws.Done()
	if c.Config.Server.Debug {
		Logger = log.New(os.Stdout, "", log.LstdFlags)
		return
	}
	// 判断文件夹是否存在没有则创建
	tool.MakeDir(c.Config.Server.LogDir)
	// 获取日志文件句柄
	logFile, err := rotatelogs.New(
		c.Config.Server.LogDir+"%Y-%m-%d-buried_point.log",
		//设置一天产生一个日志文件
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Fatalf("创建文件失败: %v", err)
	}
	// 设置存储位置
	Logger = log.New(logFile, "", log.Lshortfile|log.LstdFlags)
}
