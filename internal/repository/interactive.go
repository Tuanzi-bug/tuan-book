package repository

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/cache"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"go.uber.org/zap"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
}

func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	// 阅读数获取是一个高频访问数据，需要缓存，防止超高的访问压力，但是数据不一致问题在该场景下可以接受的
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	go func() {
		er := c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
		if er != nil {
			// 记录日志，不影响主流程
			zap.L().Error("cache IncrReadCntIfPresent failed", zap.Error(er), zap.String("biz", biz), zap.Int64("bizId", bizId))
		}
	}()
	return nil
}

func NewCachedInteractiveRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
	}
}
