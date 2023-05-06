package message

import (
	"strings"

	"github.com/google/uuid"
)

// Message 队列消息体
type Message struct {
	id   string // 消息ID
	body string // 消息体
	key  string // 队列key
}

// NewMessage 创建队列消息体
func NewMessage(key string, id, body string) Message {
	if id == "" {
		id = strings.Replace(uuid.New().String(), "-", "", -1)
	}
	return Message{
		id:   id,
		body: body,
		key:  key,
	}
}

func (message *Message) Id() string {
	return message.id
}

func (message *Message) Body() string {
	return message.body
}

func (message *Message) Key() string {
	return message.key
}
