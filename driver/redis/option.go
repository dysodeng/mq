package redis

import (
	"fmt"
	"time"

	"github.com/dysodeng/mq/contract"
)

const defaultMaxConn = 1000
const defaultIdleTimeout = 1 * time.Minute

type Config struct {
	Addr     string
	Username string
	Password string
	DB       int
	Pool     *contract.Pool
}

func (config *Config) String() string {
	dsn := "redis://"
	if config.Username != "" && config.Password != "" {
		dsn += fmt.Sprintf("%s:%s@", config.Username, config.Password)
	} else if config.Password != "" {
		dsn += fmt.Sprintf(":%s@", config.Password)
	}
	return fmt.Sprintf("%s%s/%d", dsn, config.Addr, config.DB)
}

func (config *Config) PoolConfig() contract.Pool {
	if config.Pool != nil {
		return *config.Pool
	}
	return contract.Pool{
		MaxConn:     defaultMaxConn,
		IdleTimeout: defaultIdleTimeout,
	}
}
