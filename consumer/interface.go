package consumer

import "github.com/dysodeng/mq/message"

// Handler 消费者接口
type Handler interface {
	// Handle 消费者处理器
	Handle(message message.Message) error
}
