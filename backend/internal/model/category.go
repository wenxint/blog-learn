// Package model 是数据库表在 Go 里的样子。一个结构体对应一张表，一个字段对应一列。
// GORM 靠字段上的反引号标签决定列类型、索引、JSON 名字，不靠字段在结构体里的先后顺序。
package model

import "time"

// Category 对应表 categories。
// 标签怎么读，以 Name 这行为例：
//
//	gorm:"size:50;not null;uniqueIndex:uk_name"  → 列 VARCHAR(50)、非空、唯一索引名叫 uk_name
//	json:"name"                                  → 接口 JSON 里的键是 name
type Category struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:50;not null;uniqueIndex:uk_name" json:"name"`
	Slug        string    `gorm:"size:50;not null;uniqueIndex:uk_slug" json:"slug"`
	Description string    `gorm:"size:200;not null;default:''" json:"description"`
	Sort        int       `gorm:"not null;default:0" json:"sort"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 显式指定表名。接收者写成 (Category) 而不是 (c Category)，
// 是因为函数体里用不到那一份数据，只要类型信息。
func (Category) TableName() string { return "categories" }

// Tag 对应表 tags
type Tag struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:50;not null;uniqueIndex:uk_name" json:"name"`
	Slug      string    `gorm:"size:50;not null;uniqueIndex:uk_slug" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Tag) TableName() string { return "tags" }
