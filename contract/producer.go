package contract

import "github.com/dysodeng/mq/message"

// Producer 消息生产者
type Producer interface {
	// QueuePublish 发送普通队列消息
	QueuePublish(messageBody string) (message.Message, error)
	// DelayQueuePublish 发送延时队列消息
	DelayQueuePublish(messageBody string, ttl int64) (message.Message, error)
}
