package redis

import "fmt"

// KeyGenerator Redis键名生成器
type KeyGenerator struct {
	keyPrefix string
}

// NewKeyGenerator 创建键名生成器
func NewKeyGenerator(keyPrefix string) *KeyGenerator {
	return &KeyGenerator{keyPrefix: keyPrefix}
}

// QueueKey 生成队列键名
func (kg *KeyGenerator) QueueKey(topic string) string {
	return fmt.Sprintf("%s:queue:%s", kg.keyPrefix, topic)
}

// DelayQueueKey 生成延时队列键名
func (kg *KeyGenerator) DelayQueueKey() string {
	return fmt.Sprintf("%s:delay:queue", kg.keyPrefix)
}

// DelayMessageKey 生成延时消息键名
func (kg *KeyGenerator) DelayMessageKey(msgID string) string {
	return fmt.Sprintf("%s:delay:msg:%s", kg.keyPrefix, msgID)
}
