package amqp

import (
	"log"

	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/message"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// amqpConsumer AMQP消息消费者
type amqpConsumer struct {
	config  contract.Config
	key     message.Key
	connect *amqp.Connection
	channel *amqp.Channel
	isDelay bool
}

// NewConsumerConn new consumer connection
func NewConsumerConn(key message.Key, config contract.Config) (contract.Consumer, error) {
	var err error
	var connect *amqp.Connection

	connect, err = amqp.Dial(config.String())
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to RabbitMQ")
	}

	return &amqpConsumer{
		key:     key,
		config:  config,
		connect: connect,
	}, nil
}

// QueueConsume 普通队列消费
func (amqpConsumer *amqpConsumer) QueueConsume(handler contract.Handler) error {
	ch, err := amqpConsumer.connect.Channel()
	if err != nil {
		_ = amqpConsumer.connect.Close()
		return errors.Wrap(err, "Failed to open a channel")
	}

	err = ch.ExchangeDeclare(
		amqpConsumer.key.ExchangeName,
		amqp.ExchangeDirect,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		_ = amqpConsumer.connect.Close()
		_ = ch.Close()
		return errors.Wrap(err, "Failed to declare an exchange")
	}

	queue, err := ch.QueueDeclare(
		amqpConsumer.key.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = amqpConsumer.Close()
		return errors.Wrap(err, "Failed to declare a queue")
	}

	err = ch.QueueBind(
		queue.Name,
		amqpConsumer.key.RouteKey,
		amqpConsumer.key.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to bind a queue")
	}
	defer func() {
		err = ch.QueueUnbind(
			queue.Name,
			amqpConsumer.key.RouteKey,
			amqpConsumer.key.ExchangeName,
			nil,
		)
		log.Printf("%+v", err)
	}()

	msg, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to register a consumer")
	}

	log.Printf(" [*] mq:" + amqpConsumer.key.QueueName + " Waiting for messages.\n")

	for m := range msg {
		consumerError := handler.Handle(message.NewMessage(amqpConsumer.key, m.MessageId, string(m.Body)))
		if consumerError == nil {
			_ = m.Ack(false)
		}
	}

	return nil
}

// DelayQueueConsume 延时队列消费
func (amqpConsumer *amqpConsumer) DelayQueueConsume(handler contract.Handler) error {
	ch, err := amqpConsumer.connect.Channel()
	if err != nil {
		_ = amqpConsumer.connect.Close()
		return errors.Wrap(err, "Failed to open a channel")
	}

	err = ch.ExchangeDeclare(
		amqpConsumer.key.ExchangeName,
		"x-delayed-message",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-delayed-type": "direct",
		},
	)

	if err != nil {
		_ = ch.Close()
		_ = amqpConsumer.connect.Close()
		return errors.Wrap(err, "Failed to declare an exchange")
	}

	queue, err := ch.QueueDeclare(
		amqpConsumer.key.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = amqpConsumer.connect.Close()
		return errors.Wrap(err, "Failed to declare a queue")
	}

	err = ch.QueueBind(
		queue.Name,
		amqpConsumer.key.RouteKey,
		amqpConsumer.key.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = amqpConsumer.connect.Close()
		return errors.Wrap(err, "Failed to bind a queue")
	}

	msg, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = amqpConsumer.connect.Close()
		return errors.Wrap(err, "Failed to register a consumer")
	}

	log.Printf(" [*] mq:" + amqpConsumer.key.QueueName + " Waiting for messages.\n")

	for m := range msg {
		consumerError := handler.Handle(message.NewMessage(amqpConsumer.key, m.MessageId, string(m.Body)))
		if consumerError == nil {
			_ = m.Ack(false)
		}
	}

	return nil
}

// Close 关闭连接
func (amqpConsumer *amqpConsumer) Close() error {
	_ = amqpConsumer.channel.Close()
	_ = amqpConsumer.connect.Close()
	return nil
}
