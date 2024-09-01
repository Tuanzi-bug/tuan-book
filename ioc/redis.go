package ioc

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: config.Config.Redis.Password,
	})
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	return redisClient
}
