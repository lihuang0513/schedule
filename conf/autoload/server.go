package autoload

// ServerConfig 定义项目配置
type ServerConfig struct {
	Host   string `ini:"host"`
	Port   string `ini:"port"`
	LogDir string `ini:"log_dir"`
	Debug  bool   `ini:"debug"`
}

var Server = ServerConfig{
	Host:   "127.0.0.1",
	Port:   "8080",
	LogDir: "./logs/",
	Debug:  true,
}
