package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
}

type ArticleStatus uint8

func (a ArticleStatus) ToUint8() uint8 {
	return uint8(a)
}

type Author struct {
	Id int64
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
