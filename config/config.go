package config

import (
	"time"
)

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

// RedisConfig Redis配置
type RedisConfig struct {
	// Addr Redis地址
	Addr string `json:"addr" yaml:"addr"`
	// Password Redis密码
	Password string `json:"password" yaml:"password"`
	// DB Redis数据库
	DB int `json:"db" yaml:"db"`
	// PoolSize 连接池大小
	PoolSize int `json:"pool_size" yaml:"pool_size"`
	// MinIdleConns 最小空闲连接数
	MinIdleConns int `json:"min_idle_conns" yaml:"min_idle_conns"`
	// MaxConnAge 连接最大存活时间
	MaxConnAge time.Duration `json:"max_conn_age" yaml:"max_conn_age"`
	// PoolTimeout 连接池超时时间
	PoolTimeout time.Duration `json:"pool_timeout" yaml:"pool_timeout"`
	// IdleTimeout 空闲连接超时时间
	IdleTimeout time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	// IdleCheckFrequency 空闲连接检查频率
	IdleCheckFrequency time.Duration `json:"idle_check_frequency" yaml:"idle_check_frequency"`
	// KeyPrefix Redis key前缀
	KeyPrefix string `json:"key_prefix" yaml:"key_prefix"`
}

// RabbitMQConfig RabbitMQ配置
type RabbitMQConfig struct {
	// URL RabbitMQ连接URL
	URL string `json:"url" yaml:"url"`
	// Host RabbitMQ主机
	Host string `json:"host" yaml:"host"`
	// Port RabbitMQ端口
	Port int `json:"port" yaml:"port"`
	// Username 用户名
	Username string `json:"username" yaml:"username"`
	// Password 密码
	Password string `json:"password" yaml:"password"`
	// VHost 虚拟主机
	VHost string `json:"vhost" yaml:"vhost"`
	// Exchange 默认交换机
	Exchange string `json:"exchange" yaml:"exchange"`
	// ExchangeType 交换机类型
	ExchangeType string `json:"exchange_type" yaml:"exchange_type"`
	// QueueDurable 队列是否持久化
	QueueDurable bool `json:"queue_durable" yaml:"queue_durable"`
	// QueueAutoDelete 队列是否自动删除
	QueueAutoDelete bool `json:"queue_auto_delete" yaml:"queue_auto_delete"`
	// QueueExclusive 队列是否独占
	QueueExclusive bool `json:"queue_exclusive" yaml:"queue_exclusive"`
	// QueueNoWait 队列是否等待
	QueueNoWait bool `json:"queue_no_wait" yaml:"queue_no_wait"`
	// QoS 消费者QoS设置
	QoS int `json:"qos" yaml:"qos"`
	// Heartbeat 心跳间隔
	Heartbeat time.Duration `json:"heartbeat" yaml:"heartbeat"`
	// ConnectionTimeout 连接超时时间
	ConnectionTimeout time.Duration `json:"connection_timeout" yaml:"connection_timeout"`
	// ChannelMax 最大通道数
	ChannelMax int `json:"channel_max" yaml:"channel_max"`
	// FrameSize 帧大小
	FrameSize int `json:"frame_size" yaml:"frame_size"`
	// KeyPrefix 队列/交换机前缀
	KeyPrefix string `json:"key_prefix" yaml:"key_prefix"`
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	// Brokers Kafka代理地址列表
	Brokers []string `json:"brokers" yaml:"brokers"`
	// GroupID 消费者组ID
	GroupID string `json:"group_id" yaml:"group_id"`
	// ClientID 客户端ID
	ClientID string `json:"client_id" yaml:"client_id"`
	// Version Kafka版本
	Version string `json:"version" yaml:"version"`
	// Producer 生产者配置
	Producer KafkaProducerConfig `json:"producer" yaml:"producer"`
	// Consumer 消费者配置
	Consumer KafkaConsumerConfig `json:"consumer" yaml:"consumer"`
	// SASL SASL认证配置
	SASL KafkaSASLConfig `json:"sasl" yaml:"sasl"`
	// TLS TLS配置
	TLS KafkaTLSConfig `json:"tls" yaml:"tls"`
	// KeyPrefix 主题前缀
	KeyPrefix string `json:"key_prefix" yaml:"key_prefix"`
}

