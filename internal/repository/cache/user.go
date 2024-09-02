package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}

type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Minute * 15, // 暂时进行固定，不设置配置
	}
}

func (c *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	// user:info:id
	key := c.key(id)
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (c *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	key := c.key(u.Id)
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, val, c.expiration).Err()
}

func (c *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
