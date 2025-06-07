package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/dysodeng/mq/message"
	"github.com/dysodeng/mq/observability"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// Consumer Redis消费者
type Consumer struct {
	client      *redis.Client
	meter       metric.Meter
	logger      *zap.Logger
	subscribers map[string]*subscription
	mu          sync.RWMutex
	closed      bool
	keyPrefix   string
}

// subscription 订阅信息
type subscription struct {
	topic   string
	handler message.Handler
	cancel  context.CancelFunc
	done    chan struct{}
}

// NewRedisConsumer 创建Redis消费者
func NewRedisConsumer(client *redis.Client, observer observability.Observer, keyPrefix string) *Consumer {
	return &Consumer{
		client:      client,
		meter:       observer.GetMeter(),
		logger:      observer.GetLogger(),
		subscribers: make(map[string]*subscription),
		keyPrefix:   keyPrefix,
	}
}

// Subscribe 订阅消息
func (c *Consumer) Subscribe(ctx context.Context, topic string, handler message.Handler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("consumer is closed")
	}

	if _, exists := c.subscribers[topic]; exists {
		return fmt.Errorf("topic %s already subscribed", topic)
	}

	// 创建订阅上下文
	subCtx, cancel := context.WithCancel(ctx)
	sub := &subscription{
		topic:   topic,
		handler: handler,
		cancel:  cancel,
		done:    make(chan struct{}),
	}
	c.subscribers[topic] = sub

	// 启动消费协程
	go c.consumeLoop(subCtx, sub)

	c.logger.Info("consumer started", zap.String("topic", topic))
	return nil
}

// consumeLoop 消费循环
func (c *Consumer) consumeLoop(ctx context.Context, sub *subscription) {
	defer close(sub.done)

	queueKey := fmt.Sprintf("%s:queue:%s", c.keyPrefix, sub.topic)

	for {
		// 使用BRPop的阻塞特性，当context取消时会自动返回错误
		result := c.client.BRPop(ctx, 0, queueKey) // 0表示无限等待
		if result.Err() != nil {
			if errors.Is(result.Err(), context.Canceled) || errors.Is(result.Err(), context.DeadlineExceeded) {
				c.logger.Info("consumer stopped", zap.String("topic", sub.topic))
				return
			}
			if errors.Is(result.Err(), redis.Nil) {
				continue
			}
			if result.Err().Error() == "redis: client is closed" {
				c.logger.Info("redis client closed, stopping consumer", zap.String("topic", sub.topic))
				return
			}
			c.logger.Error("pop message failed", zap.Error(result.Err()), zap.String("topic", sub.topic))
			continue
		}

		if len(result.Val()) < 2 {
			continue
		}

		msgData := result.Val()[1]
		var msg message.Message
		if err := json.Unmarshal([]byte(msgData), &msg); err != nil {
			c.logger.Error("unmarshal message failed", zap.Error(err), zap.String("topic", sub.topic))
			if counter, cerr := c.meter.Int64Counter("consumer_unmarshal_errors"); cerr == nil {
				counter.Add(ctx, 1)
			}
			continue
		}

		// 处理消息
		if err := sub.handler(ctx, &msg); err != nil {
			c.logger.Error("handle message failed", zap.Error(err), zap.String("topic", sub.topic), zap.String("msgId", msg.ID))
			if counter, cerr := c.meter.Int64Counter("consumer_handle_errors"); cerr == nil {
				counter.Add(ctx, 1)
			}
		} else {
			if counter, cerr := c.meter.Int64Counter("consumer_handle_success"); cerr == nil {
				counter.Add(ctx, 1)
			}
		}
	}
}

// Unsubscribe 取消订阅
func (c *Consumer) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if sub, exists := c.subscribers[topic]; exists {
		sub.cancel()
		<-sub.done // 等待消费协程结束
		delete(c.subscribers, topic)
		c.logger.Info("unsubscribed", zap.String("topic", topic))
	}
	return nil
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true

	// 关闭所有订阅
	for topic, sub := range c.subscribers {
		sub.cancel()
		<-sub.done // 等待消费协程结束
		delete(c.subscribers, topic)
	}

	return nil
}