// KafkaProducerConfig Kafka生产者配置
type KafkaProducerConfig struct {
	// MaxMessageBytes 最大消息大小
	MaxMessageBytes int `json:"max_message_bytes" yaml:"max_message_bytes"`
	// RequiredAcks 确认级别
	RequiredAcks int `json:"required_acks" yaml:"required_acks"`
	// Timeout 超时时间
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	// Compression 压缩类型
	Compression string `json:"compression" yaml:"compression"`
	// Flush 刷新配置
	Flush KafkaFlushConfig `json:"flush" yaml:"flush"`
	// Retry 重试配置
	Retry KafkaRetryConfig `json:"retry" yaml:"retry"`
	// Idempotent 是否幂等
	Idempotent bool `json:"idempotent" yaml:"idempotent"`
	// BatchSize 批量大小
	BatchSize int `json:"batch_size" yaml:"batch_size"`
	// BatchTimeout 批量超时时间
	BatchTimeout time.Duration `json:"batch_timeout" yaml:"batch_timeout"`
}

// KafkaConsumerConfig Kafka消费者配置
type KafkaConsumerConfig struct {
	// MinBytes 最小字节数
	MinBytes int `json:"min_bytes" yaml:"min_bytes"`
	// MaxBytes 最大字节数
	MaxBytes int `json:"max_bytes" yaml:"max_bytes"`
	// MaxWait 最大等待时间
	MaxWait time.Duration `json:"max_wait" yaml:"max_wait"`
	// CommitInterval 提交间隔
	CommitInterval time.Duration `json:"commit_interval" yaml:"commit_interval"`
	// StartOffset 起始偏移量
	StartOffset int64 `json:"start_offset" yaml:"start_offset"`
	// HeartbeatInterval 心跳间隔
	HeartbeatInterval time.Duration `json:"heartbeat_interval" yaml:"heartbeat_interval"`
	// SessionTimeout 会话超时时间
	SessionTimeout time.Duration `json:"session_timeout" yaml:"session_timeout"`
	// RebalanceTimeout 重平衡超时时间
	RebalanceTimeout time.Duration `json:"rebalance_timeout" yaml:"rebalance_timeout"`
	// RetentionTime 保留时间
	RetentionTime time.Duration `json:"retention_time" yaml:"retention_time"`
}

// KafkaFlushConfig Kafka刷新配置
type KafkaFlushConfig struct {
	// Frequency 刷新频率
	Frequency time.Duration `json:"frequency" yaml:"frequency"`
	// Messages 消息数量
	Messages int `json:"messages" yaml:"messages"`
	// Bytes 字节数
	Bytes int `json:"bytes" yaml:"bytes"`
}

// KafkaRetryConfig Kafka重试配置
type KafkaRetryConfig struct {
	// Max 最大重试次数
	Max int `json:"max" yaml:"max"`
	// Backoff 退避时间
	Backoff time.Duration `json:"backoff" yaml:"backoff"`
}

// KafkaSASLConfig Kafka SASL配置
type KafkaSASLConfig struct {
	// Enable 是否启用SASL
	Enable bool `json:"enable" yaml:"enable"`
	// Mechanism SASL机制
	Mechanism string `json:"mechanism" yaml:"mechanism"`
	// Username 用户名
	Username string `json:"username" yaml:"username"`
	// Password 密码
	Password string `json:"password" yaml:"password"`
}

