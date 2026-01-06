package autoload

// RedisConfig 定义Redis配置
type RedisConfig struct {
	Enable   bool   `ini:"enable"`
	Host     string `ini:"host"`
	Port     string `ini:"port"`
	Password string `ini:"password"`
}

var Redis = RedisConfig{
	Enable:   false,
	Host:     "127.0.0.1",
	Password: "123456",
	Port:     "6379",
}
