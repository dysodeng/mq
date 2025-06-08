package config

import "time"

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

	// 连接池配置
	PoolSize        int `json:"pool_size" yaml:"pool_size"`
	MinConnections  int `json:"min_connections" yaml:"min_connections"`
	MaxConnections  int `json:"max_connections" yaml:"max_connections"`
	ChannelPoolSize int `json:"channel_pool_size" yaml:"channel_pool_size"`

	// 重连配置
	MaxRetries     int           `json:"max_retries" yaml:"max_retries"`
	RetryInterval  time.Duration `json:"retry_interval" yaml:"retry_interval"`
	ReconnectDelay time.Duration `json:"reconnect_delay" yaml:"reconnect_delay"`

	// 性能配置
	Performance PerformanceConfig `json:"performance" yaml:"performance"`
}

// SetDefaults 设置默认值
func (c *RabbitMQConfig) SetDefaults() {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 5672
	}
	if c.Username == "" {
		c.Username = "guest"
	}
	if c.Password == "" {
		c.Password = "guest"
	}
	if c.VHost == "" {
		c.VHost = "/"
	}
	if c.QoS == 0 {
		c.QoS = 10
	}
	if c.Heartbeat == 0 {
		c.Heartbeat = 60 * time.Second
	}
	if c.ConnectionTimeout == 0 {
		c.ConnectionTimeout = 30 * time.Second
	}

	// 连接池默认值
	if c.PoolSize == 0 {
		c.PoolSize = 10
	}
	if c.MinConnections == 0 {
		c.MinConnections = 2
	}
	if c.MaxConnections == 0 {
		c.MaxConnections = 20
	}
	if c.ChannelPoolSize == 0 {
		c.ChannelPoolSize = 0
	}

	// 重连默认值
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if c.RetryInterval == 0 {
		c.RetryInterval = time.Second
	}
	if c.ReconnectDelay == 0 {
		c.ReconnectDelay = 5 * time.Second
	}

	// 性能配置默认值
	if c.Performance.Consumer.WorkerCount == 0 {
		c.Performance = DefaultPerformanceConfig()
	}
}

// GetProducerConfig 获取生产者配置
func (c *RabbitMQConfig) GetProducerConfig() ProducerPerformanceConfig {
	return c.Performance.Producer
}

// GetConsumerConfig 获取消费者配置
func (c *RabbitMQConfig) GetConsumerConfig() ConsumerPerformanceConfig {
	return c.Performance.Consumer
}

// GetSerializationConfig 获取序列化配置
func (c *RabbitMQConfig) GetSerializationConfig() SerializationConfig {
	return c.Performance.Serialization
}

// GetObjectPoolConfig 获取对象池配置
func (c *RabbitMQConfig) GetObjectPoolConfig() ObjectPoolConfig {
	return c.Performance.ObjectPool
}

// DefaultRabbitMQConfig 默认RabbitMQ配置
func DefaultRabbitMQConfig() RabbitMQConfig {
	config := RabbitMQConfig{
		Host:              "localhost",
		Port:              5672,
		Username:          "guest",
		Password:          "guest",
		VHost:             "/",
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
	config.SetDefaults()
	return config
}