// KafkaTLSConfig Kafka TLS配置
type KafkaTLSConfig struct {
	// Enable 是否启用TLS
	Enable bool `json:"enable" yaml:"enable"`
	// InsecureSkipVerify 是否跳过证书验证
	InsecureSkipVerify bool `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
	// CertFile 证书文件路径
	CertFile string `json:"cert_file" yaml:"cert_file"`
	// KeyFile 私钥文件路径
	KeyFile string `json:"key_file" yaml:"key_file"`
	// CAFile CA证书文件路径
	CAFile string `json:"ca_file" yaml:"ca_file"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxIdle     int           `json:"max_idle" yaml:"max_idle"`
	MaxActive   int           `json:"max_active" yaml:"max_active"`
	IdleTimeout time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	MaxLifetime time.Duration `json:"max_lifetime" yaml:"max_lifetime"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Adapter:   AdapterRedis,
		KeyPrefix: "mq", // 默认全局前缀
		Redis:     DefaultRedisConfig(),
		RabbitMQ:  DefaultRabbitMQConfig(),
		Kafka:     DefaultKafkaConfig(),
	}
}

// DefaultRedisConfig 默认Redis配置
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:               "localhost:6379",
		Password:           "",
		DB:                 0,
		PoolSize:           10,
		MinIdleConns:       5,
		MaxConnAge:         30 * time.Minute,
		PoolTimeout:        4 * time.Second,
		IdleTimeout:        5 * time.Minute,
		IdleCheckFrequency: 1 * time.Minute,
	}
}

// DefaultRabbitMQConfig 默认RabbitMQ配置
func DefaultRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		Host:              "localhost",
		Port:              5672,
		Username:          "guest",
		Password:          "guest",
		VHost:             "/",
		Exchange:          "mq.direct",
		ExchangeType:      "direct",
		QueueDurable:      true,
		QueueAutoDelete:   false,
		QueueExclusive:    false,
		QueueNoWait:       false,
		QoS:               10,
		Heartbeat:         60 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		ChannelMax:        100,
		FrameSize:         131072,
	}
}

// DefaultKafkaConfig 默认Kafka配置
func DefaultKafkaConfig() KafkaConfig {
	return KafkaConfig{
		Brokers:  []string{"localhost:9092"},
		GroupID:  "mq-consumer-group",
		ClientID: "mq-client",
		Version:  "2.8.0",
		Producer: DefaultKafkaProducerConfig(),
		Consumer: DefaultKafkaConsumerConfig(),
		SASL:     DefaultKafkaSASLConfig(),
		TLS:      DefaultKafkaTLSConfig(),
	}
}

// DefaultKafkaProducerConfig 默认Kafka生产者配置
func DefaultKafkaProducerConfig() KafkaProducerConfig {
	return KafkaProducerConfig{
		MaxMessageBytes: 1000000,
		RequiredAcks:    1,
		Timeout:         10 * time.Second,
		Compression:     "snappy",
		Flush: KafkaFlushConfig{
			Frequency: 100 * time.Millisecond,
			Messages:  100,
			Bytes:     1024 * 1024,
		},
		Retry: KafkaRetryConfig{
			Max:     3,
			Backoff: 100 * time.Millisecond,
		},
		Idempotent:   true,
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
	}
}

// DefaultKafkaConsumerConfig 默认Kafka消费者配置
func DefaultKafkaConsumerConfig() KafkaConsumerConfig {
	return KafkaConsumerConfig{
		MinBytes:          1,
		MaxBytes:          1024 * 1024,
		MaxWait:           500 * time.Millisecond,
		CommitInterval:    1 * time.Second,
		StartOffset:       -1, // latest
		HeartbeatInterval: 3 * time.Second,
		SessionTimeout:    30 * time.Second,
		RebalanceTimeout:  30 * time.Second,
		RetentionTime:     24 * time.Hour,
	}
}

// DefaultKafkaSASLConfig 默认Kafka SASL配置
func DefaultKafkaSASLConfig() KafkaSASLConfig {
	return KafkaSASLConfig{
		Enable:    false,
		Mechanism: "PLAIN",
		Username:  "",
		Password:  "",
	}
}

// DefaultKafkaTLSConfig 默认Kafka TLS配置
func DefaultKafkaTLSConfig() KafkaTLSConfig {
	return KafkaTLSConfig{
		Enable:             false,
		InsecureSkipVerify: false,
		CertFile:           "",
		KeyFile:            "",
		CAFile:             "",
	}
}
