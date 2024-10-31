package service

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	events "github.com/Tuanzi-bug/tuan-book/internal/events/article"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/Tuanzi-bug/tuan-book/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//go:generate mockgen -source=./article.go -package=svcmocks -destination=./mocks/article.mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid, id int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx *gin.Context, id int64) (domain.Article, error)
	GetPubById(ctx *gin.Context, id, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer events.Producer
}

func NewArticleService(repo repository.ArticleRepository, events events.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: events,
	}
}

func (s *articleService) Withdraw(ctx context.Context, uid, id int64) error {
	return s.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func (s *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := s.repo.Update(ctx, art)
		return art.Id, err
	}
	return s.repo.Create(ctx, art)
}

func (s *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	return s.repo.Sync(ctx, article)
}

func (s *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return s.repo.GetByAuthor(ctx, uid, offset, limit)
}

func (s *articleService) GetById(ctx *gin.Context, id int64) (domain.Article, error) {
	return s.repo.GetById(ctx, id)
}

func (s *articleService) GetPubById(ctx *gin.Context, id, uid int64) (domain.Article, error) {
	res, err := s.repo.GetPubById(ctx, id)
	go func() {
		if err == nil {
			er := s.producer.ProduceReadEvent(events.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				log.Error("produce read event failed", zap.Int64("aid", id), zap.Error(er))
			}
		}
	}()
	return res, err
}
