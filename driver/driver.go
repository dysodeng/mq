package driver

import (
	"github.com/dysodeng/mq/consumer"
	"github.com/dysodeng/mq/message"
)

// Interface 消息队列驱动接口
type Interface interface {
	// QueuePublish 发送普通队列消息
	QueuePublish(messageBody string) (message.Message, error)

	// DelayQueuePublish 发送延时队列消息
	DelayQueuePublish(messageBody string, ttl int64) (message.Message, error)

	// QueueConsume 普通队列消费
	QueueConsume(consumer consumer.Handler) error

	// DelayQueueConsume 延时队列消费
	DelayQueueConsume(consumer consumer.Handler) error

	// Close 关闭连接
	Close() error
}

// Config 配置
type Config interface {
	String() string
}
