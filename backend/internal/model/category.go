package model

import "time"

// Category 对应表 categories
type Category struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:50;not null;uniqueIndex:uk_name" json:"name"`
	Slug        string    `gorm:"size:50;not null;uniqueIndex:uk_slug" json:"slug"`
	Description string    `gorm:"size:200;not null;default:''" json:"description"`
	Sort        int       `gorm:"not null;default:0" json:"sort"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

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
