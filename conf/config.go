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
