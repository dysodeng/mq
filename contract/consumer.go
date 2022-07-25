package contract

// Consumer 消息消费者
type Consumer interface {
	// QueueConsume 普通队列消费
	QueueConsume(consumer Handler) error
	// DelayQueueConsume 延时队列消费
	DelayQueueConsume(consumer Handler) error
	// Close 关闭连接
	Close() error
}
