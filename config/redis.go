package config

import "time"

// RedisMode Redis连接模式
type RedisMode string

const (
	RedisModeStandalone RedisMode = "standalone"
	RedisModeCluster    RedisMode = "cluster"
	RedisModeSentinel   RedisMode = "sentinel"
)

// RedisConfig Redis配置
type RedisConfig struct {
	// 连接模式
	Mode RedisMode `json:"mode" yaml:"mode"`

	// 单机模式配置
	Addr     string `json:"addr" yaml:"addr"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`

	// 集群模式配置
	Addrs []string `json:"addrs" yaml:"addrs"`

	// 哨兵模式配置
	SentinelAddrs    []string `json:"sentinel_addrs" yaml:"sentinel_addrs"`
	SentinelPassword string   `json:"sentinel_password" yaml:"sentinel_password"`
	MasterName       string   `json:"master_name" yaml:"master_name"`

	// 连接池配置
	PoolSize           int           `json:"pool_size" yaml:"pool_size"`
	MinIdleConns       int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	MaxConnAge         time.Duration `json:"max_conn_age" yaml:"max_conn_age"`
	PoolTimeout        time.Duration `json:"pool_timeout" yaml:"pool_timeout"`
	IdleTimeout        time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	IdleCheckFrequency time.Duration `json:"idle_check_frequency" yaml:"idle_check_frequency"`

	// 重试和超时配置
	MaxRetries      int           `json:"max_retries" yaml:"max_retries"`
	MinRetryBackoff time.Duration `json:"min_retry_backoff" yaml:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `json:"max_retry_backoff" yaml:"max_retry_backoff"`
	DialTimeout     time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout     time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout" yaml:"write_timeout"`

	// 其他配置
	KeyPrefix string `json:"key_prefix" yaml:"key_prefix"`

	// 消费者性能配置
	ConsumerWorkerCount   int           `json:"consumer_worker_count" yaml:"consumer_worker_count"`
	ConsumerBufferSize    int           `json:"consumer_buffer_size" yaml:"consumer_buffer_size"`
	ConsumerBatchSize     int           `json:"consumer_batch_size" yaml:"consumer_batch_size"`
	ConsumerPollTimeout   time.Duration `json:"consumer_poll_timeout" yaml:"consumer_poll_timeout"`
	ConsumerRetryInterval time.Duration `json:"consumer_retry_interval" yaml:"consumer_retry_interval"`
	ConsumerMaxRetries    int           `json:"consumer_max_retries" yaml:"consumer_max_retries"`

	// 生产者性能配置
	ProducerBatchSize     int           `json:"producer_batch_size" yaml:"producer_batch_size"`
	ProducerFlushInterval time.Duration `json:"producer_flush_interval" yaml:"producer_flush_interval"`
	ProducerCompression   bool          `json:"producer_compression" yaml:"producer_compression"`

	// 序列化配置
	SerializationType        string `json:"serialization_type" yaml:"serialization_type"`
	SerializationCompression bool   `json:"serialization_compression" yaml:"serialization_compression"`

	// 对象池配置
	ObjectPoolEnabled           bool `json:"object_pool_enabled" yaml:"object_pool_enabled"`
	ObjectPoolMaxMessageObjects int  `json:"object_pool_max_message_objects" yaml:"object_pool_max_message_objects"`
	ObjectPoolMaxBufferObjects  int  `json:"object_pool_max_buffer_objects" yaml:"object_pool_max_buffer_objects"`
}

// GetConsumerConfig 获取消费者配置
func (r *RedisConfig) GetConsumerConfig() ConsumerPerformanceConfig {
	return ConsumerPerformanceConfig{
		WorkerCount:   r.ConsumerWorkerCount,
		BufferSize:    r.ConsumerBufferSize,
		BatchSize:     r.ConsumerBatchSize,
		PollTimeout:   r.ConsumerPollTimeout,
		RetryInterval: r.ConsumerRetryInterval,
		MaxRetries:    r.ConsumerMaxRetries,
	}
}

