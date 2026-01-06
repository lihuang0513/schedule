package autoload

// ScheduleEsConfig 赛程ES配置
type ScheduleEsConfig struct {
	Host     string `ini:"host"`
	Port     string `ini:"port"`
	User     string `ini:"user"`
	Password string `ini:"password"`
}

// ScheduleEs 赛程ES默认配置
var ScheduleEs = ScheduleEsConfig{
	Host:     "127.0.0.1",
	Port:     "9200",
	User:     "elastic",
	Password: "",
}

