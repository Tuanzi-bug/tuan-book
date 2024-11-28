package job

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	"github.com/Tuanzi-bug/tuan-book/pkg/log"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
)

/*
目前多个节点都会进行计算，目前保证一个时刻只有一个节点进行计算。
使用分布式锁，保证只有一个节点进行计算。
*/

type RankingJob struct {
	svc       service.RankingService
	timeout   time.Duration
	client    *rlock.Client
	lock      *rlock.Lock
	localLock *sync.Mutex
	key       string
}

func (r *RankingJob) Name() string {
	return "RankingJob"
}

func (r *RankingJob) Run() error {
	lock := r.lock
	// 为了保证其他实例不会去计算热榜，扩大加锁的范围，保证只有一个节点进行计算
	if lock == nil {
		// 首先尝试去获取分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		// 设置重试策略：重试3次，每次间隔100微秒
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Microsecond * 100,
			Max:      3,
		}, time.Second)
		if err != nil {
			r.localLock.Unlock()
			log.Warn("获取分布式锁失败", log.Err(err))
			return nil
		}
		// 拿到分布式锁
		r.lock = lock
		r.localLock.Unlock()
		go func() {
			// 考虑续约机制
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

// Close 考虑关闭的时候，释放锁
func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

func NewRankingJob(svc service.RankingService, timeout time.Duration, client *rlock.Client) Job {
	return &RankingJob{
		svc:       svc,
		key:       "rlock:corn:ranking-job",
		timeout:   timeout,
		client:    client,
		localLock: &sync.Mutex{},
	}
}
