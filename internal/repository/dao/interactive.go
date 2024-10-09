package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
}

type GROMInteractiveDAO struct {
	db *gorm.DB
}

func (G *GROMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	// 可能数据库中不存在，所以需要upsert语义
	// upsert语义：如果不存在则插入，如果存在则更新
	now := time.Now().UnixMilli()
	return G.db.WithContext(ctx).Clauses(clause.OnConflict{DoUpdates: clause.Assignments(map[string]interface{}{
		// read_cnt = read_cnt + 1
		"read_cnt": gorm.Expr("read_cnt + ?", 1),
		"utime":    now,
	})}).Create(&Interactive{
		BizId:   bizId,
		Biz:     biz,
		ReadCnt: 1,
		Utime:   now,
		Ctime:   now,
	}).Error
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GROMInteractiveDAO{db: db}

}

// Interactive 存储交互数据（阅读、点赞、收藏）表
type Interactive struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// <bizid, biz>
	BizId int64 `gorm:"uniqueIndex:biz_type_id"`
	// WHERE biz = ?
	// 默认请情况下是BLOB/TEXT类型，需要指定长度
	Biz string `gorm:"uniqueIndex:biz_type_id;type:varchar(128)"`

	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Utime      int64
	Ctime      int64
}
