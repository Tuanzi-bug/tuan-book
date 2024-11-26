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
	// 状态
	Status uint8
	Ctime  int64
	// 更新时间
	Utime int64
}

type PublishedArticle Article

//go:generate mockgen -source=./article.go -package=daomocks -destination=./mocks/article.mock.go ArticleDAO
type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	Sync(ctx context.Context, entity Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error)
}

type GROMArticleDAO struct {
	db *gorm.DB
}

func (dao *GROMArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) {
	var arts []PublishedArticle
	const ArticleStatusPublished = 2
	err := dao.db.WithContext(ctx).Where("utime < ? and status = ?", start.UnixMilli(), ArticleStatusPublished).Offset(offset).Limit(limit).Find(&arts).Error
	return arts, err
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
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
		"status":  article.Status,
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
		dao := NewGORMArticleDAO(tx)
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
					"status":  article.Status,
				}),
			}).Create(&pubArt).Error
		return err
	})
	return id, err
}

func (dao *GROMArticleDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先修改制作库的状态
		res := tx.Model(&Article{}).Where("id=? and author_id=?", id, uid).Updates(map[string]any{
			"utime":  now,
			"status": status,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("ID 不对或者创作者不对")
		}
		// 再修改线上库的状态
		return tx.Model(&PublishedArticle{}).Where("id=? and author_id=?", id, uid).Updates(map[string]any{
			"utime":  now,
			"status": status,
		}).Error
	})
}

func (dao *GROMArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).
		Where("author_id = ?", uid).
		Offset(offset).Limit(limit).
		// a ASC, B DESC
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (dao *GROMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&art).Error
	return art, err
}

func (dao *GROMArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var art PublishedArticle
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&art).Error
	return art, err
}
