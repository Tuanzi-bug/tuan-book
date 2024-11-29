package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

const (
	// jobStatusWaiting 没人抢
	jobStatusWaiting = iota
	// jobStatusRunning 已经被人抢了
	jobStatusRunning
	// jobStatusPaused 不再需要调度了
	jobStatusPaused
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, t time.Time) error
	Stop(ctx context.Context, id int64) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func (G *GORMJobDAO) Stop(ctx context.Context, id int64) error {
	return G.db.WithContext(ctx).
		Where("id = ?", id).Updates(map[string]any{
		"status": jobStatusPaused,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (G *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := G.db.WithContext(ctx)
	for {
		var j Job
		now := time.Now().UnixMilli()
		// 首先先查是否存在未被抢占的任务
		err := db.Where("status=? and next_time<=?", jobStatusWaiting, now).First(&j).Error
		if err != nil {
			return Job{}, err
		}
		// 尝试抢占
		res := db.Model(&Job{}).Where("id=? and version=?", j.Id, j.Version).Updates(map[string]interface{}{
			"status":  jobStatusRunning,
			"version": j.Version + 1,
			"utime":   now,
		})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 说明已经被人抢占了
			continue
		}
		return j, nil
	}
}

func (G *GORMJobDAO) Release(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return G.db.WithContext(ctx).Model(&Job{}).Where("id=?", jid).Updates(map[string]interface{}{
		"status": jobStatusWaiting,
		"utime":  now,
	}).Error
}

func (G *GORMJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return G.db.WithContext(ctx).Model(&Job{}).Where("id=?", id).Updates(map[string]interface{}{
		"utime": now,
	}).Error
}

func (G *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, t time.Time) error {
	now := time.Now().UnixMilli()
	return G.db.WithContext(ctx).Model(&Job{}).Where("id=?", id).Updates(map[string]interface{}{
		"utime":     now,
		"next_time": t.UnixMilli(),
	}).Error
}

func NewGORMJobDAO(db *gorm.DB) JobDAO {
	return &GORMJobDAO{db: db}
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(128);unique"`
	Executor   string
	Expression string
	Cfg        string
	// 状态来表达，是不是可以抢占，有没有被人抢占
	Status int

	Version int

	NextTime int64 `gorm:"index"`

	Utime int64
	Ctime int64
}
