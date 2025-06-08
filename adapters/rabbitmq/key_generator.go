package rabbitmq

import "fmt"

// KeyGenerator RabbitMQ键名生成器
type KeyGenerator struct {
	keyPrefix string
}

// NewKeyGenerator 创建键名生成器
func NewKeyGenerator(keyPrefix string) *KeyGenerator {
	return &KeyGenerator{keyPrefix: keyPrefix}
}

// QueueName 生成队列名称
func (kg *KeyGenerator) QueueName(topic string) string {
	return fmt.Sprintf("%s.%s", kg.keyPrefix, topic)
}

// ExchangeName 生成交换机名称
func (kg *KeyGenerator) ExchangeName() string {
	return fmt.Sprintf("%s.exchange", kg.keyPrefix)
}

// DelayQueueName 生成延时队列名称
func (kg *KeyGenerator) DelayQueueName(topic string) string {
	return fmt.Sprintf("%s.%s", kg.keyPrefix, topic)
}

// DelayExchangeName 生成延时交换机名称
func (kg *KeyGenerator) DelayExchangeName() string {
	return fmt.Sprintf("%s.delay.exchange", kg.keyPrefix)
}

// ConsumerTag 生成消费者标签
func (kg *KeyGenerator) ConsumerTag(topic string) string {
	return fmt.Sprintf("%s.consumer.%s", kg.keyPrefix, topic)
}

// RoutingKey 生成路由键
func (kg *KeyGenerator) RoutingKey(topic string) string {
	return topic
}

// DeadLetterQueueName 生成死信队列名称
func (kg *KeyGenerator) DeadLetterQueueName(topic string) string {
	return fmt.Sprintf("%s.dlq.%s", kg.keyPrefix, topic)
}

// DeadLetterExchangeName 生成死信交换机名称
func (kg *KeyGenerator) DeadLetterExchangeName() string {
	return fmt.Sprintf("%s.dlq.exchange", kg.keyPrefix)
}
