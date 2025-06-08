package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dysodeng/mq"
	"github.com/dysodeng/mq/config"
	"github.com/dysodeng/mq/contract"
	"github.com/dysodeng/mq/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
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

func initOpenTelemetry(ctx context.Context) (func(), error) {
	// 创建资源
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("mq-service"),
			semconv.ServiceVersion("v1.0.0"),
			attribute.String("env", "dev"),
		),
	)
	if err != nil {
		return nil, err
	}

	mpOpts := []sdkmetric.Option{
		sdkmetric.WithResource(res),
	}

	// 初始化OTLP Metrics导出器
	metricsExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL("http://localhost:4318"),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	// 创建Metrics Provider
	mpOpts = append(mpOpts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricsExporter, sdkmetric.WithInterval(5*time.Second))))
	meterProvider := sdkmetric.NewMeterProvider(mpOpts...)

	otel.SetMeterProvider(meterProvider)

	// 返回清理函数
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = meterProvider.Shutdown(ctx)
	}, nil
}

func main() {
	ctx := context.Background()

	// 1. 初始化OpenTelemetry
	shutdown, err := initOpenTelemetry(ctx)
	if err != nil {
		log.Fatal("Failed to initialize OpenTelemetry:", err)
	}
	defer shutdown()

	// 初始化可观测性
	meter := otel.Meter("mq-performance-example")
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	observer := &MyObserver{
		meter:  meter,
		logger: logger,
	}

	// 高性能Redis配置
	cfg := config.Config{
		Adapter:   config.AdapterRabbitMQ,
		KeyPrefix: "perf:mq",
		Redis: config.RedisConfig{
			Mode:               config.RedisModeStandalone,
			Addr:               "localhost:6379",
			Password:           "123456",
			DB:                 0,
			PoolSize:           200, // 增大连接池
			MinIdleConns:       50,  // 增加最小空闲连接
			MaxConnAge:         time.Hour,
			PoolTimeout:        30 * time.Second,
			IdleTimeout:        5 * time.Minute,
			IdleCheckFrequency: time.Minute,
			MaxRetries:         3,
			MinRetryBackoff:    8 * time.Millisecond,
			MaxRetryBackoff:    512 * time.Millisecond,
			DialTimeout:        5 * time.Second,
			ReadTimeout:        3 * time.Second,
			WriteTimeout:       3 * time.Second,

			// 消费者性能配置
			ConsumerWorkerCount:   20,                     // 消费者工作池大小
			ConsumerBufferSize:    2000,                   // 消费者缓冲区大小
			ConsumerBatchSize:     1,                      // 批处理大小
			ConsumerPollTimeout:   time.Second,            // 轮询超时
			ConsumerRetryInterval: 500 * time.Millisecond, // 重试间隔
			ConsumerMaxRetries:    5,                      // 最大重试次数

			// 生产者性能配置
			ProducerBatchSize:     200,                   // 生产者批处理大小
			ProducerFlushInterval: 50 * time.Millisecond, // 刷新间隔
			ProducerCompression:   true,                  // 启用压缩

			// 序列化配置
			SerializationType:        "msgpack", // 使用MessagePack序列化
			SerializationCompression: true,      // 启用序列化压缩

			// 对象池配置
			ObjectPoolEnabled:           true, // 启用对象池
			ObjectPoolMaxMessageObjects: 2000, // 消息对象池大小
			ObjectPoolMaxBufferObjects:  1000, // 缓冲区对象池大小
		},
		// 高性能RabbitMQ配置
		RabbitMQ: config.RabbitMQConfig{
			// 连接配置
			Host:              "localhost",
			Port:              5672,
			Username:          "guest",
			Password:          "guest",
			VHost:             "/",
			ExchangeType:      "direct",
			QueueDurable:      true,
			QueueAutoDelete:   false,
			QueueExclusive:    false,
			QueueNoWait:       false,
			QoS:               50,               // 增大预取数量
			Heartbeat:         30 * time.Second, // 心跳间隔
			ConnectionTimeout: 10 * time.Second, // 连接超时
			ChannelMax:        200,              // 最大通道数
			FrameSize:         131072,           // 帧大小

			// 连接池配置（高性能）
			PoolSize:        20, // 连接池大小
			MinConnections:  5,  // 最小连接数
			MaxConnections:  50, // 最大连接数
			ChannelPoolSize: 10, // 通道池大小

			// 重连配置
			MaxRetries:     5,                      // 最大重试次数
			RetryInterval:  500 * time.Millisecond, // 重试间隔
			ReconnectDelay: 2 * time.Second,        // 重连延迟

			// 性能配置
			Performance: config.PerformanceConfig{
				// 消费者性能配置
				Consumer: config.ConsumerPerformanceConfig{
					WorkerCount:   20,                     // 消费者工作池大小
					BufferSize:    2000,                   // 消费者缓冲区大小
					BatchSize:     10,                     // 批处理大小
					PollTimeout:   time.Second,            // 轮询超时
					RetryInterval: 500 * time.Millisecond, // 重试间隔
					MaxRetries:    5,                      // 最大重试次数
				},
				// 生产者性能配置
				Producer: config.ProducerPerformanceConfig{
					BatchSize:     200,                   // 生产者批处理大小
					FlushInterval: 50 * time.Millisecond, // 刷新间隔
					Compression:   true,                  // 启用压缩
				},
				// 序列化配置
				Serialization: config.SerializationConfig{
					Type:        "msgpack", // 使用MessagePack序列化
					Compression: true,      // 启用序列化压缩
				},
				// 对象池配置
				ObjectPool: config.ObjectPoolConfig{
					Enabled:           true, // 启用对象池
					MaxMessageObjects: 2000, // 消息对象池大小
					MaxBufferObjects:  1000, // 缓冲区对象池大小
				},
			},
		},
	}

	// 创建消息队列实例（使用工厂模式，自动应用性能配置）
	factory := mq.NewFactory(cfg, mq.WithObserver(observer))
	mqInstance, err := factory.CreateMQ()
	if err != nil {
		log.Fatal("Failed to create MQ:", err)
	}
	defer mqInstance.Close()

	// 获取消费者和生产者（已经是增强版本）
	consumer := mqInstance.Consumer()
	producer := mqInstance.Producer()

	// 高性能消费者示例
	handler := func(ctx context.Context, msg *message.Message) error {
		// 模拟处理时间
		time.Sleep(10 * time.Millisecond)
		fmt.Printf("[%s] Processed: %s\n", time.Now().Format(time.DateTime), string(msg.Payload))
		return nil
	}

	// 创建中间件链
	middlewareChain := contract.NewMiddlewareChain(
		contract.LoggingMiddleware(logger),
		contract.TimeoutMiddleware(30*time.Second),
		contract.RetryMiddleware(3, time.Second),
	)

	handler = middlewareChain.Apply(handler)

	err = consumer.Subscribe(context.Background(), "perf-topic", handler)
	if err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	err = consumer.Subscribe(context.Background(), "perf-delay-topic", handler)
	if err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	// 批量发送消息
	start := time.Now()
	messages := make([]*message.Message, 1000)
	for i := 0; i < 1000; i++ {
		messages[i] = &message.Message{
			Topic:   "perf-topic",
			Payload: []byte(fmt.Sprintf("High performance message %d", i)),
			Headers: map[string]string{
				"batch_id": "batch-001",
				"index":    fmt.Sprintf("%d", i),
			},
		}
	}

	err = producer.SendBatch(ctx, messages)
	if err != nil {
		log.Fatal("Failed to send batch:", err)
	}

	duration := time.Since(start)
	fmt.Printf("Sent 1000 messages in %v (%.2f msg/s)\n", duration, 1000.0/duration.Seconds())

	go func() {
		i := 1
		for {
			msg := &message.Message{
				Topic:   "perf-topic",
				Payload: []byte(fmt.Sprintf("High performance message watch %d", i)),
				Headers: map[string]string{
					"batch_id": "batch-001",
					"index":    fmt.Sprintf("%d", i),
				},
			}
			if err = producer.Send(ctx, msg); err != nil {
				log.Printf("Failed to send batch: %v", err)
			}
			i++
			rand.New(rand.NewSource(rand.Int63()))
			randomNum := rand.Intn(5) + 1
			time.Sleep(time.Duration(randomNum) * time.Second)
		}
	}()
	go func() {
		i := 1
		for {
			msg := &message.Message{
				Topic:   "perf-delay-topic",
				Payload: []byte(fmt.Sprintf("High performance delay message watch %d", i)),
				Headers: map[string]string{
					"batch_id": "batch-002",
					"index":    fmt.Sprintf("%d", i),
				},
			}
			log.Printf("[%s] Sending message: %s", time.Now().Format(time.DateTime), string(msg.Payload))
			if err = producer.SendDelay(ctx, msg, 5*time.Second); err != nil {
				log.Printf("Failed to send batch: %v", err)
			}
			i++
			rand.New(rand.NewSource(rand.Int63()))
			randomNum := rand.Intn(5) + 1
			time.Sleep(time.Duration(randomNum) * time.Second)
		}
	}()

	// 等待消息处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down...")
}
