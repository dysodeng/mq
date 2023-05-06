package amqp

import (
	"time"

	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/message"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

const (
	defaultMinConn     = 1 // 默认最小连接数
	defaultMaxIdleConn = 1 // 默认最大空闲连接数
	defaultMaxConn     = 1 // 默认最大连接数
	defaultIdleTimeout = 1 * time.Minute
)

// amqpProducer AMQP消息生产者
type amqpProducer struct {
	config contract.Config
	pool   *amqpConnectionPool
}

// NewProducerConn amqp producer connection
func NewProducerConn(config contract.Config) (contract.Producer, error) {
	pc := config.PoolConfig()
	pool, err := newAmqpConnectionPool(poolConfig{
		MinConn:    pc.MinConn,
		MaxConn:    pc.MaxConn,
		MaxIdle:    pc.MaxIdleConn,
		amqpConfig: config,
	})
	if err != nil {
		return nil, err
	}
	return &amqpProducer{
		config: config,
		pool:   pool,
	}, nil
}

func (producer *amqpProducer) QueuePublish(queueKey, messageBody string) (message.Message, error) {
	conn, err := producer.pool.Get()
	if err != nil {
		return message.Message{}, errors.Wrap(err, "")
	}
	defer func() {
		_ = producer.pool.Put(conn)
	}()

	channel, err := conn.Channel()
	if err != nil {
		return message.Message{}, errors.Wrap(err, "Failed to declare a channel")
	}
	defer func() {
		_ = channel.Close()
	}()

	err = channel.ExchangeDeclare(
		queueKey,
		amqp.ExchangeDirect,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return message.Message{}, err
	}

	msg := message.NewMessage(queueKey, "", messageBody)

	err = channel.Publish(
		queueKey,
		queueKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(messageBody),
			MessageId:    msg.Id(),
		},
	)

	if err != nil {
		return message.Message{}, errors.Wrap(err, "Failed to publish a message")
	}

	return msg, nil
}

func (producer *amqpProducer) DelayQueuePublish(queueKey, messageBody string, ttl int64) (message.Message, error) {
	conn, err := producer.pool.Get()
	if err != nil {
		return message.Message{}, errors.Wrap(err, "")
	}
	defer func() {
		_ = producer.pool.Put(conn)
	}()

	channel, err := conn.Channel()
	if err != nil {
		return message.Message{}, errors.Wrap(err, "Failed to declare a channel")
	}
	defer func() {
		_ = channel.Close()
	}()

	err = channel.ExchangeDeclare(
		queueKey,
		"x-delayed-message",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-delayed-type": "direct",
		},
	)

	msg := message.NewMessage(queueKey, "", messageBody)

	err = channel.Publish(
		queueKey,
		queueKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(messageBody),
			MessageId:    msg.Id(),
			Headers: amqp.Table{
				"x-delay": ttl * 1000,
			},
		},
	)

	if err != nil {
		return message.Message{}, errors.Wrap(err, "Failed to publish a message")
	}

	return msg, nil
}
