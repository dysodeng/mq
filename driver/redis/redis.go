package redis

import (
	"github.com/dysoodeng/mq/consumer"
	"github.com/dysoodeng/mq/driver"
	"github.com/dysoodeng/mq/message"
)

type redisDriver struct {
	config Config
	key    message.Key
}

type Config struct {
}

func (config *Config) String() string {
	return ""
}

func New(key message.Key, config driver.Config) (*redisDriver, error) {
	return &redisDriver{}, nil
}

func (redisDriver *redisDriver) QueuePublish(messageBody string) (message.Message, error) {
	return message.Message{}, nil
}

func (redisDriver *redisDriver) DelayQueuePublish(messageBody string, ttl int64) (message.Message, error) {
	return message.Message{}, nil
}

func (redisDriver *redisDriver) QueueConsume(consumer consumer.Handler) error {
	return nil
}

func (redisDriver *redisDriver) DelayQueueConsume(consumer consumer.Handler) error {
	return nil
}

func (redisDriver *redisDriver) Close() error {
	return nil
}
