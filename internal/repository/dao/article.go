package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Article struct {
	Id      int64  `gorm:"primaryKey,autoIncrement"`
	Title   string `gorm:"type=varchar(4096)"`
	Content string `gorm:"type=BLOB"`
	// 我要根据创作者ID来查询
	AuthorId int64 `gorm:"index"`
	Ctime    int64
	// 更新时间
	Utime int64
}

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	Sync(ctx context.Context, entity Article) (int64, error)
}

type GROMArticleDAO struct {
	db *gorm.DB
}

func NewGROMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GROMArticleDAO{db: db}
}

func (dao *GROMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (dao *GROMArticleDAO) UpdateById(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	article.Utime = now
	res := dao.db.WithContext(ctx).Model(&article).Where("id=? and author_id=?", article.Id, article.AuthorId).Updates(map[string]interface{}{
		"Title":   article.Title,
		"Content": article.Content,
		"Utime":   article.Utime,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("id 不正确 或者 创作者不正确")
	}
	return nil
}

func (dao *GROMArticleDAO) Sync(ctx context.Context, article Article) (int64, error) {
	var id = article.Id
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		dao := NewGROMArticleDAO(tx)
		// 先对制作库进行更新
		if id > 0 {
			err = dao.UpdateById(ctx, article)
		} else {
			id, err = dao.Insert(ctx, article)
		}
		if err != nil {
			return err
		}
		article.Id = id
		now := time.Now().UnixMilli()
		pubArt := PublishedArticle(article)
		pubArt.Ctime = now
		pubArt.Utime = now
		// 接下来对线上库进行更新。不存在就创建，存在就修改部分值
		err = tx.Clauses(
			clause.OnConflict{
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"title":   pubArt.Title,
					"content": pubArt.Content,
					"utime":   now,
				}),
			}).Create(&pubArt).Error
		return err
	})
	return id, err
}

type PublishedArticle Article