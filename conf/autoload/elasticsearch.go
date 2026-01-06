package autoload

// EsConfig 定义Redis配置
type EsConfig struct {
	Enable   bool   `ini:"enable"`
	Host     string `ini:"host"`
	Port     string `ini:"port"`
	User     string `ini:"user"`
	Password string `ini:"password"`
}

var Es = EsConfig{
	Enable:   false,
	Host:     "127.0.0.1",
	User:     "root",
	Password: "123456",
	Port:     "6379",
}
