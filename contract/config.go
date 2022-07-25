package contract

import "time"

// Config 配置
type Config interface {
	String() string
	PoolConfig() Pool
}

type Pool struct {
	MinConn     int
	MaxConn     int
	MaxIdleConn int
	IdleTimeout time.Duration
}
