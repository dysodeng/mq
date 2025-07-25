package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// Observer 可观测性接口，由外部调用方实现并注入
type Observer interface {
	// GetMeter 获取OpenTelemetry Meter实例
	GetMeter() metric.Meter
	// GetLogger 获取zap Logger实例
	GetLogger() *zap.Logger
}

// PoolStatistics 池统计信息结构体
type PoolStatistics struct {
	HitRate           float64 // 命中率
	ObjectCount       int64   // 对象数量
	AverageBufferSize float64 // 平均缓冲区大小
}

// MetricsRecorder 指标记录器
type MetricsRecorder struct {
	meter   metric.Meter
	logger  *zap.Logger
	adapter string

	// 基础指标
	messagesSent     metric.Int64Counter
	messagesReceived metric.Int64Counter
	messagesFailed   metric.Int64Counter
	queueSize        metric.Int64Gauge
	errorCount       metric.Int64Counter
	processingTime   metric.Float64Histogram

	// 增强指标
	connectionPoolSize metric.Int64Gauge
	messageLatency     metric.Float64Histogram
	queueBacklog       metric.Int64Gauge
	errorRate          metric.Float64Gauge
	throughput         metric.Float64Counter
	processingErrors   metric.Int64Counter
	retryAttempts      metric.Int64Counter

	// 对象池指标
	poolHitRate         metric.Float64Gauge
	poolObjectCount     metric.Int64Gauge
	objectGetTotal      metric.Int64Counter
	objectPutTotal      metric.Int64Counter
	objectCreateTotal   metric.Int64Counter
	objectUsageDuration metric.Float64Histogram

	// 缓冲池指标
	bufferPoolSize     metric.Int64Gauge
	bufferAverageSize  metric.Float64Gauge
	bufferDiscardTotal metric.Int64Counter
}