// GetProducerConfig 获取生产者配置
func (r *RedisConfig) GetProducerConfig() ProducerPerformanceConfig {
	return ProducerPerformanceConfig{
		BatchSize:     r.ProducerBatchSize,
		FlushInterval: r.ProducerFlushInterval,
		Compression:   r.ProducerCompression,
	}
}

// GetSerializationConfig 获取序列化配置
func (r *RedisConfig) GetSerializationConfig() SerializationConfig {
	return SerializationConfig{
		Type:        r.SerializationType,
		Compression: r.SerializationCompression,
	}
}

// GetObjectPoolConfig 获取对象池配置
func (r *RedisConfig) GetObjectPoolConfig() ObjectPoolConfig {
	return ObjectPoolConfig{
		Enabled:           r.ObjectPoolEnabled,
		MaxMessageObjects: r.ObjectPoolMaxMessageObjects,
		MaxBufferObjects:  r.ObjectPoolMaxBufferObjects,
	}
}

// SetDefaults 设置默认值
func (r *RedisConfig) SetDefaults() {
	if r.ConsumerWorkerCount == 0 {
		r.ConsumerWorkerCount = 4
	}
	if r.ConsumerBufferSize == 0 {
		r.ConsumerBufferSize = 1000
	}
	if r.ConsumerBatchSize == 0 {
		r.ConsumerBatchSize = 10
	}
	if r.ConsumerPollTimeout == 0 {
		r.ConsumerPollTimeout = 5 * time.Second
	}
	if r.ConsumerRetryInterval == 0 {
		r.ConsumerRetryInterval = time.Second
	}
	if r.ConsumerMaxRetries == 0 {
		r.ConsumerMaxRetries = 3
	}

	if r.ProducerBatchSize == 0 {
		r.ProducerBatchSize = 100
	}
	if r.ProducerFlushInterval == 0 {
		r.ProducerFlushInterval = 100 * time.Millisecond
	}

	if r.SerializationType == "" {
		r.SerializationType = "json"
	}

	if r.ObjectPoolMaxMessageObjects == 0 {
		r.ObjectPoolMaxMessageObjects = 1000
	}
	if r.ObjectPoolMaxBufferObjects == 0 {
		r.ObjectPoolMaxBufferObjects = 100
	}
	// 默认启用对象池
	r.ObjectPoolEnabled = true
}

// DefaultRedisConfig 默认Redis配置
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Mode:               RedisModeStandalone,
		Addr:               "localhost:6379",
		Password:           "123456",
		DB:                 0,
		PoolSize:           100,
		MinIdleConns:       10,
		MaxConnAge:         time.Hour,
		PoolTimeout:        30 * time.Second,
		IdleTimeout:        5 * time.Minute,
		IdleCheckFrequency: time.Minute,
		MaxRetries:         3,
		MinRetryBackoff:    8 * time.Millisecond,
		MaxRetryBackoff:    512 * time.Millisecond,
		DialTimeout:        5 * time.Second,
		ReadTimeout:        3 * time.Second,
		WriteTimeout:       3 * time.Second,
	}
}

// DefaultRedisClusterConfig 默认Redis集群配置
func DefaultRedisClusterConfig() RedisConfig {
	config := DefaultRedisConfig()
	config.Mode = RedisModeCluster
	config.Addrs = []string{"localhost:7000", "localhost:7001", "localhost:7002"}
	return config
}

// DefaultRedisSentinelConfig 默认Redis哨兵配置
func DefaultRedisSentinelConfig() RedisConfig {
	config := DefaultRedisConfig()
	config.Mode = RedisModeSentinel
	config.SentinelAddrs = []string{"localhost:26379", "localhost:26380", "localhost:26381"}
	config.MasterName = "mymaster"
	return config
}
