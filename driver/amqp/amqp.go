package amqp

import (
	"fmt"
	"log"

	"github.com/dysoodeng/mq/consumer"
	"github.com/dysoodeng/mq/driver"
	"github.com/dysoodeng/mq/message"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// AMQP amqp驱动
type AMQP struct {
	config  driver.Config
	key     message.Key
	connect *amqp.Connection
	channel *amqp.Channel
}

// Config 配置
type Config struct {
	Host     string
	Username string
	Password string
	VHost    string
}

func (config *Config) String() string {
	if config.VHost == "" {
		config.VHost = "/"
	}
	return fmt.Sprintf(
		"amqp://%s:%s@%s%s",
		config.Username,
		config.Password,
		config.Host,
		config.VHost,
	)
}

// NewAMQP new amqp driver
func NewAMQP(key message.Key, config driver.Config) (*AMQP, error) {
	connect, err := amqp.Dial(config.String())
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to RabbitMQ")
	}

	ch, err := connect.Channel()
	if err != nil {
		_ = connect.Close()
		return nil, errors.Wrap(err, "Failed to open a channel")
	}

	err = ch.ExchangeDeclare(
		key.ExchangeName,
		amqp.ExchangeDirect,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		_ = connect.Close()
		_ = ch.Close()
		return nil, errors.Wrap(err, "Failed to declare an exchange")
	}

	return &AMQP{
		key:     key,
		config:  config,
		connect: connect,
		channel: ch,
	}, nil
}

func (amqpDriver *AMQP) QueuePublish(messageBody string) (message.Message, error) {
	msg := message.NewMessage(amqpDriver.key, "", messageBody)

	err := amqpDriver.channel.Publish(
		amqpDriver.key.ExchangeName,
		amqpDriver.key.RouteKey,
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

func (amqpDriver *AMQP) DelayQueuePublish(messageBody string, ttl int64) (message.Message, error) {
	return message.NewMessage(amqpDriver.key, "", messageBody), nil
}

func (amqpDriver *AMQP) QueueConsume(consumer consumer.Handler) error {
	queue, err := amqpDriver.channel.QueueDeclare(
		amqpDriver.key.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = amqpDriver.Close()
		return errors.Wrap(err, "Failed to declare a queue")
	}

	err = amqpDriver.channel.QueueBind(
		queue.Name,
		amqpDriver.key.RouteKey,
		amqpDriver.key.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		_ = amqpDriver.Close()
		return errors.Wrap(err, "Failed to bind a queue")
	}

	msg, err := amqpDriver.channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = amqpDriver.Close()
		return errors.Wrap(err, "Failed to register a consumer")
	}

	log.Printf(" [*] mq:" + amqpDriver.key.QueueName + " Waiting for messages.\n")

	for m := range msg {
		consumerError := consumer.Handle(message.NewMessage(amqpDriver.key, m.MessageId, string(m.Body)))
		if consumerError == nil {
			_ = m.Ack(false)
		}
	}

	return nil
}

func (amqpDriver *AMQP) DelayQueueConsume(consumer consumer.Handler) error {
	return nil
}

func (amqpDriver *AMQP) Close() error {
	_ = amqpDriver.channel.Close()
	_ = amqpDriver.connect.Close()
	return nil
}
