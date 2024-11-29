package service

import (
	"context"
	"github.com/Tuanzi-bug/tuan-book/internal/domain"
	"github.com/Tuanzi-bug/tuan-book/internal/repository"
	"github.com/Tuanzi-bug/tuan-book/pkg/log"
	"time"
)

type CronJobService interface {
	// 从数据库中获取下一个要执行的任务
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
	//Release(ctx context.Context, job domain.Job) error
	// 暴露 job 的增删改查方法
}

type cronJobService struct {
	repo            repository.CronJobRepository
	refreshInterval time.Duration
}

func (c *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	log.Debug("获取抢占任务", log.Int64("id", j.Id))
	// 续约机制
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.Id)
		}
	}()
	// 取消函数：释放任务+停止续约
	j.CancelFunc = func() {
		ticker.Stop()
		log.Info("释放任务", log.Int64("id", j.Id))
		ctx, _ = context.WithTimeout(context.Background(), time.Second*10)
		//defer cancel()
		err = c.repo.Release(ctx, j.Id)
		if err != nil {
			log.Error("释放任务失败", log.Err(err), log.Int64("id", j.Id))
		}
	}
	return j, nil
}

func (c *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	if nextTime.IsZero() {
		return c.repo.Stop(ctx, j.Id)
	}
	return c.repo.UpdateNextTime(ctx, j.Id, nextTime)
}

func NewCronJobService(repo repository.CronJobRepository) CronJobService {
	return &cronJobService{repo: repo, refreshInterval: time.Minute}
}

func (c *cronJobService) refresh(id int64) {
	// 本质上就是更新一下更新时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		log.Error("更新更新时间失败", log.Err(err), log.Int64("id", id))
	}
}
