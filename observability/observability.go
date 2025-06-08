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
	queueBacklog       metric.Int64UpDownCounter
	errorRate          metric.Float64Gauge
	throughput         metric.Float64Counter
	processingErrors   metric.Int64Counter
	retryAttempts      metric.Int64Counter
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

	queueBacklog, err := meter.Int64UpDownCounter(
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

	return &MetricsRecorder{
		meter:              meter,
		adapter:            adapter,
		messagesSent:       messagesSent,
		messagesReceived:   messagesReceived,
		errorCount:         errorCount,
		processingTime:     processingTime,
		connectionPoolSize: connectionPoolSize,
		messageLatency:     messageLatency,
		queueBacklog:       queueBacklog,
		errorRate:          errorRate,
		throughput:         throughput,
		processingErrors:   processingErrors,
		retryAttempts:      retryAttempts,
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
	m.queueBacklog.Add(ctx, backlog, metric.WithAttributes(
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
