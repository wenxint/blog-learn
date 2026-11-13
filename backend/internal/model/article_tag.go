package model

// ArticleTag 对应中间表 article_tags
//
// 这是一张「纯关联表」，没有自己的业务字段，
// 所以用 (article_id, tag_id) 联合主键，不需要自增 id。
//
// 注意：Article model 里的 Tags 字段仍然用 many2many 声明，
// 那是给 Preload 读取用的；写入关联我们手写这张表，避免 GORM 的级联 upsert。
type ArticleTag struct {
	ArticleID uint64 `gorm:"primaryKey" json:"article_id"`
	TagID     uint64 `gorm:"primaryKey" json:"tag_id"`
}

func (ArticleTag) TableName() string { return "article_tags" }
