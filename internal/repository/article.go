package repository

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/cache"
	"github.com/Tuanzi-bug/tuan-book/internal/repository/dao"
	"github.com/Tuanzi-bug/tuan-book/pkg/log"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

//go:generate mockgen -source=./article.go -package=repomocks -destination=./mocks/article.mock.go ArticleRepository
type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx *gin.Context, id int64) (domain.Article, error)
	GetPubById(ctx *gin.Context, id int64) (domain.Article, error)
}

type CacheArticleRepository struct {
	dao   dao.ArticleDAO
	cache cache.ArticleCache

	// repository 存在一些缓存机制，不直接访问dao层
	userRepo UserRepository
}

func NewCacheArticleRepository(dao dao.ArticleDAO, articleCache cache.ArticleCache, userRepo UserRepository) ArticleRepository {
	return &CacheArticleRepository{
		dao:      dao,
		cache:    articleCache,
		userRepo: userRepo,
	}
}

func (repo *CacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := repo.dao.Insert(ctx, repo.toEntity(art))
	if err == nil {
		er := repo.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			log.Error("Create 删除缓存失败", zap.Error(er), zap.Int64("uid", art.Author.Id))
		}
	}
	return id, err
}

func (repo *CacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := repo.dao.UpdateById(ctx, repo.toEntity(art))
	if err == nil {
		er := repo.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			log.Error("Update 删除缓存失败", zap.Error(er), zap.Int64("uid", art.Author.Id))
		}
	}
	return err
}

func (repo *CacheArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := repo.dao.Sync(ctx, repo.toEntity(art))
	if err == nil {
		er := repo.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
			log.Error("Sync 删除缓存失败", zap.Error(er), zap.Int64("uid", art.Author.Id))
		}
	}
	// 当新帖子发布时候，就会被人访问，考虑当做缓存预热
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 优化点：缓存过期时间，可以根据粉丝数量进行调整
		user, er := repo.userRepo.FindById(ctx, art.Author.Id)
		if er != nil {
			log.Error("获取创作者信息，缓存失败", zap.Error(er), zap.Int64("uid", art.Author.Id))
			return
		}
		art.Author = domain.Author{
			Id:   art.Author.Id,
			Name: user.Nickname,
		}
		er = repo.cache.SetPub(ctx, art)
		if er != nil {
			log.Error("Sync 预热缓存失败", zap.Error(er), zap.Int64("id", art.Id))
		}
	}()

	return id, err
}

func (repo *CacheArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	err := repo.dao.SyncStatus(ctx, uid, id, status.ToUint8())
	if err == nil {
		er := repo.cache.DelFirstPage(ctx, uid)
		if er != nil {
			log.Error("SyncStatus 删除缓存失败", zap.Error(er), zap.Int64("uid", uid))
		}
	}
	return err
}

func (repo *CacheArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 在这里集成复杂的缓存策略
	// 第一页 limit<=100都可以查询缓存
	if offset == 0 && limit == 100 {
		res, err := repo.cache.GetFirstPage(ctx, uid)
		if err != nil {
			// 如果缓存出现了问题，进行记录日志
			log.Error("缓存未命中", zap.Error(err), zap.Int64("uid", uid))
		} else {
			return res, nil
		}
	}
	arts, err := repo.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return repo.toDomain(src)
	})
	// 第一页数据进行缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if offset == 0 && limit == 100 {
			err := repo.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				log.Error("缓存写入失败", zap.Error(err), zap.Int64("uid", uid))
			}
		}
	}()
	// 预测：你会点击第一个文章的详细信息
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.preCache(ctx, res)
	}()

	return res, nil
}

func (repo *CacheArticleRepository) GetById(ctx *gin.Context, id int64) (domain.Article, error) {
	// 先查询缓存内的数据
	res, err := repo.cache.Get(ctx, id)
	if err == nil {
		return res, err
	} // 缓存未命中还是错误其实都是没有关系的
	art, err := repo.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = repo.toDomain(art)
	// 缓存数据
	go func() {

		err := repo.cache.Set(ctx, res)
		if err != nil {
			log.Error("缓存写入失败", zap.Error(err), zap.Int64("id", id))
		}
	}()
	return res, nil
}

func (repo *CacheArticleRepository) GetPubById(ctx *gin.Context, id int64) (domain.Article, error) {
	// 预测：新帖子发布时候，就会被人访问，考虑当做缓存预热
	res, err := repo.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	}
	art, err := repo.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = repo.toDomain(dao.Article(art))
	// 还需要获取创作者的信息
	author, err := repo.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
	}
	res.Author.Name = author.Nickname
	// 缓存数据
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := repo.cache.SetPub(ctx, res)
		if er != nil {
			log.Error("GetPubById 缓存写入失败", zap.Error(er), zap.Int64("id", id))
		}
	}()

	return res, err
}

func (repo *CacheArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	// 预测：缓存第一个文章的详细信息
	// 优化点1：不缓存大对象：文章内容过大则不缓存
	const contentSizeThreshold = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < contentSizeThreshold {
		err := repo.cache.Set(ctx, arts[0])
		if err != nil {
			log.Error("preCache 缓存第一个文章数据写入失败", zap.Error(err), zap.Int64("id", arts[0].Id))
		}
	}
}

func (repo *CacheArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.Id,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

func (repo *CacheArticleRepository) toEntity(article domain.Article) dao.Article {
	return dao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   article.Status.ToUint8(),
	}
}
