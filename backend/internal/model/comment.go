package model

import "time"

// Comment 对应表 comments
// 楼中楼用 parent_id 自关联：parent_id = 0 表示一级评论
type Comment struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleID uint64 `gorm:"index:idx_article;not null" json:"article_id"`
	ParentID  uint64 `gorm:"not null;default:0" json:"parent_id"`

	Nickname string `gorm:"size:50;not null" json:"nickname"`

	// 邮箱和 IP 不对外展示
	Email string `gorm:"size:100;not null;default:''" json:"-"`

	Content string `gorm:"size:1000;not null" json:"content"`
	Status  int8   `gorm:"index:idx_status;not null;default:0" json:"status"` // 1已通过 0待审核
	IP      string `gorm:"size:45;not null;default:''" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 显式指定表名，避免依赖 GORM 的复数推导。
func (Comment) TableName() string { return "comments" }
