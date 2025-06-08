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
