package router

import (
	"bytes"
	// 导入项目内部的配置包，用于读取和解析配置信息
	c "app/conf"
	// 导入项目内部的控制器包，处理HTTP请求和响应
	"app/controller"
	// 导入项目内部的工具包，提供辅助功能和工具函数
	"app/tool"
	// 导入gin的gzip中间件，用于压缩HTTP响应体
	"github.com/gin-contrib/gzip"
	// 导入gin框架，用于构建web服务器
	"github.com/gin-gonic/gin"
	// 导入日志轮转库，用于生成滚动日志文件
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	// 导入标准库的io包，提供基本的输入输出功能
	"io"
	// 导入标准库的io/ioutil包，提供旧版的文件和IO工具函数
	"io/ioutil"
	// 导入标准库的时间包，用于处理时间和时区相关操作
	"time"
)

// SetRouters 设置路由
func SetRouters() *gin.Engine {
	var r *gin.Engine

	if c.Config.Server.Debug == false {
		// 生产模式
		r = ReleaseRouter()
	} else {
		// 开发调试模式
		r = gin.Default()
	}

	// 使用Gzip中间件
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	match := r.Group("/match")
	{
		// 推荐
		match.GET("/pgame_recommend/list", controller.RecommendList)
	}

	// 完赛列表
	matchRecord := r.Group("/match_record")
	{
		matchRecord.GET("/list", controller.MatchRecordList)
	}

	return r
}

// ReleaseRouter 生产模式使用官方建议设置为 release 模式
func ReleaseRouter() *gin.Engine {
	// 切换到生产模式
	gin.SetMode(gin.ReleaseMode)
	// 禁用 gin 输出接口访问日志
	gin.DefaultWriter = ioutil.Discard
	// 判断文件夹是否存在没有则创建
	tool.MakeDir(c.Config.Server.LogDir)
	// 记录到文件。
	f, _ := rotatelogs.New(
		c.Config.Server.LogDir+"%Y-%m-%d_log.log",
		//设置一天产生一个日志文件
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	//gin.DefaultWriter = io.MultiWriter(f)

	engine := gin.New()
	//日志记录
	//engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	//	if param.StatusCode == 404 {
	//		return ""
	//	}
	//	return fmt.Sprintf("客户端IP:%s,请求时间:[%s],请求方式:%s,请求地址:%s,http协议版本:%s,请求状态码:%d,响应时间:%s,客户端:%s，返回大小:%d\n",
	//		param.ClientIP,
	//		param.TimeStamp.Format(time.RFC1123),
	//		param.Method,
	//		param.Path,
	//		param.Request.Proto,
	//		param.StatusCode,
	//		param.Latency,
	//		param.Request.UserAgent(),
	//		param.BodySize,
	//	)
	//}))
	// 应用崩溃写入日志同时通知ES
	tee := newTeeWriter(f)
	engine.Use(gin.RecoveryWithWriter(tee))
	return engine
}

// TeeWriter 是一个io.Writer，同时进行日志记录和通知第三方
type teeWriter struct {
	mainWriter io.Writer
}

// 创建一个新的TeeWriter
func newTeeWriter(main io.Writer) *teeWriter {
	return &teeWriter{
		mainWriter: main,
	}
}

// Write 实现io.Writer接口
func (t *teeWriter) Write(p []byte) (n int, err error) {
	if bytes.Contains(p, []byte("broken pipe")) || bytes.Contains(p, []byte("connection reset by peer")) {
		return n, nil
	}
	// 主要的writer，用于日志记录
	n, err = t.mainWriter.Write(p)
	if err != nil {
		return n, err
	}
	// 通知ES
	if bytes.Contains(p, []byte("panic")) {
		bugToES(string(p))
	}
	return n, nil
}
