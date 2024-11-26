package service

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
)

//go:generate mockgen -source=./ranking.go -package=svcmocks -destination=./mocks/ranking.mock.go RankingService
type RankingService interface {
	// TopN 前 100 的
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	articleSvc     ArticleService
	interactiveSvc InteractiveService
	rankingRepo    repository.RankingRepository
	scoreFunc      func(likeCnt int64, utime time.Time) float64
	n              int // 容量
	batchSize      int
}

func NewBatchRankingService(intrSvc InteractiveService, artSvc ArticleService, rankingRepository repository.RankingRepository) RankingService {
	return &BatchRankingService{
		articleSvc:     artSvc,
		interactiveSvc: intrSvc,
		n:              100,
		batchSize:      100,
		rankingRepo:    rankingRepository,
		// 计算热榜分数
		scoreFunc: func(likeCnt int64, utime time.Time) float64 {
			duration := time.Since(utime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
	}
}

func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	// 保存到缓存
	return b.rankingRepo.ReplaceTopN(ctx, arts)
}

func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	ddl := start.Add(-time.Hour * 24 * 7) // 一周内的文章
	// 定义一个结构体，用于存储分数和文章
	type Score struct {
		score float64
		art   domain.Article
	}
	// 构建一个小顶堆
	topN := queue.NewPriorityQueue[Score](b.n, func(src Score, dst Score) int {
		if src.score > dst.score {
			return 1
		} else if src.score < dst.score {
			return -1
		}
		return 0
	})
	for {
		// 取数据
		articles, err := b.articleSvc.ListPub(ctx, start, offset, b.batchSize)
		if err != nil {
			return nil, err
		}
		// 取出相关信息 (包括：点赞数，时间)
		ids := slice.Map(articles, func(idx int, src domain.Article) int64 {
			return src.Id
		})
		intrMap, err := b.interactiveSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
		// 计算分数
		for _, art := range articles {
			// 点赞信息
			interactive := intrMap[art.Id]
			score := b.scoreFunc(interactive.LikeCnt, art.Utime)
			ele := Score{
				score: score,
				art:   art,
			}
			// 如果堆未满，直接插入
			if topN.Len() < b.n {
				_ = topN.Enqueue(ele)
			} else {
				// 如果堆满了，比较堆顶元素
				minEle, _ := topN.Dequeue()
				// 如果分数大于堆顶元素，替换
				if minEle.score < score {
					_ = topN.Enqueue(ele)
				} else {
					// 如果分数小于堆顶元素，重新插入
					_ = topN.Enqueue(minEle)
				}
			}
		}
		// 往后取数据
		offset = offset + len(articles)
		// 如果数据不足或者 超过时间限制（文章太古老了） ，直接返回
		if len(articles) < b.batchSize || articles[len(articles)-1].Utime.Before(ddl) {
			break
		}
	}
	// 返回最终结果
	res := make([]domain.Article, topN.Len())
	for i := topN.Len() - 1; i >= 0; i-- {
		ele, _ := topN.Dequeue()
		res[i] = ele.art
	}
	return res, nil
}

func (b *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return b.rankingRepo.GetTopN(ctx)
}
