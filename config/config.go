package config

import (
	"time"
)

// Config 消息队列配置
type Config struct {
	// Adapter 消息队列适配器类型
	Adapter Adapter `json:"type" yaml:"adapter"`
	// KeyPrefix 全局key前缀
	KeyPrefix string `json:"key_prefix" yaml:"key_prefix"`
	// Redis配置
	Redis RedisConfig `json:"redis" yaml:"redis"`
	// RabbitMQ配置
	RabbitMQ RabbitMQConfig `json:"rabbitmq" yaml:"rabbitmq"`
	// Kafka配置
	Kafka KafkaConfig `json:"kafka" yaml:"kafka"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxIdle     int           `json:"max_idle" yaml:"max_idle"`
	MaxActive   int           `json:"max_active" yaml:"max_active"`
	IdleTimeout time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	MaxLifetime time.Duration `json:"max_lifetime" yaml:"max_lifetime"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	// 消费者配置
	Consumer ConsumerPerformanceConfig `json:"consumer" yaml:"consumer"`
	// 生产者配置
	Producer ProducerPerformanceConfig `json:"producer" yaml:"producer"`
	// 序列化配置
	Serialization SerializationConfig `json:"serialization" yaml:"serialization"`
	// 对象池配置
	ObjectPool ObjectPoolConfig `json:"object_pool" yaml:"object_pool"`
}

// ConsumerPerformanceConfig 消费者性能配置
type ConsumerPerformanceConfig struct {
	WorkerCount   int           `json:"worker_count" yaml:"worker_count"`
	BufferSize    int           `json:"buffer_size" yaml:"buffer_size"`
	BatchSize     int           `json:"batch_size" yaml:"batch_size"`
	PollTimeout   time.Duration `json:"poll_timeout" yaml:"poll_timeout"`
	RetryInterval time.Duration `json:"retry_interval" yaml:"retry_interval"`
	MaxRetries    int           `json:"max_retries" yaml:"max_retries"`
}

// ProducerPerformanceConfig 生产者性能配置
type ProducerPerformanceConfig struct {
	BatchSize     int           `json:"batch_size" yaml:"batch_size"`
	FlushInterval time.Duration `json:"flush_interval" yaml:"flush_interval"`
	Compression   bool          `json:"compression" yaml:"compression"`
}

// SerializationConfig 序列化配置
type SerializationConfig struct {
	Type        string `json:"type" yaml:"type"` // json, msgpack
	Compression bool   `json:"compression" yaml:"compression"`
}

// ObjectPoolConfig 对象池配置
type ObjectPoolConfig struct {
	Enabled           bool `json:"enabled" yaml:"enabled"`
	MaxMessageObjects int  `json:"max_message_objects" yaml:"max_message_objects"`
	MaxBufferObjects  int  `json:"max_buffer_objects" yaml:"max_buffer_objects"`
}

// DefaultPerformanceConfig 默认性能配置
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		Consumer: ConsumerPerformanceConfig{
			WorkerCount:   4,
			BufferSize:    1000,
			BatchSize:     10,
			PollTimeout:   5 * time.Second,
			RetryInterval: time.Second,
			MaxRetries:    3,
		},
		Producer: ProducerPerformanceConfig{
			BatchSize:     100,
			FlushInterval: 100 * time.Millisecond,
			Compression:   false,
		},
		Serialization: SerializationConfig{
			Type:        "json",
			Compression: false,
		},
		ObjectPool: ObjectPoolConfig{
			Enabled:           true,
			MaxMessageObjects: 1000,
			MaxBufferObjects:  100,
		},
	}
}
