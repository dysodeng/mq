package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dysodeng/mq"
	"github.com/dysodeng/mq/config"
	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type MyObserver struct {
	meter  metric.Meter
	logger *zap.Logger
}

func (o *MyObserver) GetMeter() metric.Meter {
	return o.meter
}

func (o *MyObserver) GetLogger() *zap.Logger {
	return o.logger
}

func main() {
	// 初始化可观测性
	meter := otel.Meter("mq-enhanced-example")
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	observer := &MyObserver{
		meter:  meter,
		logger: logger,
	}

	// 使用Redis集群配置
	cfg := config.Config{
		Adapter:   config.AdapterRedis,
		KeyPrefix: "enhanced:mq",
		Redis:     config.DefaultRedisConfig(),
	}

	// 创建消息队列实例
	factory := mq.NewFactory(cfg, mq.WithObserver(observer))
	mqInstance, err := factory.CreateMQ()
	if err != nil {
		log.Fatal("Failed to create MQ:", err)
	}
	defer mqInstance.Close()

	// 创建中间件链
	middlewareChain := contract.NewMiddlewareChain(
		contract.LoggingMiddleware(logger),
		contract.TimeoutMiddleware(30*time.Second),
		contract.RetryMiddleware(3, time.Second),
	)

	// 消费者示例（带中间件）
	consumer := mqInstance.Consumer()
	handler := func(ctx context.Context, msg *message.Message) error {
		fmt.Printf("[%s] Processing message: %s\n", time.Now().Format(time.DateTime), string(msg.Payload))
		return nil
	}

	// 应用中间件
	enhancedHandler := middlewareChain.Apply(handler)

	err = consumer.Subscribe(context.Background(), "enhanced-topic", enhancedHandler)
	if err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	// 生产者示例
	producer := mqInstance.Producer()
	msg := &message.Message{
		Topic:   "enhanced-topic",
		Payload: []byte("Enhanced message with middleware support"),
		Headers: map[string]string{
			"source":    "enhanced-example",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	ctx := context.Background()
	err = producer.Send(ctx, msg)
	if err != nil {
		log.Fatal("Failed to send message:", err)
	}
	fmt.Println("Enhanced message sent successfully")

	// 等待消息处理
	time.Sleep(10 * time.Second)
}
