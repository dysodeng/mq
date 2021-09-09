package message

import (
	"github.com/google/uuid"
	"strings"
)

// Message 队列消息体
type Message struct {
	// id 消息ID
	id string
	// body 消息体
	body string
	// key 队列key
	key Key
}

// Key 消息队列key
type Key struct {
	// ExchangeName 交换机名称
	ExchangeName string
	// QueueName 队列名称
	QueueName string
	// RouteKey 队列路由
	RouteKey string
}

// NewMessage 创建队列消息体
func NewMessage(key Key, id, body string) Message {
	if id == "" {
		id = strings.Replace(uuid.New().String(), "-", "", -1)
	}
	return Message{
		id: id,
		body: body,
		key: key,
	}
}

func (message *Message) Id() string {
	return message.id
}

func (message *Message) Body() string {
	return message.body
}

func (message *Message) ExchangeName() string {
	return message.key.ExchangeName
}

func (message *Message) QueueName() string {
	return message.key.QueueName
}

func (message *Message) RouteKey() string {
	return message.key.RouteKey
}
