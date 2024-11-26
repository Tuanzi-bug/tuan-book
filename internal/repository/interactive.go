package repository

import (
	"context"
	"errors"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/cache"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"github.com/Tuanzi-bug/tuan-book/pkg/log"
	"github.com/ecodeclub/ekit/slice"
	"go.uber.org/zap"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, id int64, uid int64) error
	DecrLike(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
}
type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
}

func (c *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	inters, err := c.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Interactive, domain.Interactive](inters, func(idx int, src dao.Interactive) domain.Interactive {
		return c.toDomain(src)
	})
	return res, nil
}

func (c *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	err := c.dao.BatchIncrReadCnt(ctx, bizs, bizIds)
	if err != nil {
		return err
	}
	//// 添加缓存方案
	//go func() {
	//	for i, biz := range bizs {
	//		er := c.cache.IncrReadCntIfPresent(ctx, biz, bizIds[i])
	//		if er != nil {
	//			log.Error("cache IncrReadCntIfPresent failed", zap.Error(er), zap.String("biz", biz), zap.Int64("id", bizIds[i]))
	//		}
	//	}
	//}()
	return nil
}

func (c *CachedInteractiveRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, id, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrRecordNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	// 获取文章的交互信息
	// 高频数据需要从缓存中获取，如果缓存中不存在，需要从数据库中获取
	// 这里面会出现数据不一致的问题，但是阅读数，点赞数，收藏数的数据不一致是可以接受的,不用保证强一致性
	intr, err := c.cache.Get(ctx, biz, id)
	if err == nil {
		return intr, nil
	}
	ie, err := c.dao.Get(ctx, biz, id)
	if err != nil {
		return domain.Interactive{}, err
	}
	res := c.toDomain(ie)
	// 回写缓存
	go func() {
		er := c.cache.Set(ctx, biz, id, res)
		if er != nil {
			// 记录日志，不影响主流程
			log.Error("cache Set failed", zap.Error(er), zap.String("biz", biz), zap.Int64("id", id))
		}
	}()
	return res, err

}

func (c *CachedInteractiveRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, id, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrRecordNotFound):
		return false, nil
	default:
		return false, err
	}
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
			log.Error("cache IncrCollectCntIfPresent failed", zap.Error(er),
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
			log.Error("cache IncrLikeCntIfPresent failed", zap.Error(er), zap.String("biz", biz), zap.Int64("id", id), zap.Int64("uid", uid))
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
			log.Error("cache DecrLikeCntIfPresent failed", zap.Error(er), zap.String("biz", biz), zap.Int64("id", id), zap.Int64("uid", uid))
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
			log.Error("cache IncrReadCntIfPresent failed", zap.Error(er), zap.String("biz", biz), zap.Int64("bizId", bizId))
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
func (c *CachedInteractiveRepository) toDomain(ie dao.Interactive) domain.Interactive {
	return domain.Interactive{
		ReadCnt:    ie.ReadCnt,
		LikeCnt:    ie.LikeCnt,
		CollectCnt: ie.CollectCnt,
	}
}
