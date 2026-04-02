package database

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"sync"
)

var (
	r         *redis.Redis
	onceRedis sync.Once
)

func InitRedis(cfg redis.RedisConf) *redis.Redis {
	onceRedis.Do(func() {
		r = redis.MustNewRedis(cfg)
	})
	return r
}
