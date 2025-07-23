package config

import "time"

// MemoryConfig 内存消息队列配置
type MemoryConfig struct {
	// MaxQueueSize 每个topic的最大队列大小，0表示无限制
	MaxQueueSize int `json:"max_queue_size" yaml:"max_queue_size"`
	// MaxDelayQueueSize 延时队列的最大大小，0表示无限制
	MaxDelayQueueSize int `json:"max_delay_queue_size" yaml:"max_delay_queue_size"`
	// DelayCheckInterval 延时消息检查间隔
	DelayCheckInterval time.Duration `json:"delay_check_interval" yaml:"delay_check_interval"`
	// EnableMetrics 是否启用指标收集
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`
}

// SetDefaults 设置默认值
func (c *MemoryConfig) SetDefaults() {
	if c.MaxQueueSize == 0 {
		c.MaxQueueSize = 10000 // 默认每个topic最大1万条消息
	}
	if c.MaxDelayQueueSize == 0 {
		c.MaxDelayQueueSize = 1000 // 默认延时队列最大1000条消息
	}
	if c.DelayCheckInterval == 0 {
		c.DelayCheckInterval = 100 * time.Millisecond // 默认100ms检查一次
	}
}
