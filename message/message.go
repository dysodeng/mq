package message

import (
	"context"
	"time"

	"github.com/google/uuid"
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

func New(topic string, payload []byte) *Message {
	return &Message{
		ID:      uuid.NewString(),
		Topic:   topic,
		Payload: payload,
		Headers: make(map[string]string),
	}
}

func (m *Message) SetID(id string) *Message {
	m.ID = id
	return m
}

func (m *Message) SetHeaders(headers map[string]string) *Message {
	m.Headers = headers
	return m
}

func (m *Message) SetDelay(delay time.Duration) *Message {
	m.Delay = delay
	return m
}

func (m *Message) SetRetry(retry int) *Message {
	m.Retry = retry
	return m
}

func (m *Message) SetCreateAt(createAt time.Time) *Message {
	m.CreateAt = createAt
	return m
}

func (m *Message) GetID() string {
	return m.ID
}

func (m *Message) GetTopic() string {
	return m.Topic
}

func (m *Message) GetPayload() []byte {
	return m.Payload
}

func (m *Message) GetHeaders() map[string]string {
	return m.Headers
}

func (m *Message) GetDelay() time.Duration {
	return m.Delay
}

func (m *Message) GetRetry() int {
	return m.Retry
}

func (m *Message) GetCreateAt() time.Time {
	return m.CreateAt
}

// Handler 消息处理函数
type Handler func(ctx context.Context, msg *Message) error
