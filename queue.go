package mq

import (
	"github.com/dysodeng/mq/driver"
	"github.com/dysodeng/mq/driver/amqp"
	"github.com/dysodeng/mq/driver/redis"
	"github.com/dysodeng/mq/message"
	"github.com/pkg/errors"
)

type MessageQueue struct{}

type Driver string

const (
	Amqp  Driver = "amqp"
	Redis Driver = "reds"
)

func NewQueue(driver Driver, key message.Key, config driver.Config) (driver.Interface, error) {
	switch driver {
	case Amqp:
		return amqp.New(key, config)
	case Redis:
		return redis.New(key, config)
	default:
		return nil, errors.New("queue driver not found.")
	}
}
