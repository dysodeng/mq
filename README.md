# 消息队列 (Message Queue)

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/dysodeng/mq)](https://goreportcard.com/report/github.com/dysodeng/mq)

一个高性能、可扩展的Go语言消息队列包，支持多种底层实现和企业级特性。

## ✨ 特性

- 🚀 **高性能**: 基于连接池和批处理优化
- 🔧 **多适配器支持**: 支持Redis、RabbitMQ、Kafka等主流消息队列
- ⏰ **延时队列**: 基于时间轮算法的高效延时消息处理
- 📊 **可观测性**: 集成Prometheus指标、OpenTelemetry链路追踪和结构化日志
- 🏗️ **优雅架构**: 统一接口设计，易于扩展和维护
- 🛡️ **企业级**: 完善的错误处理、重试机制和健康检查
- 🔑 **Key前缀支持**: 全局key前缀，支持多租户隔离
- 🎯 **类型安全**: 强类型设计，适配器验证
- 📦 **零依赖**: 可选的可观测性组件

## 🚀 快速开始

### 安装

```bash
go get github.com/dysodeng/mq
```

### 基本用法
```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/dysodeng/mq"
    "github.com/dysodeng/mq/config"
    "github.com/dysodeng/mq/message"
)

func main() {
    // 创建配置
    cfg := config.Config{
        Adapter:   config.AdapterRedis,
        KeyPrefix: "app:mq",
        Redis:     config.DefaultRedisConfig(),
    }

    // 创建MQ工厂
    factory := mq.NewFactory(cfg)
    mqInstance, err := factory.CreateMQ()
    if err != nil {
        log.Fatal("创建MQ失败:", err)
    }
    defer mqInstance.Close()

    ctx := context.Background()

    // 生产者示例
    producer := mqInstance.Producer()
    msg := &message.Message{
        Topic:   "test-topic",
        Payload: []byte("Hello, World!"),
        Headers: map[string]string{
            "source": "example",
        },
    }

    err = producer.Send(ctx, msg)
    if err != nil {
        log.Fatal("发送消息失败:", err)
    }

    // 消费者示例
    consumer := mqInstance.Consumer()
    err = consumer.Subscribe(ctx, "test-topic", func(ctx context.Context, msg *message.Message) error {
        log.Printf("收到消息: %s", string(msg.Payload))
        return nil
    })
    if err != nil {
        log.Fatal("订阅失败:", err)
    }

    // 延时队列示例
    delayMsg := &message.Message{
        Topic:   "delay-topic",
        Payload: []byte("延时消息"),
    }
    err = mqInstance.DelayQueue().Push(ctx, delayMsg, 10*time.Second)
    if err != nil {
        log.Fatal("发送延时消息失败:", err)
    }

    time.Sleep(30 * time.Second)
}
```

## 📋 支持的适配器

| 适配器      | 状态 | 特性                                  |
|----------|----|-------------------------------------|
| Redis    | ✅  | 基于List的队列，Sorted sets实现延时，支持集群和哨兵模式 |
| RabbitMQ | ✅  | AMQP协议，Exchange路由，持久化支持             |
| Kafka    | ✅  | 分布式流处理，分区支持，高吞吐量                    |

## ⚙️ 配置
### Redis 配置
```go
cfg := config.Config{
    Adapter:   config.AdapterRedis,
    KeyPrefix: "myapp",
    Redis: config.RedisConfig{
        // 连接配置
        Mode:     config.RedisModeSingle, // 支持Single, Cluster, Sentinel
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
        
        // 连接池配置
        PoolSize:     10,
        MinIdleConns: 5,
        MaxConnAge:   time.Hour,
        PoolTimeout:  30 * time.Second,
        IdleTimeout:  5 * time.Minute,
        
        // 性能配置
        ConsumerWorkerCount: 5,
        ConsumerBatchSize:   10,
        ProducerBatchSize:   100,
        
        // 序列化配置
        SerializationType: "json", // 支持json, msgpack, protobuf
    },
}
```
### RabbitMQ 配置
```go
cfg := config.Config{
    Adapter:   config.AdapterRabbitMQ,
    KeyPrefix: "myapp",
    RabbitMQ: config.RabbitMQConfig{
        URL:              "amqp://guest:guest@localhost:5672/",
        Exchange:         "mq.direct",
        ExchangeType:     "direct",
        QueueDurable:     true,
        QueueAutoDelete:  false,
        QueueExclusive:   false,
        QoS:              10,
        Heartbeat:        60 * time.Second,
        ConnectionTimeout: 30 * time.Second,
    },
}
```
### Kafka 配置
```go
cfg := config.Config{
    Adapter:   config.AdapterKafka,
    KeyPrefix: "myapp",
    Kafka: config.KafkaConfig{
        Brokers:  []string{"localhost:9092"},
        GroupID:  "mq-consumer-group",
        ClientID: "mq-client",
        Version:  "2.8.0",
        Producer: config.KafkaProducerConfig{
            MaxMessageBytes: 1000000,
            RequiredAcks:    1,
            Timeout:         30 * time.Second,
            Compression:     "snappy",
            Idempotent:      true,
        },
        Consumer: config.KafkaConsumerConfig{
            MinBytes:          1,
            MaxBytes:          1048576,
            MaxWait:           500 * time.Millisecond,
            CommitInterval:    1 * time.Second,
            StartOffset:       -1,
            HeartbeatInterval: 3 * time.Second,
            SessionTimeout:    30 * time.Second,
            RebalanceTimeout:  30 * time.Second,
        },
    },
}
```

## 📊 可观测性
该包通过OpenTelemetry和结构化日志提供全面的可观测性支持。

### 使用自定义Observer

```go
package main

import (
    "github.com/dysodeng/mq"
    "github.com/dysodeng/mq/config"
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
    // 初始化OpenTelemetry和zap日志
    meter := otel.Meter("mq-service")
    logger, _ := zap.NewProduction()
    
    observer := &MyObserver{
        meter:  meter,
        logger: logger,
    }

    cfg := config.Config{
        Adapter:   config.AdapterRedis,
        KeyPrefix: "app:mq",
        Redis:     config.DefaultRedisConfig(),
    }

    // 使用observer创建工厂
    factory := mq.NewFactory(cfg, mq.WithObserver(observer))
    mqInstance, err := factory.CreateMQ()
    if err != nil {
        panic(err)
    }
    defer mqInstance.Close()

    // 你的应用逻辑...
}
```

### 可用指标

#### 基础指标
- `mq_messages_sent_total` - 发送消息总数
- `mq_messages_received_total` - 接收消息总数
- `mq_errors_total` - 错误总数
- `mq_processing_duration_seconds` - 消息处理耗时（直方图）

#### 增强指标
- `mq_connection_pool_size` - 当前连接池大小
- `mq_message_latency_seconds` - 消息端到端延迟（直方图）
- `mq_queue_backlog` - 当前队列积压大小
- `mq_error_rate` - 当前错误率
- `mq_throughput_total` - 总吞吐量
- `mq_processing_errors_total` - 处理错误总数
- `mq_retry_attempts_total` - 重试尝试总数

#### 指标标签
所有指标都包含以下标签：
- `adapter` - 适配器类型（redis/rabbitmq/kafka）
- `topic` - 消息主题
- `error` - 错误信息（仅错误相关指标）
- `error_type` - 错误类型（仅处理错误指标）
- `attempt` - 重试次数（仅重试指标）

#### 指标类型说明
- **Counter**: 累计计数器，只增不减
- **Gauge**: 瞬时值，可增可减
- **Histogram**: 直方图，记录数值分布
- **UpDownCounter**: 可增减计数器

## 🏗️ 架构
```
┌─────────────────┐
│   Application   │
└─────────┬───────┘
          │
┌─────────▼───────┐
│     Factory     │  ← 工厂模式，支持多种适配器
└─────────┬───────┘
          │
┌─────────▼───────┐
│   MQ Interface  │  ← 统一接口层
├─────────────────┤
│   • Producer    │  ← 生产者：支持普通和延时消息
│   • Consumer    │  ← 消费者：支持多topic订阅
│   • DelayQueue  │  ← 延时队列：独立的延时消息管理
│   • HealthCheck │  ← 健康检查
└─────────┬───────┘
          │
┌─────────▼───────┐
│    Adapters     │  ← 适配器层
├─────────────────┤
│ Redis │RabbitMQ │  ← 支持多种消息队列后端
│ Kafka │  ...    │
└─────────┬───────┘
          │
┌─────────▼───────┐
│  Infrastructure │  ← 基础设施层
├─────────────────┤
│ • Serializer    │  ← 序列化：JSON/MessagePack/Protobuf
│ • Object Pool   │  ← 对象池：优化内存分配
│ • Observability │  ← 可观测性：指标/日志/链路追踪
│ • Worker Pool   │  ← 工作池：并发处理优化
└─────────────────┘
```

## 🔧 高级特性

### 延时队列
延时队列使用时间轮算法实现高效的延时消息处理：

```go
// 方式1：使用DelayQueue接口
msg := &message.Message{
    Topic:   "notification",
    Payload: []byte("Reminder: Meeting in 1 hour"),
}

// 1小时后投递
err := mqInstance.DelayQueue().Push(ctx, msg, time.Hour)

// 方式2：使用Producer的SendDelay方法
err := producer.SendDelay(ctx, msg, time.Hour)

// 查询延时队列大小
size, err := mqInstance.DelayQueue().Size(ctx)

// 移除特定延时消息
err = mqInstance.DelayQueue().Remove(ctx, "message-id")
```

### 批量操作
```go
// 批量发送消息
messages := []*message.Message{
    {Topic: "topic1", Payload: []byte("msg1")},
    {Topic: "topic2", Payload: []byte("msg2")},
}

err := producer.SendBatch(ctx, messages)
```

### 消息结构
```go
type Message struct {
    ID       string            `json:"id"`        // 消息唯一标识
    Topic    string            `json:"topic"`     // 消息主题
    Payload  []byte            `json:"payload"`   // 消息内容
    Headers  map[string]string `json:"headers"`   // 消息头
    Delay    time.Duration     `json:"delay"`     // 延时时间（可选）
    Retry    int               `json:"retry"`     // 重试次数
    CreateAt time.Time         `json:"create_at"` // 创建时间
}
```

### 健康检查
```go
// 检查MQ健康状态
if err := mqInstance.HealthCheck(); err != nil {
    log.Printf("MQ health check failed: %v", err)
}
```

### 序列化支持
支持多种序列化格式：

- JSON : 默认格式，易于调试
- MessagePack : 二进制格式，更高效
- Protobuf : 强类型，跨语言支持
```go
// 配置序列化类型
redisConfig.SerializationType = "msgpack"
redisConfig.SerializationCompression = true // 启用压缩
```

### 性能优化
```go
// Redis性能配置示例
redisConfig := config.RedisConfig{
    // 消费者优化
    ConsumerWorkerCount:   10,              // 消费者工作协程数
    ConsumerBatchSize:     50,              // 批量消费大小
    ConsumerPollTimeout:   time.Second,     // 轮询超时
    
    // 生产者优化
    ProducerBatchSize:     100,             // 批量发送大小
    ProducerFlushInterval: 100*time.Millisecond, // 刷新间隔
    
    // 连接池优化
    PoolSize:              20,              // 连接池大小
    MinIdleConns:          5,               // 最小空闲连接
}
```

## 🧪 测试
```bash
# 运行测试
go test ./...

# 运行测试并查看覆盖率
go test -cover ./...

# 运行基准测试
go test -bench=. ./...
```

## 📝 示例
查看 examples 目录获取更多完整的使用示例：

- basic - 基本的生产者/消费者示例
- performance - 性能测试和批量处理示例
- with_observability - 完整的可观测性设置示例

## 📄 许可证
本项目采用 MIT 许可证 - 查看 LICENSE 文件了解详情

## 🙏 致谢
- [go-redis](https://github.com/go-redis/redis) 提供Redis客户端
- [amqp091-go](https://github.com/rabbitmq/amqp091-go) 提供RabbitMQ客户端
- [kafka-go](https://github.com/segmentio/kafka-go) 提供Kafka客户端
- [OpenTelemetry](https://opentelemetry.io/docs/) 提供可观测性支持
- [Zap](https://github.com/uber-go/zap) 提供结构化日志