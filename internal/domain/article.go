package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
	Ctime   time.Time
	Utime   time.Time
}

// Abstract 返回文章的摘要，取前128字节返回
// todo 可以让ai去生成摘要
func (a Article) Abstract() string {
	str := []rune(a.Content)
	if len(str) > 128 {
		str = str[:128]
	}
	return string(str)
}

type ArticleStatus uint8

func (a ArticleStatus) ToUint8() uint8 {
	return uint8(a)
}

type Author struct {
	Id   int64
	Name string
}

const (
	// ArticleStatusUnknown 这是一个未知状态
	ArticleStatusUnknown = iota
	// ArticleStatusUnpublished 未发表
	ArticleStatusUnpublished
	// ArticleStatusPublished 已发表
	ArticleStatusPublished
	// ArticleStatusPrivate 仅自己可见
	ArticleStatusPrivate
)
