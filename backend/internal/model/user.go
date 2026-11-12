package model

import "time"

// User 对应表 users
//
// GORM 默认命名规则：结构体 User -> 表名 users（复数、小写下划线）
// 字段 Username -> 列名 username
type User struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string `gorm:"size:50;not null;uniqueIndex:uk_username" json:"username"`

	// json:"-" 是关键：密码哈希永远不会出现在任何接口响应里
	PasswordHash string `gorm:"size:100;not null" json:"-"`

	Nickname  string    `gorm:"size:50;not null;default:''" json:"nickname"`
	Email     string    `gorm:"size:100;not null;default:''" json:"email"`
	Avatar    string    `gorm:"size:255;not null;default:''" json:"avatar"`
	Role      string    `gorm:"size:20;not null;default:'author'" json:"role"`
	Status    int8      `gorm:"not null;default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 显式指定表名。这里其实和默认推导结果一致，
// 写出来是为了让「表名从哪来」一目了然；以后表名和结构体不一致时必须用它。
func (User) TableName() string { return "users" }
