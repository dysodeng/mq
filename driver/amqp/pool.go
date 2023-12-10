package amqp

import (
	"fmt"
	"sync"
	"time"

	"github.com/dysodeng/mq/contract"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrClosed = errors.New("amqp connection pool is closed")
)

// poolConfig 连接池配置
type poolConfig struct {
	MinConn     int             // 连接池中拥有的最小连接数
	MaxConn     int             // 最大存活连接数
	MaxIdle     int             // 最大空闲连接数
	IdleTimeout time.Duration   // 连接最大空闲时间，超过该时间连接失效
	amqpConfig  contract.Config // amqp连接配置
}

// amqpConnectionPool amqp 连接池
type amqpConnectionPool struct {
	mu              sync.RWMutex
	connections     chan *idleConnection // 连接池
	idleTimeout     time.Duration        // 空闲过期时间
	maxActive       int                  // 最大活跃连接数
	openConnections int                  // 已打开的连接数
	amqpConfig      contract.Config      // amqp连接配置
}

// idleConnection 空闲连接
type idleConnection struct {
	conn *amqp.Connection // amqp连接
	t    time.Time        // 空闲时间
}

func newAmqpConnectionPool(poolConfig poolConfig) (*amqpConnectionPool, error) {
	if !(poolConfig.MinConn <= poolConfig.MaxIdle && poolConfig.MaxConn >= poolConfig.MaxIdle && poolConfig.MinConn >= 0) {
		return nil, errors.New("invalid config")
	}

	connPool := &amqpConnectionPool{
		connections:     make(chan *idleConnection, poolConfig.MaxIdle),
		idleTimeout:     poolConfig.IdleTimeout,
		maxActive:       poolConfig.MaxConn,
		openConnections: poolConfig.MinConn,
		amqpConfig:      poolConfig.amqpConfig,
	}

	for i := 0; i < poolConfig.MinConn; i++ {
		conn, err := connPool.factory()
		if err != nil {
			connPool.Release()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}

		connPool.connections <- &idleConnection{conn: conn, t: time.Now()}
	}

	return connPool, nil
}

// factory 创建连接
func (pool *amqpConnectionPool) factory() (*amqp.Connection, error) {
	return amqp.Dial(pool.amqpConfig.String())
}

// close 关闭连接
func (pool *amqpConnectionPool) close(conn *amqp.Connection) error {
	return conn.Close()
}

// ping 连接健康检查
func (pool *amqpConnectionPool) ping(conn *amqp.Connection) error {
	return nil
}

// getConnections 获取所有连接
func (pool *amqpConnectionPool) getConnections() chan *idleConnection {
	pool.mu.Lock()
	connections := pool.connections
	pool.mu.Unlock()
	return connections
}

// Get 从连接池中获取连接
func (pool *amqpConnectionPool) Get() (*amqp.Connection, error) {
	connections := pool.getConnections()
	if connections == nil {
		return nil, ErrClosed
	}

	for {
		select {
		case wrapConn := <-connections: // 从连接池获取连接
			if wrapConn == nil {
				return nil, ErrClosed
			}

			// 连接空闲超时
			if timeout := pool.idleTimeout; timeout > 0 {
				if wrapConn.t.Add(timeout).Before(time.Now()) {
					_ = pool.Close(wrapConn.conn)
					continue
				}
			}

			// 连接是否可用
			if err := pool.ping(wrapConn.conn); err != nil {
				_ = pool.Close(wrapConn.conn)
				continue
			}

			return wrapConn.conn, nil

		default: // 创建新连接
			pool.mu.Lock()

			// 连接数是否超时最大活跃数
			if pool.openConnections >= pool.maxActive {
				pool.mu.Unlock()
				continue
			}

			// 创建新的连接
			conn, err := pool.factory()
			if err != nil {
				pool.mu.Unlock()
				return nil, err
			}

			pool.openConnections++
			pool.mu.Unlock()

			return conn, nil
		}
	}
}

// Put 将连接放回连接池
func (pool *amqpConnectionPool) Put(conn *amqp.Connection) error {
	if conn == nil {
		return errors.New("connection is nil.")
	}

	pool.mu.Lock()

	if pool.connections == nil {
		pool.mu.Unlock()
		return pool.Close(conn)
	}

	select {
	case pool.connections <- &idleConnection{conn: conn, t: time.Now()}: // 放回连接池，并重置空闲时间
		pool.mu.Unlock()
		return nil
	default:
		pool.mu.Unlock()
		// 连接池已满，直接关闭该连接
		return pool.Close(conn)
	}
}

// Close 关闭连接
func (pool *amqpConnectionPool) Close(conn *amqp.Connection) error {
	if conn == nil {
		return errors.New("connection is nil.")
	}
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.openConnections--
	return pool.close(conn)
}

// Release 释放连接池中所有连接
func (pool *amqpConnectionPool) Release() {
	pool.mu.Lock()
	connections := pool.connections
	pool.connections = nil
	pool.mu.Unlock()

	if connections == nil {
		return
	}

	close(connections)
	for wrapConn := range connections {
		_ = pool.Close(wrapConn.conn)
	}
}

// Len 连接池连接数量
func (pool *amqpConnectionPool) Len() int {
	return len(pool.getConnections())
}
