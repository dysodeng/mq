package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/dysodeng/mq/message"
	"github.com/dysodeng/mq/observability"
	"go.uber.org/zap"
)

// DelayProcessor 延时消息处理器
type DelayProcessor struct {
	client    Client
	logger    *zap.Logger
	keyPrefix string

	// 性能优化配置
	batchSize    int           // 批处理大小
	pollInterval time.Duration // 轮询间隔
	maxBackoff   time.Duration // 最大退避时间
}

// NewDelayProcessor 创建延时消息处理器
func NewDelayProcessor(client Client, observer observability.Observer, keyPrefix string) *DelayProcessor {
	return &DelayProcessor{
		client:       client,
		logger:       observer.GetLogger(),
		keyPrefix:    keyPrefix,
		batchSize:    100,
		pollInterval: time.Second,
		maxBackoff:   30 * time.Second,
	}
}

// Start 启动延时消息处理
func (dp *DelayProcessor) Start(ctx context.Context) {
	backoff := dp.pollInterval

	for {
		select {
		case <-ctx.Done():
			dp.logger.Info("delay processor stopped")
			return
		case <-time.After(backoff):
			processed, err := dp.processExpiredMessages(ctx)
			if err != nil {
				dp.logger.Error("process expired messages failed", zap.Error(err))
				backoff = dp.calculateBackoff(backoff, true)
			} else if processed == 0 {
				// 没有消息时增加退避时间
				backoff = dp.calculateBackoff(backoff, false)
			} else {
				// 有消息时重置退避时间
				backoff = dp.pollInterval
				dp.logger.Debug("processed expired messages", zap.Int("count", processed))
			}
		}
	}
}

// processExpiredMessages 处理到期消息
func (dp *DelayProcessor) processExpiredMessages(ctx context.Context) (int, error) {
	now := time.Now().Unix()
	delayKey := fmt.Sprintf("%s:delay:queue", dp.keyPrefix)

	// 批量获取到期消息
	result := dp.client.ZRangeByScore(ctx, delayKey, &redis.ZRangeBy{
		Min:   "0",
		Max:   strconv.FormatInt(now, 10),
		Count: int64(dp.batchSize),
	})

	msgIDs, err := result.Result()
	if err != nil {
		return 0, fmt.Errorf("get expired messages failed: %w", err)
	}

	if len(msgIDs) == 0 {
		return 0, nil
	}

	// 批量处理消息
	processed := 0
	for _, msgID := range msgIDs {
		if err := dp.processMessage(ctx, msgID); err != nil {
			dp.logger.Error("process message failed",
				zap.String("msg_id", msgID),
				zap.Error(err))
			continue
		}
		processed++
	}

	return processed, nil
}

// processMessage 处理单个消息
func (dp *DelayProcessor) processMessage(ctx context.Context, msgID string) error {
	msgKey := fmt.Sprintf("%s:delay:msg:%s", dp.keyPrefix, msgID)
	delayKey := fmt.Sprintf("%s:delay:queue", dp.keyPrefix)

	// 使用Lua脚本确保原子性
	luaScript := `
		local msgKey = KEYS[1]
		local delayKey = KEYS[2]
		local msgID = ARGV[1]
		
		-- 获取消息内容
		local msgData = redis.call('GET', msgKey)
		if not msgData then
			return nil
		end
		
		-- 从延时队列中移除
		redis.call('ZREM', delayKey, msgID)
		-- 删除消息数据
		redis.call('DEL', msgKey)
		
		return msgData
	`

	result := dp.client.Eval(ctx, luaScript, []string{msgKey, delayKey}, msgID)
	msgData, err := result.Result()
	if err != nil {
		return fmt.Errorf("eval lua script failed: %w", err)
	}

	if msgData == nil {
		return nil // 消息已被处理
	}

	// 解析消息
	var msg message.Message
	if err := json.Unmarshal([]byte(msgData.(string)), &msg); err != nil {
		return fmt.Errorf("unmarshal message failed: %w", err)
	}

	// 将消息发送到普通队列
	queueKey := fmt.Sprintf("%s:queue:%s", dp.keyPrefix, msg.Topic)
	data, _ := json.Marshal(msg)
	return dp.client.LPush(ctx, queueKey, data).Err()
}

// calculateBackoff 计算退避时间
func (dp *DelayProcessor) calculateBackoff(current time.Duration, hasError bool) time.Duration {
	if hasError {
		// 有错误时使用指数退避
		next := current * 2
		if next > dp.maxBackoff {
			return dp.maxBackoff
		}
		return next
	}

	// 无消息时线性增加
	next := current + dp.pollInterval
	if next > dp.maxBackoff {
		return dp.maxBackoff
	}
	return next
}
