package job

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	"time"
)

type RankingJob struct {
	svc     service.RankingService
	timeout time.Duration
}

func (r *RankingJob) Name() string {
	return "RankingJob"
}

func (r *RankingJob) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func NewRankingJob(svc service.RankingService, timeout time.Duration) Job {
	return &RankingJob{
		svc:     svc,
		timeout: timeout,
	}
}
