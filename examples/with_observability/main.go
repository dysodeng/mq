package main

import (
	"context"
	"log"
	"time"

	"github.com/dysodeng/mq"
	"github.com/dysodeng/mq/config"
	"github.com/dysodeng/mq/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
)

// MyObserver 实现Observer接口
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
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("mq-service"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
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
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricsExporter,
			sdkmetric.WithInterval(30*time.Second))),
	)
	otel.SetMeterProvider(meterProvider)

	// 初始化OTLP Trace导出器
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL("http://localhost:4318"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// 创建Trace Provider
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// 返回清理函数
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		meterProvider.Shutdown(ctx)
		traceProvider.Shutdown(ctx)
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

	// 2. 初始化zap日志
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// 3. 获取Meter
	meter := otel.Meter("mq-example")

	// 4. 创建Observer实例
	observer := &MyObserver{
		meter:  meter,
		logger: logger,
	}

	// 5. 创建配置
	cfg := config.Config{
		Adapter:   config.AdapterRedis,
		KeyPrefix: "app:mq",
		Redis:     config.DefaultRedisConfig(),
	}

	// 6. 创建MQ工厂
	factory := mq.NewFactory(cfg, mq.WithObserver(observer))

	// 7. 创建Redis MQ实例
	mqInstance, err := factory.CreateMQ()
	if err != nil {
		log.Fatal("Failed to create MQ instance:", err)
	}
	defer mqInstance.Close()

	// 8. 发送消息
	producer := mqInstance.Producer()
	go func() {
		msg := &message.Message{
			Topic:   "test-topic",
			Payload: []byte("Hello, World---1!"),
		}
		err = producer.Send(ctx, msg)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
		}

		err = producer.Send(ctx, &message.Message{
			Topic:   "test-topic",
			Payload: []byte("Hello, World---2!"),
		})

		go func() {
			time.Sleep(5 * time.Second)
			err = producer.Send(ctx, &message.Message{
				Topic:   "test-topic",
				Payload: []byte("Hello, World---3!"),
			})
		}()

		err = producer.SendDelay(ctx, &message.Message{
			Topic:   "delay-test-topic",
			Payload: []byte("Hello, Delay Message!"),
		}, 10*time.Second)

		_ = producer.SendDelay(ctx, &message.Message{
			Topic:   "delay2-test-topic",
			Payload: []byte("Hello, Delay2 Message---1!"),
		}, 20*time.Second)

		time.Sleep(5 * time.Second)
		_ = producer.SendDelay(ctx, &message.Message{
			Topic:   "delay2-test-topic",
			Payload: []byte("Hello, Delay2 Message---2!"),
		}, 20*time.Second)
	}()

	// 9. 消费消息
	consumer := mqInstance.Consumer()
	err = consumer.Subscribe(context.Background(), "test-topic", func(ctx context.Context, msg *message.Message) error {
		log.Printf("Received message: %s", string(msg.Payload))
		return nil
	})
	if err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	err = consumer.Subscribe(context.Background(), "delay-test-topic", func(ctx context.Context, msg *message.Message) error {
		log.Printf("Received delay message: %s", string(msg.Payload))
		return nil
	})
	if err != nil {
		log.Fatal("Failed to delay subscribe:", err)
	}

	err = consumer.Subscribe(context.Background(), "delay2-test-topic", func(ctx context.Context, msg *message.Message) error {
		log.Printf("Received delay message: %s", string(msg.Payload))
		return nil
	})
	if err != nil {
		log.Fatal("Failed to delay subscribe:", err)
	}

	// 等待一段时间让消息处理完成
	time.Sleep(1 * time.Minute)

	log.Println("Example completed successfully")
}
