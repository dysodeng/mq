package redis

import (
	"context"
	"log"
	"strings"

	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/message"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// redisConsumer Redis消息消费者
type redisConsumer struct {
	config  contract.Config
	key     message.Key
	connect *redis.Client
	isDelay bool
}

// NewConsumerConn new consumer connection
func NewConsumerConn(key message.Key, config contract.Config) (contract.Consumer, error) {
	opts, err := redis.ParseURL(config.String())
	if err != nil {
		return nil, errors.Wrap(err, "Redis connect config error.")
	}

	connect := redis.NewClient(opts)

	return &redisConsumer{
		key:     key,
		config:  config,
		connect: connect,
	}, nil
}

func (consumer *redisConsumer) Key() string {
	return consumer.key.ExchangeName + "." + consumer.key.QueueName + "." + consumer.key.RouteKey
}

// QueueConsume 普通队列消费
func (consumer *redisConsumer) QueueConsume(handler contract.Handler) error {
	ctx := context.Background()
	consumer.connect.XAdd(ctx, &redis.XAddArgs{
		Stream: consumer.Key(),
		ID:     "*",
		MaxLen: 0,
		Values: map[string]string{
			"type":    "init",
			"payload": "init",
		},
	})
	consumer.connect.XGroupCreate(ctx, consumer.Key(), consumer.key.QueueName, "0")

	log.Printf(" [*] mq:" + consumer.key.QueueName + " Waiting for messages.\n")

	uuidItem, _ := uuid.NewUUID()
	queueConsume := strings.Replace(uuidItem.String(), "-", "", -1)

	for {
		res, err := consumer.connect.XReadGroup(context.Background(), &redis.XReadGroupArgs{
			Group:    consumer.key.QueueName,
			Consumer: consumer.key.QueueName + queueConsume,
			Streams:  []string{consumer.Key(), ">"},
			Count:    1,
			Block:    0,
			NoAck:    false,
		}).Result()
		if err != nil {
			return err
		}

		for _, stream := range res {
			m := stream.Messages[0]
			if t, ok := m.Values["type"]; ok {
				if t == "init" {
					consumer.connect.XAck(context.Background(), consumer.Key(), consumer.key.QueueName, m.ID)
					break
				}
			}
			var payload string
			if p, ok := m.Values["payload"]; ok {
				if pp, ok := p.(string); ok {
					payload = pp
				}
			}
			consumerError := handler.Handle(message.NewMessage(consumer.key, m.ID, payload))
			if consumerError == nil {
				consumer.connect.XAck(context.Background(), consumer.Key(), consumer.key.QueueName, m.ID)
			}
		}
	}
}

// DelayQueueConsume 延时队列消费
func (consumer *redisConsumer) DelayQueueConsume(handler contract.Handler) error {
	return nil
}

// Close 关闭连接
func (consumer *redisConsumer) Close() error {
	return consumer.connect.Close()
}
