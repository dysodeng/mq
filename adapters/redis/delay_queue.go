package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/dysodeng/mq/message"
	"github.com/dysodeng/mq/observability"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// DelayQueue Redis延时队列实现
type DelayQueue struct {
	client    redis.Cmdable
	meter     metric.Meter
	logger    *zap.Logger
	keyPrefix string
}

// NewRedisDelayQueue 创建Redis延时队列
func NewRedisDelayQueue(client redis.Cmdable, observer observability.Observer, keyPrefix string) *DelayQueue {
	return &DelayQueue{
		client:    client,
		meter:     observer.GetMeter(),
		logger:    observer.GetLogger(),
		keyPrefix: keyPrefix,
	}
}

// Push 推送延时消息
func (dq *DelayQueue) Push(ctx context.Context, msg *message.Message, delay time.Duration) error {
	executeTime := time.Now().Add(delay).Unix()
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.CreateAt = time.Now()
	msg.Delay = delay

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	delayKey := fmt.Sprintf("%s:delay:queue", dq.keyPrefix)
	msgKey := fmt.Sprintf("%s:delay:msg:%s", dq.keyPrefix, msg.ID)

	// 使用事务确保原子性
	pipe := dq.client.TxPipeline()
	pipe.ZAdd(ctx, delayKey, &redis.Z{
		Score:  float64(executeTime),
		Member: msg.ID,
	})
	pipe.Set(ctx, msgKey, data, delay+time.Hour) // 设置过期时间

	_, err = pipe.Exec(ctx)
	if err != nil {
		if counter, cerr := dq.meter.Int64Counter("delay_queue_push_errors"); cerr == nil {
			counter.Add(ctx, 1)
		}
		return fmt.Errorf("push delay message failed: %w", err)
	}

	if counter, cerr := dq.meter.Int64Counter("delay_queue_push_total"); cerr == nil {
		counter.Add(ctx, 1)
	}
	dq.logger.Info("delay message pushed", zap.String("id", msg.ID), zap.Duration("delay", delay))
	return nil
}

// Pop 弹出到期消息
func (dq *DelayQueue) Pop(ctx context.Context) (*message.Message, error) {
	now := time.Now().Unix()
	delayKey := fmt.Sprintf("%s:delay:queue", dq.keyPrefix)

	// 获取到期的消息ID
	result := dq.client.ZRangeByScore(ctx, delayKey, &redis.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(now, 10),
	})

	msgIDs, err := result.Result()
	if err != nil {
		return nil, fmt.Errorf("get expired messages failed: %w", err)
	}

	if len(msgIDs) == 0 {
		return nil, nil // 没有到期消息
	}

	// 处理第一条消息（保持当前接口不变）
	msgID := msgIDs[0]
	msgKey := fmt.Sprintf("%s:delay:msg:%s", dq.keyPrefix, msgID)

	// 原子性地移除消息并获取内容
	luaScript := `
		local msgKey = KEYS[1]
		local delayKey = KEYS[2]
		local msgID = ARGV[1]
		
		local data = redis.call('GET', msgKey)
		if data then
			redis.call('DEL', msgKey)
			redis.call('ZREM', delayKey, msgID)
			return data
		end
		return nil
	`

	data, err := dq.client.Eval(ctx, luaScript, []string{msgKey, delayKey}, msgID).Result()
	if err != nil {
		return nil, fmt.Errorf("pop delay message failed: %w", err)
	}

	if data == nil {
		return nil, nil
	}

	var msg message.Message
	err = json.Unmarshal([]byte(data.(string)), &msg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal message failed: %w", err)
	}

	if counter, cerr := dq.meter.Int64Counter("delay_queue_pop_total"); cerr == nil {
		counter.Add(ctx, 1)
	}
	return &msg, nil
}

// Remove 移除消息
func (dq *DelayQueue) Remove(ctx context.Context, msgID string) error {
	delayKey := fmt.Sprintf("%s:delay:queue", dq.keyPrefix)
	msgKey := fmt.Sprintf("%s:delay:msg:%s", dq.keyPrefix, msgID)

	pipe := dq.client.TxPipeline()
	pipe.ZRem(ctx, delayKey, msgID)
	pipe.Del(ctx, msgKey)

	_, err := pipe.Exec(ctx)
	return err
}

// Size 获取队列大小
func (dq *DelayQueue) Size(ctx context.Context) (int64, error) {
	delayKey := fmt.Sprintf("%s:delay:queue", dq.keyPrefix)
	return dq.client.ZCard(ctx, delayKey).Result()
}
