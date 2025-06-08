package config

// Adapter 适配器类型
type Adapter string

// 支持的适配器常量
const (
	AdapterRedis    Adapter = "redis"
	AdapterRabbitMQ Adapter = "rabbitmq"
	AdapterKafka    Adapter = "kafka"
)

// String 实现Stringer接口
func (a Adapter) String() string {
	return string(a)
}

// IsValid 检查适配器是否有效
func (a Adapter) IsValid() bool {
	switch a {
	case AdapterRedis, AdapterRabbitMQ, AdapterKafka:
		return true
	default:
		return false
	}
}
