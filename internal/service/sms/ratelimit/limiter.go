package ratelimit

import (
	"context"
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/internal/service/sms"
	"github.com/Tuanzi-bug/tuan-book/pkg/limiter"
)

var errLimited = fmt.Errorf("触发了限流")

// RateLimitDecorator 采用装饰器模式，给第三方服务套一个限流器
type RateLimitDecorator struct {
	svc     sms.Service
	limiter limiter.Limiter
	key     string
}

func NewRateLimitDecorator(svc sms.Service, limiter limiter.Limiter) *RateLimitDecorator {
	return &RateLimitDecorator{
		svc:     svc,
		limiter: limiter,
		key:     "sms-limiter",
	}
}

func (r *RateLimitDecorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := r.limiter.Limit(ctx, r.key)
	if err != nil {
		// 系统错误
		// 可以限流：保守策略，你的下游很坑的时候，
		// 可以不限：你的下游很强，业务可用性要求很高，尽量容错策略
		return fmt.Errorf("短信服务限流出现问题，%w", err)
	}
	// 触发限流
	if limited {
		return errLimited
	}
	return r.svc.Send(ctx, tplId, args, numbers...)
}
