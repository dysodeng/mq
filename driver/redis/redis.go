package redis

import "github.com/dysoodeng/mq/message"

type Redis struct {
	config Config
	key message.Key
}

type Config struct {

}

func NewRedis(key message.Key, config Config) *Redis {
	return nil
}
