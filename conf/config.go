package config

import (
	. "app/conf/autoload"
	"log"
	"sync"

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
	CacheDaysRecent        = 10                                   // 近期缓存天数（高频刷新）
	CacheDaysExtended      = 30                                   // 扩展缓存天数（低频刷新）
	StaticRecordURL        = "http://s.qiumibao.com/json/record/" // 静态数据接口地址
	PgameRecommendCacheKey = "pgame:league:recommend"             // 推荐数据内存缓存 key
	MatchRecordCachePrefix = "pgame:league:finished:"             // 完赛内存缓存 key 前缀

	// Redis 版本号 key（用于增量更新判断）
	PgameLeagueSchedulePrefix  = "pgame:league:schedule:"        // 完赛数据 Redis key 前缀
	PgameLeagueScheduleCodeKey = "pgame:league:schedule:%s:code" // 完赛数据版本号 key，%s 为日期
	PgameRecommendRedisKey     = "pgame:recommend"               // 推荐数据 Redis key
	PgameRecommendCodeKey      = "pgame:recommend:code"          // 推荐数据版本号 key
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
