package consumer

import "github.com/dysoodeng/mq/message"

// Handler 消费者接口
type Handler interface {
	// Handle 消费者处理器
	Handle(message message.Message) error
}
