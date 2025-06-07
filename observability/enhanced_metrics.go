package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// EnhancedMetricsRecorder 增强的指标记录器
type EnhancedMetricsRecorder struct {
	*MetricsRecorder

	// 新增指标
	connectionPool   metric.Int64Gauge
	messageLatency   metric.Float64Histogram
	queueBacklog     metric.Int64Gauge
	errorRate        metric.Float64Gauge
	throughput       metric.Float64Counter
	processingErrors metric.Int64Counter
	retryAttempts    metric.Int64Counter
}

// NewEnhancedMetricsRecorder 创建增强指标记录器
func NewEnhancedMetricsRecorder(observer Observer, adapter string) (*EnhancedMetricsRecorder, error) {
	base, err := NewMetricsRecorder(observer, adapter)
	if err != nil {
		return nil, err
	}

	if observer == nil {
		return &EnhancedMetricsRecorder{MetricsRecorder: base}, nil
	}

	meter := observer.GetMeter()

	// 创建增强指标
	connectionPool, err := meter.Int64Gauge(
		"mq_connection_pool_active",
		metric.WithDescription("Active connections in pool"),
	)
	if err != nil {
		return nil, err
	}

	messageLatency, err := meter.Float64Histogram(
		"mq_message_latency_seconds",
		metric.WithDescription("Message end-to-end latency"),
	)
	if err != nil {
		return nil, err
	}

	queueBacklog, err := meter.Int64Gauge(
		"mq_queue_backlog",
		metric.WithDescription("Number of messages waiting in queue"),
	)
	if err != nil {
		return nil, err
	}

	errorRate, err := meter.Float64Gauge(
		"mq_error_rate",
		metric.WithDescription("Error rate percentage"),
	)
	if err != nil {
		return nil, err
	}

	throughput, err := meter.Float64Counter(
		"mq_throughput_messages_per_second",
		metric.WithDescription("Message throughput per second"),
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

	return &EnhancedMetricsRecorder{
		MetricsRecorder:  base,
		connectionPool:   connectionPool,
		messageLatency:   messageLatency,
		queueBacklog:     queueBacklog,
		errorRate:        errorRate,
		throughput:       throughput,
		processingErrors: processingErrors,
		retryAttempts:    retryAttempts,
	}, nil
}

// RecordConnectionPoolSize 记录连接池大小
func (e *EnhancedMetricsRecorder) RecordConnectionPoolSize(ctx context.Context, size int64) {
	if e.connectionPool != nil {
		e.connectionPool.Record(ctx, size, metric.WithAttributes(
			attribute.String("adapter", e.adapter),
		))
	}
}

// RecordMessageLatency 记录消息延迟
func (e *EnhancedMetricsRecorder) RecordMessageLatency(ctx context.Context, topic string, latency time.Duration) {
	if e.messageLatency != nil {
		e.messageLatency.Record(ctx, latency.Seconds(), metric.WithAttributes(
			attribute.String("adapter", e.adapter),
			attribute.String("topic", topic),
		))
	}
}

// RecordQueueBacklog 记录队列积压
func (e *EnhancedMetricsRecorder) RecordQueueBacklog(ctx context.Context, topic string, backlog int64) {
	if e.queueBacklog != nil {
		e.queueBacklog.Record(ctx, backlog, metric.WithAttributes(
			attribute.String("adapter", e.adapter),
			attribute.String("topic", topic),
		))
	}
}

// RecordErrorRate 记录错误率
func (e *EnhancedMetricsRecorder) RecordErrorRate(ctx context.Context, topic string, rate float64) {
	if e.errorRate != nil {
		e.errorRate.Record(ctx, rate, metric.WithAttributes(
			attribute.String("adapter", e.adapter),
			attribute.String("topic", topic),
		))
	}
}

// RecordThroughput 记录吞吐量
func (e *EnhancedMetricsRecorder) RecordThroughput(ctx context.Context, topic string, count float64) {
	if e.throughput != nil {
		e.throughput.Add(ctx, count, metric.WithAttributes(
			attribute.String("adapter", e.adapter),
			attribute.String("topic", topic),
		))
	}
}

// RecordProcessingError 记录处理错误
func (e *EnhancedMetricsRecorder) RecordProcessingError(ctx context.Context, topic string, errorType string) {
	if e.processingErrors != nil {
		e.processingErrors.Add(ctx, 1, metric.WithAttributes(
			attribute.String("adapter", e.adapter),
			attribute.String("topic", topic),
			attribute.String("error_type", errorType),
		))
	}
}

// RecordRetryAttempt 记录重试尝试
func (e *EnhancedMetricsRecorder) RecordRetryAttempt(ctx context.Context, topic string, attempt int) {
	if e.retryAttempts != nil {
		e.retryAttempts.Add(ctx, 1, metric.WithAttributes(
			attribute.String("adapter", e.adapter),
			attribute.String("topic", topic),
			attribute.Int("attempt", attempt),
		))
	}
}
