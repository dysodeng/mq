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

	// 预定义的指标
	messagesSent     metric.Int64Counter
	messagesReceived metric.Int64Counter
	messagesFailed   metric.Int64Counter
	processingTime   metric.Float64Histogram
	queueSize        metric.Int64Gauge
}

// NewMetricsRecorder 创建指标记录器
func NewMetricsRecorder(observer Observer, adapter string) (*MetricsRecorder, error) {
	if observer == nil {
		return &MetricsRecorder{adapter: adapter}, nil
	}

	meter := observer.GetMeter()
	logger := observer.GetLogger()

	// 创建指标
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
		metric.WithDescription("Total number of failed messages"),
	)
	if err != nil {
		return nil, err
	}

	processingTime, err := meter.Float64Histogram(
		"mq_message_processing_duration_seconds",
		metric.WithDescription("Message processing duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	queueSize, err := meter.Int64Gauge(
		"mq_queue_size",
		metric.WithDescription("Current queue size"),
	)
	if err != nil {
		return nil, err
	}

	return &MetricsRecorder{
		meter:            meter,
		logger:           logger,
		adapter:          adapter,
		messagesSent:     messagesSent,
		messagesReceived: messagesReceived,
		messagesFailed:   messagesFailed,
		processingTime:   processingTime,
		queueSize:        queueSize,
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
