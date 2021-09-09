package redis

import (
	"github.com/dysoodeng/mq/consumer"
	"github.com/dysoodeng/mq/driver"
	"github.com/dysoodeng/mq/message"
)

type Redis struct {
	config Config
	key    message.Key
}

type Config struct {
}

func (config *Config) String() string {
	return ""
}

func NewRedis(key message.Key, config driver.Config) (*Redis, error) {
	return nil, nil
}

func (redisDriver Redis) QueuePublish(messageBody string) (message.Message, error) {
	return message.Message{}, nil
}

func (redisDriver Redis) DelayQueuePublish(messageBody string, ttl int64) (message.Message, error) {
	return message.Message{}, nil
}

func (redisDriver Redis) QueueConsume(consumer consumer.Handler) error {
	return nil
}

func (redisDriver Redis) DelayQueueConsume(consumer consumer.Handler) error {
	return nil
}

func (redisDriver Redis) Close() error {
	return nil
}