// NewMetricsRecorder 创建指标记录器
func NewMetricsRecorder(observer Observer, adapter string) (*MetricsRecorder, error) {
	meter := observer.GetMeter()

	// 创建基础指标
	messagesSent, err := meter.Int64Counter(
		"mq_messages_sent_total",
		metric.WithDescription("Total number of messages sent"),
	)
	if err != nil {
		return nil, err
	}

	messagesReceived, err := meter.Int64Counter(
		"mq_messages_received_total",
		metric.WithDescription("Total number of messages received"),
	)
	if err != nil {
		return nil, err
	}

	messagesFailed, err := meter.Int64Counter(
		"mq_messages_failed_total",
		metric.WithDescription("Total number of messages failed"),
	)
	if err != nil {
		return nil, err
	}

	errorCount, err := meter.Int64Counter(
		"mq_errors_total",
		metric.WithDescription("Total number of errors"),
	)
	if err != nil {
		return nil, err
	}

	processingTime, err := meter.Float64Histogram(
		"mq_processing_duration_seconds",
		metric.WithDescription("Message processing duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	// 创建增强指标
	connectionPoolSize, err := meter.Int64Gauge(
		"mq_connection_pool_size",
		metric.WithDescription("Current connection pool size"),
	)
	if err != nil {
		return nil, err
	}

	messageLatency, err := meter.Float64Histogram(
		"mq_message_latency_seconds",
		metric.WithDescription("Message end-to-end latency in seconds"),
	)
	if err != nil {
		return nil, err
	}

	queueBacklog, err := meter.Int64Gauge(
		"mq_queue_backlog",
		metric.WithDescription("Current queue backlog size"),
	)
	if err != nil {
		return nil, err
	}

	errorRate, err := meter.Float64Gauge(
		"mq_error_rate",
		metric.WithDescription("Current error rate"),
	)
	if err != nil {
		return nil, err
	}

	throughput, err := meter.Float64Counter(
		"mq_throughput_total",
		metric.WithDescription("Total throughput"),
	)
	if err != nil {
		return nil, err
	}

	processingErrors, err := meter.Int64Counter(
		"mq_processing_errors_total",
		metric.WithDescription("Total processing errors"),
	)
	if err != nil {
		return nil, err
	}

	retryAttempts, err := meter.Int64Counter(
		"mq_retry_attempts_total",
		metric.WithDescription("Total retry attempts"),
	)
	if err != nil {
		return nil, err
	}

	// 创建对象池指标
	poolHitRate, err := meter.Float64Gauge(
		"mq_pool_hit_rate",
		metric.WithDescription("Object pool hit rate"),
	)
	if err != nil {
		return nil, err
	}

	poolObjectCount, err := meter.Int64Gauge(
		"mq_pool_object_count",
		metric.WithDescription("Current number of objects in pool"),
	)
	if err != nil {
		return nil, err
	}

	objectGetTotal, err := meter.Int64Counter(
		"mq_object_get_total",
		metric.WithDescription("Total number of object get operations"),
	)
	if err != nil {
		return nil, err
	}

	objectPutTotal, err := meter.Int64Counter(
		"mq_object_put_total",
		metric.WithDescription("Total number of object put operations"),
	)
	if err != nil {
		return nil, err
	}

	objectCreateTotal, err := meter.Int64Counter(
		"mq_object_create_total",
		metric.WithDescription("Total number of new objects created"),
	)
	if err != nil {
		return nil, err
	}

	objectUsageDuration, err := meter.Float64Histogram(
		"mq_object_usage_duration_seconds",
		metric.WithDescription("Object usage duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	// 创建缓冲池指标
	bufferPoolSize, err := meter.Int64Gauge(
		"mq_buffer_pool_size",
		metric.WithDescription("Current buffer pool size"),
	)
	if err != nil {
		return nil, err
	}

	bufferAverageSize, err := meter.Float64Gauge(
		"mq_buffer_average_size_bytes",
		metric.WithDescription("Average buffer size in bytes"),
	)
	if err != nil {
		return nil, err
	}

	bufferDiscardTotal, err := meter.Int64Counter(
		"mq_buffer_discard_total",
		metric.WithDescription("Total number of buffers discarded due to size limit"),
	)
	if err != nil {
		return nil, err
	}

	return &MetricsRecorder{
		meter:              meter,
		adapter:            adapter,
		messagesSent:       messagesSent,
		messagesReceived:   messagesReceived,
		messagesFailed:     messagesFailed,
		errorCount:         errorCount,
		processingTime:     processingTime,
		connectionPoolSize: connectionPoolSize,
		messageLatency:     messageLatency,
		queueBacklog:       queueBacklog,
		errorRate:          errorRate,
		throughput:         throughput,
		processingErrors:   processingErrors,
		retryAttempts:      retryAttempts,
		// 对象池指标
		poolHitRate:         poolHitRate,
		poolObjectCount:     poolObjectCount,
		objectGetTotal:      objectGetTotal,
		objectPutTotal:      objectPutTotal,
		objectCreateTotal:   objectCreateTotal,
		objectUsageDuration: objectUsageDuration,
		// 缓冲池指标
		bufferPoolSize:     bufferPoolSize,
		bufferAverageSize:  bufferAverageSize,
		bufferDiscardTotal: bufferDiscardTotal,
	}, nil
}

// RecordMessageSent 记录消息发送
func (m *MetricsRecorder) RecordMessageSent(ctx context.Context, topic string) {
	if m.messagesSent != nil {
		m.messagesSent.Add(ctx, 1, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.String("topic", topic),
		))
	}
	if m.logger != nil {
		m.logger.Debug("Message sent",
			zap.String("adapter", m.adapter),
			zap.String("topic", topic),
		)
	}
}

// RecordMessageReceived 记录消息接收
func (m *MetricsRecorder) RecordMessageReceived(ctx context.Context, topic string) {
	if m.messagesReceived != nil {
		m.messagesReceived.Add(ctx, 1, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.String("topic", topic),
		))
	}
	if m.logger != nil {
		m.logger.Debug("Message received",
			zap.String("adapter", m.adapter),
			zap.String("topic", topic),
		)
	}
}

// RecordMessageFailed 记录消息失败
func (m *MetricsRecorder) RecordMessageFailed(ctx context.Context, topic string, err error) {
	if m.messagesFailed != nil {
		m.messagesFailed.Add(ctx, 1, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.String("topic", topic),
			attribute.String("error", err.Error()),
		))
	}
	if m.logger != nil {
		m.logger.Error("Message failed",
			zap.String("adapter", m.adapter),
			zap.String("topic", topic),
			zap.Error(err),
		)
	}
}

// RecordProcessingTime 记录处理时间
func (m *MetricsRecorder) RecordProcessingTime(ctx context.Context, topic string, duration time.Duration) {
	if m.processingTime != nil {
		m.processingTime.Record(ctx, duration.Seconds(), metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.String("topic", topic),
		))
	}
	if m.logger != nil {
		m.logger.Debug("Message processed",
			zap.String("adapter", m.adapter),
			zap.String("topic", topic),
			zap.Duration("duration", duration),
		)
	}
}

