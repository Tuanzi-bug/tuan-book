package service

import (
	"context"
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/Tuanzi-bug/tuan-book/internal/service/sms"
	"math/rand"
)

const codeTplId = "1877556"

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type VerifyCodeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &VerifyCodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *VerifyCodeService) Send(ctx context.Context, biz, phone string) error {
	code := svc.generateCode()
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送出去
	return svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
}

func (svc *VerifyCodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *VerifyCodeService) generateCode() string {
	// 生成随机6位数，不足补0
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
