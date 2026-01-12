package config

import (
	. "app/conf/autoload"
	"log"
	"sync"
	"time"

	"github.com/go-ini/ini"
)

const (
	ZbbDidSalt    = "#j&MA%j3FSj1D*lB"
	LogChk        = "$safa1231@sd123"
	SignKey       = "ew67oIruFw20iu99roPw" // zhibo8 签名密钥
	DQDSignKey    = "edQlKvOrR%Fcm1o4yB5q" // dongqiudi 签名密钥
	SignExpireSec = 300                    // 签名有效期（秒）
)

// Conf 配置项主结构体
type Conf struct {
	Server        ServerConfig     `ini:"server"`
	Redis         RedisConfig      `ini:"redis"`
	ElasticSearch EsConfig         `ini:"elasticsearch"`
	ScheduleES    ScheduleEsConfig `ini:"schedule-elasticsearch"`
}

// Config 配置文件
var Config = &Conf{
	Server:        Server,
	Redis:         Redis,
	ElasticSearch: Es,
	ScheduleES:    ScheduleEs,
}

// once 只加载一次
var once sync.Once

// 常量定义
const (
	Ios          = "ios"
	Android      = "android"
	MainSplash   = "main_splash"
	MainInter    = "main_inter"
	ZbbDidKey    = "#j&MA%j3FSj1D*lB"
	DqdZbbDidKey = "#dQ8!5bc&Faj%dB7"
	DQD          = "dongqiudi"
)

// 缓存配置常量
const (

	// 相关阈值
	CacheDaysRecent   = 10  // 缓存近10天的完赛赛程阈值
	CacheDaysExtended = 30  // 缓存10-30天的完赛赛程阈值
	CacheDaysExpire   = 180 // 清理完赛赛程数据，60天以上

	// 内存key
	PgameRecommendCacheKey = "pgame:league:recommend" // 全民推荐赛程内存key
	MatchRecordCachePrefix = "pgame:league:finished:" // 完赛内存动态列表缓存key 前缀

	// 请求url
	StaticRecordURL = "http://s.qiumibao.com/json/record/" // 静态数据接口地址

	// Redis版本号key
	PgameLeagueSchedulePrefix  = "pgame:finished:"        // 完赛数据 Redis key 前缀
	PgameLeagueScheduleCodeKey = "pgame:finished:%s:code" // 完赛数据版本号 key，%s 为日期
	PgameLeagueScheduleHashKey = "pgame:finished:%s:hash" // 完赛数据 hash key，%s 为日期
	PgameRecommendRedisKey     = "pgame:recommend"        // 推荐数据 Redis key
	PgameRecommendCodeKey      = "pgame:recommend:code"   // 推荐数据版本号 key

	// 定时任务刷新间隔
	RecentRefreshInterval   = 10 * time.Second // 每10秒
	ExtendedRefreshInterval = 1 * time.Hour    // 每小时
	CleanupInterval         = 24 * time.Hour   // 每天
)

func InitConfig() {
	once.Do(func() {
		// 加载 .ini 配置
		loadIni("./conf/config.ini")
	})
}

// load 加载配置项
func loadIni(configPath string) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		//忽略行内注释
		IgnoreInlineComment: true,
	}, configPath)
	if err != nil {
		//失败
		log.Fatalf("配置文件加载失败：%q", err.Error())
	}
	err = cfg.MapTo(&Config)
	if err != nil {
		//赋值失败
		log.Fatalf("配置文件赋值失败：%q", err.Error())
	}
}
