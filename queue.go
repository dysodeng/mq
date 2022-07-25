package mq

import (
	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/driver/amqp"
	"github.com/dysodeng/mq/message"
	"github.com/pkg/errors"
)

type MessageQueue struct{}

type Driver string

const (
	Amqp Driver = "amqp"
)

// NewQueueConsumer 队列消费者
func NewQueueConsumer(driver Driver, queueKey string, config contract.Config) (contract.Consumer, error) {
	switch driver {
	case Amqp:
		return amqp.NewConsumerConn(message.Key{
			ExchangeName: queueKey,
			QueueName:    queueKey,
			RouteKey:     queueKey,
		}, config)
	default:
		return nil, errors.New("queue driver not found.")
	}
}

// NewQueueProducer 队列生产者
func NewQueueProducer(driver Driver, queueKey string, config contract.Config) (contract.Producer, error) {
	switch driver {
	case Amqp:
		return amqp.NewProducerConn(message.Key{
			ExchangeName: queueKey,
			QueueName:    queueKey,
			RouteKey:     queueKey,
		}, config)
	default:
		return nil, errors.New("queue driver not found.")
	}
}
