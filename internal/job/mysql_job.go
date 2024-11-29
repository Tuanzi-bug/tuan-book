package job

import (
	"context"
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	"github.com/Tuanzi-bug/tuan-book/pkg/log"
	"golang.org/x/sync/semaphore"
	"time"
)

type Executor interface {
	Name() string
	Exec(ctx context.Context, j domain.Job) error
}

// LocalFuncExecutor 调用本地的执行器
type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未注册本地方法 %s", j.Name)
	}
	return fn(ctx, j)
}

// RegisterFunc 注册执行器
func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: make(map[string]func(ctx context.Context, j domain.Job) error)}
}

type Scheduler struct {
	dbTimeout time.Duration
	svc       service.CronJobService
	executors map[string]Executor

	// 信号量，限制并发执行的任务数量
	limiter *semaphore.Weighted
}

func NewScheduler(svc service.CronJobService) *Scheduler {
	return &Scheduler{
		dbTimeout: time.Second,
		svc:       svc,
		executors: make(map[string]Executor),
		limiter:   semaphore.NewWeighted(100),
	}
}

func (s *Scheduler) RegisterExecutor(executor Executor) {
	s.executors[executor.Name()] = executor
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		// 从数据库中获取一个任务
		j, err := s.svc.Preempt(dbCtx)
		//cancel()
		if err != nil {
			continue
		}
		// 调度执行
		executor, ok := s.executors[j.Executor]
		if !ok {
			log.Error("未找到执行器", log.String("executor", j.Executor))
			continue
		}
		go func() {
			defer func() {
				s.limiter.Release(1)
				// 释放资源
				j.CancelFunc()
			}()
			// 执行任务
			er := executor.Exec(ctx, j)
			log.Debug("执行任务", log.String("name", j.Name), log.Err(er))
			if er != nil {
				log.Error("执行任务失败", log.Err(er))
				return
			}
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			// 执行成功，更新任务状态
			er = s.svc.ResetNextTime(ctx, j)
			if er != nil {
				log.Error("更新任务状态失败", log.Err(er))
			}
		}()
	}
}
