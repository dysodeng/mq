package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/dysodeng/mq/message"
	"github.com/dysodeng/mq/observability"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// Consumer RabbitMQ消费者
type Consumer struct {
	conn        *amqp.Connection
	ch          *amqp.Channel
	meter       metric.Meter
	logger      *zap.Logger
	subscribers map[string]chan bool // 用于取消订阅
	mu          sync.RWMutex
	keyPrefix   string
}

// NewRabbitConsumer 创建RabbitMQ消费者
func NewRabbitConsumer(conn *amqp.Connection, observer observability.Observer, keyPrefix string) *Consumer {
	ch, err := conn.Channel()
	if err != nil {
		observer.GetLogger().Error("failed to create channel", zap.Error(err))
		return nil
	}

	return &Consumer{
		conn:        conn,
		ch:          ch,
		meter:       observer.GetMeter(),
		logger:      observer.GetLogger(),
		subscribers: make(map[string]chan bool),
		keyPrefix:   keyPrefix,
	}
}

// Subscribe 订阅消息
func (c *Consumer) Subscribe(ctx context.Context, topic string, handler message.Handler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	queueName := fmt.Sprintf("%s.%s", c.keyPrefix, topic)

	// 声明队列
	_, err := c.ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// 设置QoS
	err = c.ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// 开始消费
	msgs, err := c.ch.Consume(
		topic, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// 创建取消通道
	stopCh := make(chan bool)
	c.subscribers[topic] = stopCh

	// 启动消费协程
	go func() {
		for {
			select {
			case <-stopCh:
				c.logger.Info("consumer stopped", zap.String("topic", topic))
				return
			case <-ctx.Done():
				c.logger.Info("consumer context cancelled", zap.String("topic", topic))
				return
			case d := <-msgs:
				if err := c.handleMessage(ctx, d, handler); err != nil {
					c.logger.Error("handle message failed", zap.Error(err), zap.String("topic", topic))
					if counter, cerr := c.meter.Int64Counter("consumer_handle_errors"); cerr == nil {
						counter.Add(ctx, 1)
					}
					// 拒绝消息并重新入队
					d.Nack(false, true)
				} else {
					// 确认消息
					d.Ack(false)
					if counter, cerr := c.meter.Int64Counter("consumer_handle_success"); cerr == nil {
						counter.Add(ctx, 1)
					}
				}
			}
		}
	}()

	c.logger.Info("consumer started", zap.String("topic", topic))
	return nil
}

// handleMessage 处理消息
func (c *Consumer) handleMessage(ctx context.Context, delivery amqp.Delivery, handler message.Handler) error {
	var msg message.Message
	err := json.Unmarshal(delivery.Body, &msg)
	if err != nil {
		return fmt.Errorf("unmarshal message failed: %w", err)
	}

	// 转换头部信息
	if msg.Headers == nil {
		msg.Headers = make(map[string]string)
	}
	for k, v := range delivery.Headers {
		if str, ok := v.(string); ok {
			msg.Headers[k] = str
		}
	}

	return handler(ctx, &msg)
}

// Unsubscribe 取消订阅
func (c *Consumer) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if stopCh, exists := c.subscribers[topic]; exists {
		close(stopCh)
		delete(c.subscribers, topic)
		c.logger.Info("unsubscribed", zap.String("topic", topic))
	}
	return nil
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 关闭所有订阅
	for topic, stopCh := range c.subscribers {
		close(stopCh)
		delete(c.subscribers, topic)
	}

	return c.ch.Close()
}
