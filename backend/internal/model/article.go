package model

import (
	"time"

	"gorm.io/gorm"
)

// Article 对应表 articles
//
// 字段顺序不影响映射，映射靠的是「字段名 -> 列名」规则：
//
//	AuthorID   -> author_id
//	IsTop      -> is_top
//	ViewCount  -> view_count
//	CreatedAt  -> created_at
//
// 如果推导不对，用 gorm:"column:xxx" 强制指定。
type Article struct {
	ID      uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Title   string `gorm:"size:200;not null" json:"title"`
	Summary string `gorm:"size:500;not null;default:''" json:"summary"`

	// 列表接口不返回正文字段时，内容为空字符串，omitempty 让它在 JSON 里消失
	Content string `gorm:"type:longtext;not null" json:"content,omitempty"`

	Cover  string `gorm:"size:255;not null;default:''" json:"cover"`
	Status int8   `gorm:"not null;default:0" json:"status"` // 1已发布 0草稿
	IsTop  int8   `gorm:"not null;default:0" json:"is_top"` // 1置顶

	ViewCount int `gorm:"not null;default:0" json:"view_count"`
	LikeCount int `gorm:"not null;default:0" json:"like_count"`

	// 外键 + 关联对象。gorm:"foreignKey:AuthorID" 告诉 GORM：
	// Preload("Author") 时用 articles.author_id 去查 users.id
	AuthorID uint64 `gorm:"index:idx_author;not null" json:"author_id"`
	Author   *User  `gorm:"foreignKey:AuthorID" json:"author,omitempty"`

	CategoryID uint64    `gorm:"index:idx_category;not null;default:0" json:"category_id"`
	Category   *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`

	// 多对多：GORM 通过中间表 article_tags 关联，两张表的 id 列名自动推导
	Tags []Tag `gorm:"many2many:article_tags" json:"tags,omitempty"`

	// 指针类型 = 数据库里允许 NULL（草稿没有发布时间）
	PublishedAt *time.Time `json:"published_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// gorm.DeletedAt 是软删除的约定类型：
	//   Delete()  -> 实际执行 UPDATE deleted_at = now()
	//   查询      -> 自动追加 WHERE deleted_at IS NULL
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Article) TableName() string { return "articles" }
