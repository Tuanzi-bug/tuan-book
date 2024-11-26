package cache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/redis/go-redis/v9"
	"time"
)

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RankingLocalCache struct {
	// 原子操作保证并发安全
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

func (r *RankingLocalCache) Set(ctx context.Context, arts []domain.Article) error {
	r.topN.Store(arts)
	r.ddl.Store(time.Now().Add(r.expiration))
	return nil
}

func (r *RankingLocalCache) Get(ctx context.Context) ([]domain.Article, error) {
	ddl := r.ddl.Load()
	arts := r.topN.Load()
	if ddl.Before(time.Now()) || len(arts) == 0 {
		return nil, errors.New("本地缓存失效了")
	}
	return arts, nil
}

func NewRankingLocalCache() RankingCache {
	return &RankingLocalCache{
		topN:       atomicx.NewValue[[]domain.Article](),
		ddl:        atomicx.NewValue[time.Time](),
		expiration: time.Hour * 24,
	}
}

type RankingRedisCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
}

func (r *RankingRedisCache) Set(ctx context.Context, arts []domain.Article) error {
	// 存储信息不用全部存储，只存储部分信息（title和abstract）
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	// 序列化
	data, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key, data, r.expiration).Err()
}

func (r *RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	artsBytes, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(artsBytes, &arts)
	if err != nil {
		return nil, err
	}
	return arts, nil
}

func NewRankingRedisCache(client redis.Cmdable) RankingCache {
	return &RankingRedisCache{
		client:     client,
		key:        "ranking:topN",
		expiration: time.Hour * 24, // 设置一天的过期时间
	}
}
