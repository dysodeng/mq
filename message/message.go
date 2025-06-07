package message

import (
	"context"
	"time"
)

// Message 消息结构
type Message struct {
	ID       string            `json:"id"`
	Topic    string            `json:"topic"`
	Payload  []byte            `json:"payload"`
	Headers  map[string]string `json:"headers"`
	Delay    time.Duration     `json:"delay,omitempty"`
	Retry    int               `json:"retry,omitempty"`
	CreateAt time.Time         `json:"create_at"`
}

// Handler 消息处理函数
type Handler func(ctx context.Context, msg *Message) error
