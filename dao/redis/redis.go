package redis

import (
	"fmt"

	"github.com/chuxin0816/Scaffold/config"
	"github.com/go-redis/redis"
)

var RDB *redis.Client

func Init(conf *config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.DB,
	})
	_, err := RDB.Ping().Result()
	return err
}