// RecordQueueSize 记录队列大小
func (m *MetricsRecorder) RecordQueueSize(ctx context.Context, topic string, size int64) {
	if m.queueSize != nil {
		m.queueSize.Record(ctx, size, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.String("topic", topic),
		))
	}
}

// RecordConnectionPoolSize 记录连接池大小
func (m *MetricsRecorder) RecordConnectionPoolSize(ctx context.Context, size int64) {
	m.connectionPoolSize.Record(ctx, size, metric.WithAttributes(
		attribute.String("adapter", m.adapter),
	))
}

// RecordMessageLatency 记录消息延迟
func (m *MetricsRecorder) RecordMessageLatency(ctx context.Context, topic string, latency time.Duration) {
	m.messageLatency.Record(ctx, latency.Seconds(), metric.WithAttributes(
		attribute.String("adapter", m.adapter),
		attribute.String("topic", topic),
	))
}

// RecordQueueBacklog 记录队列积压
func (m *MetricsRecorder) RecordQueueBacklog(ctx context.Context, topic string, backlog int64) {
	m.queueBacklog.Record(ctx, backlog, metric.WithAttributes(
		attribute.String("adapter", m.adapter),
		attribute.String("topic", topic),
	))
}

// RecordErrorRate 记录错误率
func (m *MetricsRecorder) RecordErrorRate(ctx context.Context, topic string, rate float64) {
	m.errorRate.Record(ctx, rate, metric.WithAttributes(
		attribute.String("adapter", m.adapter),
		attribute.String("topic", topic),
	))
}

// RecordThroughput 记录吞吐量
func (m *MetricsRecorder) RecordThroughput(ctx context.Context, topic string, count float64) {
	m.throughput.Add(ctx, count, metric.WithAttributes(
		attribute.String("adapter", m.adapter),
		attribute.String("topic", topic),
	))
}

// RecordProcessingError 记录处理错误
func (m *MetricsRecorder) RecordProcessingError(ctx context.Context, topic string, errorType string) {
	m.processingErrors.Add(ctx, 1, metric.WithAttributes(
		attribute.String("adapter", m.adapter),
		attribute.String("topic", topic),
		attribute.String("error_type", errorType),
	))
}

// RecordRetryAttempt 记录重试尝试
func (m *MetricsRecorder) RecordRetryAttempt(ctx context.Context, topic string, attempt int) {
	m.retryAttempts.Add(ctx, 1, metric.WithAttributes(
		attribute.String("adapter", m.adapter),
		attribute.String("topic", topic),
		attribute.Int("attempt", attempt),
	))
}

// RecordPoolHitRate 记录对象池命中率
func (m *MetricsRecorder) RecordPoolHitRate(ctx context.Context, poolType string, hitRate float64) {
	if m.poolHitRate != nil {
		m.poolHitRate.Record(ctx, hitRate, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.String("pool_type", poolType),
		))
	}
}

// RecordPoolObjectCount 记录对象池中对象数量
func (m *MetricsRecorder) RecordPoolObjectCount(ctx context.Context, poolType string, count int64) {
	if m.poolObjectCount != nil {
		m.poolObjectCount.Record(ctx, count, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.String("pool_type", poolType),
		))
	}
}

