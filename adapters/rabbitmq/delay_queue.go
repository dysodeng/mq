package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/dysodeng/mq/message"
	"github.com/dysodeng/mq/observability"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// DelayQueue RabbitMQ延时队列实现 - 使用x-delayed-message插件
type DelayQueue struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	meter     metric.Meter
	logger    *zap.Logger
	keyPrefix string
}

// NewRabbitDelayQueue 创建RabbitMQ延时队列
func NewRabbitDelayQueue(conn *amqp.Connection, observer observability.Observer, keyPrefix string) *DelayQueue {
	ch, err := conn.Channel()
	if err != nil {
		observer.GetLogger().Error("failed to create channel", zap.Error(err))
		return nil
	}

	return &DelayQueue{
		conn:      conn,
		ch:        ch,
		meter:     observer.GetMeter(),
		logger:    observer.GetLogger(),
		keyPrefix: keyPrefix,
	}
}

// Push 推送延时消息 - 使用x-delayed-message插件
func (dq *DelayQueue) Push(ctx context.Context, msg *message.Message, delay time.Duration) error {
	// 使用生产者的SendDelay方法，该方法已更新为使用x-delayed-message插件
	producer := NewRabbitProducer(dq.conn, &observerWrapper{dq.meter, dq.logger}, dq.keyPrefix)
	defer producer.Close()
	return producer.SendDelay(ctx, msg, delay)
}

// observerWrapper 包装器，用于传递meter和logger
type observerWrapper struct {
	meter  metric.Meter
	logger *zap.Logger
}

func (o *observerWrapper) GetMeter() metric.Meter {
	return o.meter
}

func (o *observerWrapper) GetLogger() *zap.Logger {
	return o.logger
}

// Pop 弹出到期消息 (RabbitMQ通过x-delayed-message插件自动处理)
func (dq *DelayQueue) Pop(ctx context.Context) (*message.Message, error) {
	// x-delayed-message插件会自动将到期的延时消息路由到目标队列
	// 消费者直接从目标队列消费即可，无需主动调用此方法
	return nil, fmt.Errorf("rabbitmq delay queue with x-delayed-message uses automatic routing")
}

// Remove 移除消息
func (dq *DelayQueue) Remove(ctx context.Context, msgID string) error {
	// x-delayed-message插件中移除延时消息需要特殊处理
	// 通常在消息发送后无法直接移除，需要在应用层面处理
	return fmt.Errorf("rabbitmq delay queue with x-delayed-message does not support message removal")
}

// Size 获取队列大小
func (dq *DelayQueue) Size(ctx context.Context) (int64, error) {
	// 获取延时队列大小需要通过RabbitMQ管理API
	// x-delayed-message插件的延时消息在交换机内部暂存
	return 0, fmt.Errorf("rabbitmq delay queue size check requires management API")
}

// Close 关闭延时队列
func (dq *DelayQueue) Close() error {
	return dq.ch.Close()
}
