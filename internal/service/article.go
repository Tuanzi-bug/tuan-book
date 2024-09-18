package service

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid, id int64) error
}

type articleService struct {
	repo repository.ArticleRepository
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
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