// RecordObjectGet 记录对象获取操作
func (m *MetricsRecorder) RecordObjectGet(ctx context.Context, poolType string) {
	if m.objectGetTotal != nil {
		m.objectGetTotal.Add(ctx, 1, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.String("pool_type", poolType),
		))
	}
}

// RecordObjectPut 记录对象归还操作
func (m *MetricsRecorder) RecordObjectPut(ctx context.Context, poolType string) {
	if m.objectPutTotal != nil {
		m.objectPutTotal.Add(ctx, 1, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.String("pool_type", poolType),
		))
	}
}

// RecordObjectCreate 记录新对象创建
func (m *MetricsRecorder) RecordObjectCreate(ctx context.Context, poolType string) {
	if m.objectCreateTotal != nil {
		m.objectCreateTotal.Add(ctx, 1, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.String("pool_type", poolType),
		))
	}
	if m.logger != nil {
		m.logger.Debug("New object created",
			zap.String("adapter", m.adapter),
			zap.String("pool_type", poolType),
		)
	}
}

// RecordObjectUsage 记录对象使用时长
func (m *MetricsRecorder) RecordObjectUsage(ctx context.Context, poolType string, duration time.Duration) {
	if m.objectUsageDuration != nil {
		m.objectUsageDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.String("pool_type", poolType),
		))
	}
}

// RecordBufferPoolSize 记录缓冲池大小
func (m *MetricsRecorder) RecordBufferPoolSize(ctx context.Context, size int64) {
	if m.bufferPoolSize != nil {
		m.bufferPoolSize.Record(ctx, size, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
		))
	}
}

// RecordBufferAverageSize 记录缓冲区平均大小
func (m *MetricsRecorder) RecordBufferAverageSize(ctx context.Context, avgSize float64) {
	if m.bufferAverageSize != nil {
		m.bufferAverageSize.Record(ctx, avgSize, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
		))
	}
}

// RecordBufferDiscard 记录缓冲区丢弃事件
func (m *MetricsRecorder) RecordBufferDiscard(ctx context.Context, bufferSize int) {
	if m.bufferDiscardTotal != nil {
		m.bufferDiscardTotal.Add(ctx, 1, metric.WithAttributes(
			attribute.String("adapter", m.adapter),
			attribute.Int("buffer_size", bufferSize),
		))
	}
	if m.logger != nil {
		m.logger.Debug("Buffer discarded due to size limit",
			zap.String("adapter", m.adapter),
			zap.Int("buffer_size", bufferSize),
		)
	}
}

// RecordPoolStatistics 批量记录池统计信息
func (m *MetricsRecorder) RecordPoolStatistics(ctx context.Context, poolType string, stats PoolStatistics) {
	m.RecordPoolHitRate(ctx, poolType, stats.HitRate)
	m.RecordPoolObjectCount(ctx, poolType, stats.ObjectCount)
	if stats.AverageBufferSize > 0 {
		m.RecordBufferAverageSize(ctx, stats.AverageBufferSize)
	}
}

// LogInfo 记录信息日志
func (m *MetricsRecorder) LogInfo(msg string, fields ...zap.Field) {
	if m.logger != nil {
		fields = append(fields, zap.String("adapter", m.adapter))
		m.logger.Info(msg, fields...)
	}
}

// LogError 记录错误日志
func (m *MetricsRecorder) LogError(msg string, err error, fields ...zap.Field) {
	if m.logger != nil {
		fields = append(fields, zap.String("adapter", m.adapter), zap.Error(err))
		m.logger.Error(msg, fields...)
	}
}

// LogWarn 记录警告日志
func (m *MetricsRecorder) LogWarn(msg string, fields ...zap.Field) {
	if m.logger != nil {
		fields = append(fields, zap.String("adapter", m.adapter))
		m.logger.Warn(msg, fields...)
	}
}

// LogDebug 记录调试日志
func (m *MetricsRecorder) LogDebug(msg string, fields ...zap.Field) {
	if m.logger != nil {
		fields = append(fields, zap.String("adapter", m.adapter))
		m.logger.Debug(msg, fields...)
	}
}
