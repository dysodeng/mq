package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dysodeng/mq/message"
	"github.com/dysodeng/mq/observability"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// Producer RabbitMQ生产者
type Producer struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	meter     metric.Meter
	logger    *zap.Logger
	keyPrefix string
}

// NewRabbitProducer 创建RabbitMQ生产者
func NewRabbitProducer(conn *amqp.Connection, observer observability.Observer, keyPrefix string) *Producer {
	ch, err := conn.Channel()
	if err != nil {
		observer.GetLogger().Error("failed to create channel", zap.Error(err))
		return nil
	}

	return &Producer{
		conn:      conn,
		ch:        ch,
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

	queueName := fmt.Sprintf("%s.%s", p.keyPrefix, msg.Topic)

	// 声明队列
	_, err := p.ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		if counter, cerr := p.meter.Int64Counter("producer_send_errors"); cerr == nil {
			counter.Add(ctx, 1)
		}
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	body, err := json.Marshal(msg)
	if err != nil {
		if counter, cerr := p.meter.Int64Counter("producer_send_errors"); cerr == nil {
			counter.Add(ctx, 1)
		}
		return fmt.Errorf("marshal message failed: %w", err)
	}

	// 构建AMQP消息头
	headers := make(amqp.Table)
	for k, v := range msg.Headers {
		headers[k] = v
	}

	err = p.ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			MessageId:    msg.ID,
			Timestamp:    msg.CreateAt,
			Headers:      headers,
			DeliveryMode: amqp.Persistent, // 持久化消息
		},
	)

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

// SendDelay 发送延时消息 - 使用x-delayed-message插件
func (p *Producer) SendDelay(ctx context.Context, msg *message.Message, delay time.Duration) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.CreateAt = time.Now()
	msg.Delay = delay

	// 声明x-delayed-message类型的交换机
	delayExchange := fmt.Sprintf("%s.%s.delayed", p.keyPrefix, msg.Topic)
	queueName := fmt.Sprintf("%s.%s", p.keyPrefix, msg.Topic)

	// 声明延时交换机 - 使用x-delayed-message插件
	err := p.ch.ExchangeDeclare(
		delayExchange,       // name
		"x-delayed-message", // type - 使用x-delayed-message插件
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		amqp.Table{
			"x-delayed-type": "direct", // 指定内部交换机类型
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare delayed exchange: %w", err)
	}

	// 声明目标队列
	_, err = p.ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare target queue: %w", err)
	}

	// 绑定目标队列到延时交换机
	err = p.ch.QueueBind(
		queueName,     // queue name
		msg.Topic,     // routing key
		delayExchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind target queue: %w", err)
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	// 准备消息头，包含延时信息
	headers := make(amqp.Table)
	for k, v := range msg.Headers {
		headers[k] = v
	}
	// 设置延时时间（毫秒）
	headers["x-delay"] = int64(delay / time.Millisecond)

	// 发送延时消息到x-delayed-message交换机
	err = p.ch.Publish(
		delayExchange, // exchange
		msg.Topic,     // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			MessageId:    msg.ID,
			Timestamp:    msg.CreateAt,
			Headers:      headers,
			DeliveryMode: amqp.Persistent, // 持久化消息
		},
	)

	if err != nil {
		if counter, cerr := p.meter.Int64Counter("producer_delay_send_errors"); cerr == nil {
			counter.Add(ctx, 1)
		}
		p.logger.Error("send delay message failed", zap.Error(err), zap.String("topic", msg.Topic), zap.Duration("delay", delay))
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
	for _, msg := range messages {
		if err := p.Send(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return p.ch.Close()
}
