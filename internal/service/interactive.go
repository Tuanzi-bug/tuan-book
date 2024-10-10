package service

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/gin-gonic/gin"
)

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	Collect(ctx *gin.Context, biz string, bizId int64, cid int64, uid int64) error
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func (i *interactiveService) Collect(ctx *gin.Context, biz string, bizId int64, cid int64, uid int64) error {
	return i.repo.AddCollectItem(ctx, biz, bizId, cid, uid)
}

func (i *interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.IncrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{
		repo: repo,
	}
}
