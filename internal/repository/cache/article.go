package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:generate mockgen -source=./article.go -package=cachemocks -destination=./mocks/article.mock.go ArticleCache
type ArticleCache interface {
	// 处理极小的领域：缓存列表的第一页的数据
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, res []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error
	// 进行业务预加载,相当于对于某些场景进行预测，预加载列表第一个详细数据
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, art domain.Article) error
}

type ArticleRedisCache struct {
	client redis.Cmdable
}

func NewArticleRedisCache(client redis.Cmdable) ArticleCache {
	return &ArticleRedisCache{client: client}
}

// GetFirstPage 获取创作者的缓存的首页文章
func (a *ArticleRedisCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	key := a.firstKey(uid)
	val, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var articles []domain.Article
	err = json.Unmarshal(val, &articles)
	return articles, err
}

// SetFirstPage 缓存创作者的首页文章
func (a *ArticleRedisCache) SetFirstPage(ctx context.Context, uid int64, articles []domain.Article) error {
	for i := 0; i < len(articles); i++ {
		articles[i].Content = articles[i].Abstract()
	}
	key := a.firstKey(uid)
	val, err := json.Marshal(articles)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, key, val, time.Minute*10).Err()
}

func (a *ArticleRedisCache) DelFirstPage(ctx context.Context, uid int64) error {
	return a.client.Del(ctx, a.firstKey(uid)).Err()
}

func (a *ArticleRedisCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.key(id)).Result()
	if err != nil {
		return domain.Article{}, err
	}
	var article domain.Article
	err = json.Unmarshal([]byte(val), &article)
	return article, err
}

func (a *ArticleRedisCache) Set(ctx context.Context, art domain.Article) error {
	// 缓存制作文章
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	// 过期时间设置短：防止预测失败对业务造成内存开销
	return a.client.Set(ctx, a.key(art.Id), val, time.Minute*10).Err()
}

func (a *ArticleRedisCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.pubKey(id)).Result()
	if err != nil {
		return domain.Article{}, err
	}
	var article domain.Article
	err = json.Unmarshal([]byte(val), &article)
	return article, err
}

func (a *ArticleRedisCache) SetPub(ctx context.Context, art domain.Article) error {
	// 缓存发布文章
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	// 过期时间设置短：防止预测失败对业务造成内存开销
	return a.client.Set(ctx, a.pubKey(art.Id), val, time.Minute*10).Err()
}

func (a *ArticleRedisCache) firstKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}

func (a *ArticleRedisCache) key(id int64) string {
	return fmt.Sprintf("article:detail:%d", id)
}

func (a *ArticleRedisCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}
