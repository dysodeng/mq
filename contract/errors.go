package contract

import "fmt"

// ErrorCode 错误代码
type ErrorCode int

const (
	ErrCodeUnknown ErrorCode = iota
	ErrCodeConnectionFailed
	ErrCodeMessageSendFailed
	ErrCodeMessageReceiveFailed
	ErrCodeQueueNotFound
	ErrCodeSerializationFailed
	ErrCodeDeserializationFailed
	ErrCodeTimeout
	ErrCodeInvalidConfig
	ErrCodeClientClosed
)

// MQError 消息队列错误
type MQError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Cause   error     `json:"cause,omitempty"`
}

// Error 实现error接口
func (e *MQError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *MQError) Unwrap() error {
	return e.Cause
}

// NewMQError 创建MQ错误
func NewMQError(code ErrorCode, message string, cause error) *MQError {
	return &MQError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// 预定义错误
var (
	ErrConnectionFailed      = NewMQError(ErrCodeConnectionFailed, "connection failed", nil)
	ErrMessageSendFailed     = NewMQError(ErrCodeMessageSendFailed, "message send failed", nil)
	ErrMessageReceiveFailed  = NewMQError(ErrCodeMessageReceiveFailed, "message receive failed", nil)
	ErrQueueNotFound         = NewMQError(ErrCodeQueueNotFound, "queue not found", nil)
	ErrSerializationFailed   = NewMQError(ErrCodeSerializationFailed, "serialization failed", nil)
	ErrDeserializationFailed = NewMQError(ErrCodeDeserializationFailed, "deserialization failed", nil)
	ErrTimeout               = NewMQError(ErrCodeTimeout, "operation timeout", nil)
	ErrInvalidConfig         = NewMQError(ErrCodeInvalidConfig, "invalid configuration", nil)
	ErrClientClosed          = NewMQError(ErrCodeClientClosed, "client is closed", nil)
)
