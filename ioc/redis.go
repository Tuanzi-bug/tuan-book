package ioc

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/config"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	type redisConfig struct {
		Addr     string
		Password string
	}
	var cfg = redisConfig{
		Addr:     "192.168.1.3:6379",
		Password: "123456",
	}
	err := viper.UnmarshalKey("redis", &cfg)
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: config.Config.Redis.Password,
	})
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	return redisClient
}

func InitRlockClient(client redis.Cmdable) *rlock.Client {
	return rlock.NewClient(client)
}
