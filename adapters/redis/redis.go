package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dysodeng/mq/config"
	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/observability"
	"github.com/go-redis/redis/v8"
)

// Redis Redis消息队列实现
type Redis struct {
	client     *redis.Client
	producer   *Producer
	consumer   *Consumer
	delayQueue *DelayQueue
	recorder   *observability.MetricsRecorder
	keyPrefix  string
}

// NewRedisMQ 创建Redis消息队列
// 在NewRedisMQ函数中启动延时消息处理器
func NewRedisMQ(config config.RedisConfig, observer observability.Observer, keyPrefix string) (*Redis, error) {
	if keyPrefix == "" {
		keyPrefix = "mq"
	}

	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr, // 改为单个地址
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,     // 直接使用字段
		MinIdleConns: config.MinIdleConns, // 直接使用字段
		IdleTimeout:  config.IdleTimeout,  // 直接使用字段
		MaxConnAge:   config.MaxConnAge,   // 直接使用字段
		PoolTimeout:  config.PoolTimeout,  // 新增字段
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	// 创建指标记录器
	recorder, err := observability.NewMetricsRecorder(observer, "redis")
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics recorder: %w", err)
	}

	mq := &Redis{
		client:    client,
		recorder:  recorder,
		keyPrefix: keyPrefix,
	}

	mq.producer = NewRedisProducer(client, observer, keyPrefix)
	mq.consumer = NewRedisConsumer(client, observer, keyPrefix)
	mq.delayQueue = NewRedisDelayQueue(client, observer, keyPrefix)

	// 启动延时消息处理器
	go mq.processDelayMessages(context.Background())

	return mq, nil
}

// processDelayMessages 处理延时消息
func (r *Redis) processDelayMessages(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second) // 每秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 从延时队列中弹出到期消息
			msg, err := r.delayQueue.Pop(ctx)
			if err != nil || msg == nil {
				continue
			}

			// 将到期消息发送到普通队列
			queueKey := fmt.Sprintf("%s:queue:%s", r.keyPrefix, msg.Topic)
			data, _ := json.Marshal(msg)
			r.client.LPush(ctx, queueKey, data)
		}
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return r.client.Ping(ctx).Err()
}

// Close 关闭连接
func (r *Redis) Close() error {
	return r.client.Close()
}
