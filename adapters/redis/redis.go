package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dysodeng/mq/serializer"

	"github.com/dysodeng/mq/config"
	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/observability"
)

// Redis Redis消息队列实现
type Redis struct {
	client        Client
	clientFactory *ClientFactory
	producer      *Producer
	consumer      *Consumer
	delayQueue    *DelayQueue
	recorder      *observability.MetricsRecorder
	keyPrefix     string
	config        config.RedisConfig

	// 延时消息处理
	delayProcessor *DelayProcessor
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	closed         bool
	mu             sync.RWMutex
}

// NewRedisMQ 创建Redis消息队列
func NewRedisMQ(config config.RedisConfig, observer observability.Observer, keyPrefix string) (*Redis, error) {
	if keyPrefix == "" {
		keyPrefix = "mq"
	}

	// 设置默认值
	config.SetDefaults()

	// 创建客户端工厂
	clientFactory := NewClientFactory(config)
	client, err := clientFactory.CreateClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	// 创建上下文
	mainCtx, mainCancel := context.WithCancel(context.Background())
	var success bool
	defer func() {
		if !success {
			mainCancel()
		}
	}()

	// 获取配置
	serializationConfig := config.GetSerializationConfig()
	objectPoolConfig := config.GetObjectPoolConfig()

	// 创建序列化器
	ser, err := serializer.NewSerializer(serializer.Type(serializationConfig.Type))
	if err != nil {
		fmt.Printf("Failed to create serializer, fallback to JSON: %v", err)
		ser, _ = serializer.NewSerializer(serializer.Json)
	}

	// 创建键名生成器
	keys := NewKeyGenerator(keyPrefix)

	// 创建组件（传递序列化和对象池配置）
	producer := NewRedisProducer(client, observer, keyPrefix, config.GetProducerConfig(), ser, keys)
	consumer := NewRedisConsumer(client, observer, keyPrefix, config.GetConsumerConfig(), ser, objectPoolConfig, keys)
	delayQueue := NewRedisDelayQueue(client, observer, keyPrefix, ser, keys)

	// 创建延时处理器
	delayProcessor := NewDelayProcessor(client, observer, keyPrefix, ser, keys)

	// 创建指标记录器
	recorder, err := observability.NewMetricsRecorder(observer, "redis")
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics recorder: %w", err)
	}

	mq := &Redis{
		client:         client,
		clientFactory:  clientFactory,
		producer:       producer,
		consumer:       consumer,
		delayQueue:     delayQueue,
		recorder:       recorder,
		keyPrefix:      keyPrefix,
		config:         config,
		delayProcessor: delayProcessor,
		ctx:            mainCtx,
		cancel:         mainCancel,
	}

	// 启动延时处理器
	mq.wg.Add(1)
	go func() {
		defer mq.wg.Done()
		mq.delayProcessor.Start(mainCtx)
	}()

	success = true
	return mq, nil
}

// Producer 获取生产者
func (r *Redis) Producer() contract.Producer {
	return r.producer
}

// Consumer 获取消费者
func (r *Redis) Consumer() contract.Consumer {
	return r.consumer
}

// DelayQueue 获取延时队列
func (r *Redis) DelayQueue() contract.DelayQueue {
	return r.delayQueue
}

// HealthCheck 健康检查
func (r *Redis) HealthCheck() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return fmt.Errorf("redis client is closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.client.Ping(ctx).Err()
}

// Close 关闭连接
func (r *Redis) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	r.cancel()  // 取消上下文
	r.wg.Wait() // 等待所有goroutine结束

	// 关闭组件
	if r.producer != nil {
		_ = r.producer.Close()
	}
	if r.consumer != nil {
		_ = r.consumer.Close()
	}

	return r.client.Close()
}
