package pool

import (
	"sync"
	"time"

	"github.com/dysodeng/mq/message"
)

// MessagePool 消息对象池
type MessagePool struct {
	pool sync.Pool
}

// NewMessagePool 创建消息池
func NewMessagePool() *MessagePool {
	return &MessagePool{
		pool: sync.Pool{
			New: func() interface{} {
				return &message.Message{
					Headers: make(map[string]string),
				}
			},
		},
	}
}

// Get 获取消息对象
func (p *MessagePool) Get() *message.Message {
	msg := p.pool.Get().(*message.Message)
	// 重置消息
	msg.ID = ""
	msg.Topic = ""
	msg.Payload = nil
	msg.Delay = 0
	msg.Retry = 0
	msg.CreateAt = time.Time{}
	// 清空headers但保留map
	for k := range msg.Headers {
		delete(msg.Headers, k)
	}
	return msg
}

// Put 归还消息对象
func (p *MessagePool) Put(msg *message.Message) {
	if msg != nil {
		p.pool.Put(msg)
	}
}

// ByteBufferPool 字节缓冲池
type ByteBufferPool struct {
	pool sync.Pool
}

// NewByteBufferPool 创建字节缓冲池
func NewByteBufferPool() *ByteBufferPool {
	return &ByteBufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, 1024) // 初始容量1KB
				return &buf
			},
		},
	}
}

// Get 获取字节缓冲
func (p *ByteBufferPool) Get() []byte {
	bufPtr := p.pool.Get().(*[]byte)
	*bufPtr = (*bufPtr)[:0] // 重置长度但保留容量
	return *bufPtr
}

// Put 归还字节缓冲
func (p *ByteBufferPool) Put(buf []byte) {
	if cap(buf) <= 64*1024 { // 限制最大容量64KB
		p.pool.Put(&buf) // 传递指针
	}
}
