package repository

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/cache"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"go.uber.org/zap"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, id int64, uid int64) error
	DecrLike(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
}

func (c *CachedInteractiveRepository) AddCollectItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Biz:   biz,
		BizId: id,
		Cid:   cid,
		Uid:   uid,
	})
	go func() {
		er := c.cache.IncrCollectCntIfPresent(ctx, biz, id)
		if er != nil {
			// 记录日志，不影响主流程
			zap.L().Error("cache IncrCollectCntIfPresent failed", zap.Error(er),
				zap.String("biz", biz),
				zap.Int64("id", id),
				zap.Int64("uid", uid),
				zap.Int64("cid", cid))
		}
	}()
	return err
}

func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, id int64, uid int64) error {
	// 点赞是一个高频的访问数据，需要考虑缓存方案
	err := c.dao.InsertLikeInfo(ctx, biz, id, uid)
	go func() {
		er := c.cache.IncrLikeCntIfPresent(ctx, biz, id)
		if er != nil {
			// 记录日志，不影响主流程
			zap.L().Error("cache IncrLikeCntIfPresent failed", zap.Error(er), zap.String("biz", biz), zap.Int64("id", id), zap.Int64("uid", uid))
		}
	}()
	return err
}

func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, id, uid)
	go func() {
		er := c.cache.DecrLikeCntIfPresent(ctx, biz, id)
		if er != nil {
			// 记录日志，不影响主流程
			zap.L().Error("cache DecrLikeCntIfPresent failed", zap.Error(er), zap.String("biz", biz), zap.Int64("id", id), zap.Int64("uid", uid))
		}
	}()
	return err
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
