package amqp

import (
	"fmt"

	"github.com/dysodeng/mq/contract"
)

// Config 配置
type Config struct {
	Host     string
	Username string
	Password string
	VHost    string
	Pool     *contract.Pool
}

func (config *Config) String() string {
	if config.VHost == "" {
		config.VHost = "/"
	}
	return fmt.Sprintf(
		"amqp://%s:%s@%s%s",
		config.Username,
		config.Password,
		config.Host,
		config.VHost,
	)
}

func (config *Config) PoolConfig() contract.Pool {
	if config.Pool != nil {
		return *config.Pool
	}
	return contract.Pool{
		MinConn:     defaultMinConn,
		MaxConn:     defaultMaxConn,
		MaxIdleConn: defaultMaxIdleConn,
		IdleTimeout: defaultIdleTimeout,
	}
}
