package redis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/message"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const retryNum int = 3 // 重试次数

// redisConsumer Redis消息消费者
type redisConsumer struct {
	config  contract.Config
	key     message.Key
	connect *redis.Client
	isDelay bool
	ack     map[string]int
	ackLock sync.Mutex
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
		ack:     make(map[string]int),
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
	log.Printf(" [*] mq:" + consumer.key.QueueName + " Waiting for messages.\n")
	ctx := context.Background()

	payloadKey := consumer.Key() + ".payload"
	ackKey := consumer.Key() + ".ack"

	for {
		list, _ := consumer.connect.ZRangeByScore(ctx, consumer.Key(), &redis.ZRangeBy{
			Min:    "0",
			Max:    fmt.Sprintf("%d", time.Now().Unix()),
			Offset: 0,
			Count:  1,
		}).Result()
		if len(list) > 0 {
			messageId := list[0]

			if !consumer.connect.SIsMember(context.Background(), ackKey, messageId).Val() {
				consumer.connect.SAdd(context.Background(), ackKey, messageId)

				// 取出消息内容
				payload, err := consumer.connect.HGet(context.Background(), payloadKey, messageId).Result()
				if err != nil {
					// 消息丢弃
					consumer.connect.ZRem(context.Background(), consumer.Key(), messageId)
					consumer.connect.HDel(context.Background(), payloadKey, messageId)
					consumer.connect.SRem(context.Background(), ackKey, messageId)
					continue
				}

				mark := consumer.Key() + messageId

				// 消息处理
				consumerError := handler.Handle(message.NewMessage(consumer.key, messageId, payload))
				if consumerError == nil {
					consumer.connect.ZRem(context.Background(), consumer.Key(), messageId)
					consumer.connect.HDel(context.Background(), payloadKey, messageId)
					consumer.connect.SRem(context.Background(), ackKey, messageId)
				} else {
					// 消息重试
					count := consumer.retryCount(mark, false)
					if count < retryNum {
						consumer.connect.HMSet(ctx, payloadKey, map[string]interface{}{messageId: payload})
						consumer.connect.ZIncrBy(ctx, consumer.Key(), float64(3*count), messageId)
						consumer.connect.SRem(context.Background(), ackKey, messageId)
						log.Printf(" [*] mq:%s messages %s retry...\n", consumer.key.QueueName, messageId)
					} else {
						consumer.connect.ZRem(context.Background(), consumer.Key(), messageId)
						consumer.connect.HDel(context.Background(), payloadKey, messageId)
						consumer.connect.SRem(context.Background(), ackKey, messageId)
						consumer.retryCount(mark, true)
						log.Printf("[*] mq:%s message %s fail", consumer.key.QueueName, messageId)
					}
				}
			}
		}
		time.Sleep(time.Second)
	}
}

// retryCount 获取重试次数
func (consumer *redisConsumer) retryCount(mark string, isClear bool) int {
	var count int

	consumer.ackLock.Lock()
	defer consumer.ackLock.Unlock()

	if isClear {
		if _, ok := consumer.ack[mark]; ok {
			delete(consumer.ack, mark)
		}
		return 0
	}

	if v, ok := consumer.ack[mark]; ok {
		count = v
		consumer.ack[mark] += 1
	} else {
		consumer.ack[mark] = 1
		count = 0
	}

	return count
}

// Close 关闭连接
func (consumer *redisConsumer) Close() error {
	return consumer.connect.Close()
}
