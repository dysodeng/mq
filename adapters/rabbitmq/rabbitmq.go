package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dysodeng/mq/config"
	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/observability"
	"github.com/dysodeng/mq/serializer"
	"go.uber.org/zap"
)

// RabbitMQ RabbitMQ消息队列实现
type RabbitMQ struct {
	connectionPool *ConnectionPool
	connFactory    *ConnectionFactory
	producer       *Producer
	consumer       *Consumer
	delayQueue     *DelayQueue
	recorder       *observability.MetricsRecorder
	config         config.RabbitMQConfig
	keyPrefix      string
	serializer     serializer.Serializer

	// 重连机制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed bool
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewRabbitMQ 创建RabbitMQ消息队列
func NewRabbitMQ(cfg config.RabbitMQConfig, observer observability.Observer, keyPrefix string) (*RabbitMQ, error) {
	if keyPrefix == "" {
		keyPrefix = config.DefaultKeyPrefix
	}

	// 设置默认值
	cfg.SetDefaults()

	// 创建指标记录器
	recorder, err := observability.NewMetricsRecorder(observer, config.AdapterRabbitMQ.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics recorder: %w", err)
	}

	// 创建连接工厂
	connFactory := NewConnectionFactory(cfg)
	connectionPool, err := connFactory.CreateConnectionPool(observer, recorder)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 创建序列化器
	serializationConfig := cfg.GetSerializationConfig()
	ser, err := serializer.NewSerializer(serializer.Type(serializationConfig.Type))
	if err != nil {
		observer.GetLogger().Warn("Failed to create serializer, fallback to JSON", zap.Error(err))
		ser, _ = serializer.NewSerializer(serializer.Json)
	}

	// 创建上下文
	mainCtx, mainCancel := context.WithCancel(context.Background())

	// 使用 defer 确保在错误情况下也能清理资源
	var success bool
	defer func() {
		if !success {
			mainCancel()
			_ = connectionPool.Close()
		}
	}()

	// 创建键名生成器
	keyGen := NewKeyGenerator(keyPrefix)

	mq := &RabbitMQ{
		connectionPool: connectionPool,
		connFactory:    connFactory,
		recorder:       recorder,
		config:         cfg,
		keyPrefix:      keyPrefix,
		serializer:     ser,
		ctx:            mainCtx,
		cancel:         mainCancel,
		logger:         observer.GetLogger(),
	}

	// 创建组件
	mq.producer = NewRabbitProducer(connectionPool, observer, recorder, cfg, ser, keyGen)
	mq.consumer = NewRabbitConsumer(connectionPool, observer, recorder, cfg, ser, keyGen)
	mq.delayQueue = NewRabbitDelayQueue(connectionPool, observer, recorder, cfg, ser, keyGen)

	// 启动健康检查和重连机制
	mq.startHealthCheck()

	// 标记成功，避免 defer 中的清理
	success = true
	return mq, nil
}

// startHealthCheck 启动健康检查
func (r *RabbitMQ) startHealthCheck() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.ctx.Done():
				return
			case <-ticker.C:
				if err := r.connectionPool.HealthCheck(context.Background()); err != nil {
					r.logger.Error("RabbitMQ health check failed", zap.Error(err))
					// 这里可以添加重连逻辑
				}
			}
		}
	}()
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
	return r.connectionPool.HealthCheck(context.Background())
}

// Close 关闭连接
func (r *RabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true

	// 取消上下文
	r.cancel()

	// 等待所有goroutine结束
	r.wg.Wait()

	// 关闭连接池
	return r.connectionPool.Close()
}
