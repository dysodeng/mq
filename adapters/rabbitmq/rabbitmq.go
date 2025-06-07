package rabbitmq

import (
	"fmt"

	"github.com/dysodeng/mq/config"
	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/observability"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ RabbitMQ消息队列实现
type RabbitMQ struct {
	conn       *amqp.Connection
	producer   *Producer
	consumer   *Consumer
	delayQueue *DelayQueue
	recorder   *observability.MetricsRecorder
	config     config.RabbitMQConfig
	keyPrefix  string
}

// NewRabbitMQ 创建RabbitMQ消息队列
func NewRabbitMQ(config config.RabbitMQConfig, observer observability.Observer, keyPrefix string) (*RabbitMQ, error) {
	if keyPrefix == "" {
		keyPrefix = "mq"
	}

	// 构建连接字符串 - 使用新的配置结构
	var connStr string
	if config.URL != "" {
		connStr = config.URL
	} else {
		connStr = fmt.Sprintf("amqp://%s:%s@%s:%d%s",
			config.Username, config.Password, config.Host, config.Port, config.VHost)
	}

	conn, err := amqp.Dial(connStr)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq connection failed: %w", err)
	}

	// 创建指标记录器
	recorder, err := observability.NewMetricsRecorder(observer, "rabbitmq")
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics recorder: %w", err)
	}

	mq := &RabbitMQ{
		conn:      conn,
		recorder:  recorder,
		config:    config,
		keyPrefix: keyPrefix,
	}

	mq.producer = NewRabbitProducer(conn, observer, keyPrefix)
	mq.consumer = NewRabbitConsumer(conn, observer, keyPrefix)
	mq.delayQueue = NewRabbitDelayQueue(conn, observer, keyPrefix)

	return mq, nil
}

// Producer 获取生产者
func (r *RabbitMQ) Producer() contract.Producer {
	return r.producer
}

// Consumer 获取消费者
func (r *RabbitMQ) Consumer() contract.Consumer {
	return r.consumer
}

// DelayQueue 获取延时队列
func (r *RabbitMQ) DelayQueue() contract.DelayQueue {
	return r.delayQueue
}

// HealthCheck 健康检查
func (r *RabbitMQ) HealthCheck() error {
	if r.conn.IsClosed() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	return nil
}

// Close 关闭连接
func (r *RabbitMQ) Close() error {
	return r.conn.Close()
}
