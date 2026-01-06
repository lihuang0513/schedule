package autoload

// KafkaConfig  定义Kafka配置
type KafkaConfig struct {
	Enable bool   `ini:"enable"`
	Broker string `ini:"broker"`
}

var Kafka = KafkaConfig{
	Enable: false,
	Broker: "127.0.0.1",
}
