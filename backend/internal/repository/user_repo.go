package repository

import (
	"blog/internal/model"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) WithTx(tx *gorm.DB) *UserRepository {
	return &UserRepository{db: tx}
}

// FindByUsername 按用户名查用户。找不到返回 gorm.ErrRecordNotFound。
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID 按主键查用户
func (r *UserRepository) FindByID(id uint64) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// ExistsByUsername 判断用户名是否已被占用。
// 用 Count 而不是 Find：只需要知道「有没有」，不关心具体内容。
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var n int64
	if err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

// Create 新建用户
func (r *UserRepository) Create(u *model.User) error {
	return r.db.Create(u).Error
}

// UpdatePasswordHash 只更新密码哈希一列，避免把整个用户对象写回去
func (r *UserRepository) UpdatePasswordHash(id uint64, passwordHash string) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash).Error
}
