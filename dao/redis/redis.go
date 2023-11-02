package redis

import (
	"context"
	"douyin/config"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func Init(conf *config.RedisConfig) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.DB,
	})
	return rdb.Ping(context.Background()).Err()
}

func Close() {
	rdb.Close()
}
