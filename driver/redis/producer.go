package redis

import (
	"context"

	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/message"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// redisProducer Redis消息生产者
type redisProducer struct {
	config  contract.Config
	connect *redis.Client
}

// NewProducerConn redis producer connection
func NewProducerConn(config contract.Config) (contract.Producer, error) {
	opts, err := redis.ParseURL(config.String())
	if err != nil {
		return nil, errors.Wrap(err, "Redis connect config error.")
	}
	pc := config.PoolConfig()
	opts.PoolSize = pc.MaxConn
	opts.PoolTimeout = pc.IdleTimeout

	connect := redis.NewClient(opts)

	return &redisProducer{
		config:  config,
		connect: connect,
	}, nil
}

func (producer *redisProducer) Key(queueKey string) string {
	return queueKey + "." + queueKey + "." + queueKey
}

func (producer *redisProducer) QueuePublish(queueKey, messageBody string) (message.Message, error) {
	ctx := context.Background()
	msgID := producer.connect.XAdd(ctx, &redis.XAddArgs{
		Stream: producer.Key(queueKey),
		ID:     "*",
		MaxLen: 0,
		Values: map[string]string{
			"type":    "data",
			"payload": messageBody,
		},
	}).String()
	return message.NewMessage(message.Key{
		ExchangeName: queueKey,
		QueueName:    queueKey,
		RouteKey:     queueKey,
	}, msgID, messageBody), nil
}

func (producer *redisProducer) DelayQueuePublish(queueKey, messageBody string, ttl int64) (message.Message, error) {
	return message.Message{}, nil
}
