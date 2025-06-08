package config

import "time"

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
