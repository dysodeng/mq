package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dysodeng/mq/message"
	"github.com/dysodeng/mq/observability"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// Producer Redis生产者
type Producer struct {
	client    redis.Cmdable
	meter     metric.Meter
	logger    *zap.Logger
	keyPrefix string
}

// NewRedisProducer 创建Redis生产者
func NewRedisProducer(client redis.Cmdable, observer observability.Observer, keyPrefix string) *Producer {
	return &Producer{
		client:    client,
		meter:     observer.GetMeter(),
		logger:    observer.GetLogger(),
		keyPrefix: keyPrefix,
	}
}

// Send 发送消息
func (p *Producer) Send(ctx context.Context, msg *message.Message) error {
	start := time.Now()
	defer func() {
		if duration, err := p.meter.Float64Histogram("producer_send_duration"); err == nil {
			duration.Record(ctx, time.Since(start).Seconds())
		}
	}()

	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.CreateAt = time.Now()

	data, err := json.Marshal(msg)
	if err != nil {
		if counter, cerr := p.meter.Int64Counter("producer_send_errors"); cerr == nil {
			counter.Add(ctx, 1)
		}
		return fmt.Errorf("marshal message failed: %w", err)
	}

	queueKey := fmt.Sprintf("%s:queue:%s", p.keyPrefix, msg.Topic)
	err = p.client.LPush(ctx, queueKey, data).Err()
	if err != nil {
		if counter, cerr := p.meter.Int64Counter("producer_send_errors"); cerr == nil {
			counter.Add(ctx, 1)
		}
		p.logger.Error("send message failed", zap.Error(err), zap.String("topic", msg.Topic))
		return fmt.Errorf("send message failed: %w", err)
	}

	if counter, cerr := p.meter.Int64Counter("producer_send_total"); cerr == nil {
		counter.Add(ctx, 1)
	}
	p.logger.Info("message sent", zap.String("id", msg.ID), zap.String("topic", msg.Topic))
	return nil
}

// SendDelay 发送延时消息
func (p *Producer) SendDelay(ctx context.Context, msg *message.Message, delay time.Duration) error {
	start := time.Now()
	defer func() {
		if duration, err := p.meter.Float64Histogram("producer_delay_send_duration"); err == nil {
			duration.Record(ctx, time.Since(start).Seconds())
		}
	}()

	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.CreateAt = time.Now()
	msg.Delay = delay

	// 直接使用延时队列发送，而不是普通队列
	executeTime := time.Now().Add(delay).Unix()

	data, err := json.Marshal(msg)
	if err != nil {
		if counter, cerr := p.meter.Int64Counter("producer_delay_send_errors"); cerr == nil {
			counter.Add(ctx, 1)
		}
		return fmt.Errorf("marshal message failed: %w", err)
	}

	delayKey := fmt.Sprintf("%s:delay:queue", p.keyPrefix)
	msgKey := fmt.Sprintf("%s:delay:msg:%s", p.keyPrefix, msg.ID)

	// 使用事务确保原子性
	pipe := p.client.TxPipeline()
	pipe.ZAdd(ctx, delayKey, &redis.Z{
		Score:  float64(executeTime),
		Member: msg.ID,
	})
	pipe.Set(ctx, msgKey, data, delay+time.Hour) // 设置过期时间

	_, err = pipe.Exec(ctx)
	if err != nil {
		if counter, cerr := p.meter.Int64Counter("producer_delay_send_errors"); cerr == nil {
			counter.Add(ctx, 1)
		}
		p.logger.Error("send delay message failed", zap.Error(err), zap.String("topic", msg.Topic))
		return fmt.Errorf("send delay message failed: %w", err)
	}

	if counter, cerr := p.meter.Int64Counter("producer_delay_send_total"); cerr == nil {
		counter.Add(ctx, 1)
	}
	p.logger.Info("delay message sent", zap.String("id", msg.ID), zap.String("topic", msg.Topic), zap.Duration("delay", delay))
	return nil
}

// SendBatch 批量发送消息
func (p *Producer) SendBatch(ctx context.Context, messages []*message.Message) error {
	pipe := p.client.Pipeline()

	for _, msg := range messages {
		if msg.ID == "" {
			msg.ID = uuid.New().String()
		}
		msg.CreateAt = time.Now()

		data, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("marshal message failed: %w", err)
		}

		queueKey := fmt.Sprintf("%s:queue:%s", p.keyPrefix, msg.Topic)
		pipe.LPush(ctx, queueKey, data)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		if counter, cerr := p.meter.Int64Counter("producer_batch_send_errors"); cerr == nil {
			counter.Add(ctx, 1)
		}
		return fmt.Errorf("batch send failed: %w", err)
	}

	if counter, cerr := p.meter.Int64Counter("producer_send_total"); cerr == nil {
		counter.Add(ctx, int64(len(messages)))
	}
	return nil
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return nil
}
