package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	Get(ctx context.Context, biz string, id int64) (Interactive, error)
	GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error)
	GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error)
}

type GROMInteractiveDAO struct {
	db *gorm.DB
}

func (G *GROMInteractiveDAO) GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	var ucb UserCollectionBiz
	err := G.db.WithContext(ctx).Where("uid = ? AND biz_id = ? AND biz = ?", uid, id, biz).First(&ucb).Error
	return ucb, err
}

func (G *GROMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	var ulb UserLikeBiz
	err := G.db.WithContext(ctx).Where("biz = ? AND biz_id = ? AND uid = ? AND status = ?", biz, id, uid, 1).First(&ulb).Error
	return ulb, err
}

func (G *GROMInteractiveDAO) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	var intr Interactive
	err := G.db.WithContext(ctx).Where("biz = ? AND biz_id = ?", biz, id).First(&intr).Error
	return intr, err
}

func (G *GROMInteractiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	// 这里的行为和点赞功能一致
	now := time.Now().UnixMilli()
	return G.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&cb).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{DoUpdates: clause.Assignments(map[string]interface{}{
			// collect_cnt = collect_cnt + 1
			"collect_cnt": gorm.Expr("collect_cnt + ?", 1),
			"utime":       now,
		})}).Create(&Interactive{
			BizId:      cb.BizId,
			Biz:        cb.Biz,
			CollectCnt: 1,
			Utime:      now,
			Ctime:      now,
		}).Error
	})
}

func (G *GROMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	// 对点赞数据的删除，如果真实删除会导致磁盘有很多空洞影响性能
	// 同时希望保留用户的点赞记录，所以采用逻辑删除
	// 需要修改两个表：总数据表和用户点赞表，需要使用事务保证数据一致性
	now := time.Now().UnixMilli()
	return G.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新用户点赞表
		err := tx.Model(&UserLikeBiz{}).Where("uid = ? AND biz_id = ? AND biz = ?", uid, id, biz).Updates(map[string]interface{}{
			"status": 0,
			"utime":  now,
		}).Error
		if err != nil {
			return err
		}
		// 更新总数据表
		return tx.Model(&Interactive{}).Where("biz_id = ? AND biz = ?", id, biz).Updates(map[string]interface{}{
			"like_cnt": gorm.Expr("like_cnt - ?", 1),
			"utime":    now,
		}).Error
	})
}

func (G *GROMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	// 点赞数据可能不存在，所以需要考虑upsert语义
	// 需要修改两个表：总数据表和用户点赞表，需要使用事务保证数据一致性
	now := time.Now().UnixMilli()
	return G.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新用户点赞表
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			Uid:    uid,
			BizId:  id,
			Biz:    biz,
			Status: 1,
			Utime:  now,
			Ctime:  now,
		}).Error
		if err != nil {
			return err
		}
		// 更新总数据表
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				// like_cnt = like_cnt + 1
				"like_cnt": gorm.Expr("like_cnt + ?", 1),
				"utime":    now,
			})}).Create(&Interactive{
			BizId:   id,
			Biz:     biz,
			LikeCnt: 1,
			Utime:   now,
			Ctime:   now,
		}).Error
	})
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

// UserLikeBiz 用户点赞业务表
type UserLikeBiz struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Uid   int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	// 0：未点赞 1：已点赞
	Status int
	Utime  int64
	Ctime  int64
}

type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 这边还是保留了了唯一索引
	Uid   int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	// 收藏夹的ID
	// 收藏夹ID本身有索引
	Cid   int64 `gorm:"index"`
	Utime int64
	Ctime int64
}
