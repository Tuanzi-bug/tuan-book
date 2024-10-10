package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
)

const fieldReadCnt = "read_cnt"
const fieldLikeCnt = "like_cnt"

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (i *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return i.client.Eval(ctx, luaIncrCnt, []string{i.key(biz, bizId)}, fieldLikeCnt, 1).Err()
}

func (i *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return i.client.Eval(ctx, luaIncrCnt, []string{i.key(biz, bizId)}, fieldLikeCnt, -1).Err()
}

func (i *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	/*
				分析缓存存在的情况：
				1. 缓存中可能不存在key以及字段，因为过期或者没有设置
				2. 缓存中可能存在key，但是字段不存在
				解决方法：
				可以通过HIncrBy命令，如果字段不存在则创建，否则自增
			其实存在一个冲突的点：
				假设key不存在，自增后的值与数据库中的值不一致，到时候可以通过回查时候进行解决
			最后的解决方案：
			1. 判断key是否存在
			2. 如果key存在，就把 read_cnt 对应的值自增
		需要lua脚本来保证原子性
	*/
	return i.client.Eval(ctx, luaIncrCnt, []string{i.key(biz, bizId)}, fieldReadCnt, 1).Err()
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{
		client: client,
	}
}
func (i *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)

}
