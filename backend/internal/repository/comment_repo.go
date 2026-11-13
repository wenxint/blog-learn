package repository

import (
	"blog/internal/model"

	"gorm.io/gorm"
)

// CommentRepository 评论数据访问
type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) WithTx(tx *gorm.DB) *CommentRepository {
	return &CommentRepository{db: tx}
}

// ListByArticle 查某篇文章下指定状态的评论，按时间正序（楼层顺序）
func (r *CommentRepository) ListByArticle(articleID uint64, status int8) ([]model.Comment, error) {
	var list []model.Comment
	err := r.db.
		Where("article_id = ? AND status = ?", articleID, status).
		Order("created_at ASC, id ASC").
		Find(&list).Error
	return list, err
}

// ListByStatus 后台按状态分页查评论（不带文章信息，够用）
func (r *CommentRepository) ListByStatus(status int8, page, pageSize int) ([]model.Comment, int64, error) {
	var total int64
	if err := r.db.Model(&model.Comment{}).Where("status = ?", status).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.Comment
	err := r.db.
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *CommentRepository) FindByID(id uint64) (*model.Comment, error) {
	var c model.Comment
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CommentRepository) Create(c *model.Comment) error {
	return r.db.Create(c).Error
}

func (r *CommentRepository) UpdateStatus(id uint64, status int8) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteWithReplies 删除评论及其所有回复。
// 只删父评论会留下 parent_id 指向不存在记录的孤儿回复，前台组装树的时就会丢数据。
func (r *CommentRepository) DeleteWithReplies(id uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("parent_id = ?", id).Delete(&model.Comment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Comment{}, id).Error
	})
}

// CountByArticle 统计某篇文章的评论数（用于列表页展示）
func (r *CommentRepository) CountByArticle(articleID uint64, status int8) (int64, error) {
	var n int64
	err := r.db.Model(&model.Comment{}).
		Where("article_id = ? AND status = ?", articleID, status).
		Count(&n).Error
	return n, err
}

// CountPending 待审核评论总数（后台角标用）
func (r *CommentRepository) CountPending() (int64, error) {
	var n int64
	err := r.db.Model(&model.Comment{}).Where("status = ?", 0).Count(&n).Error
	return n, err
}
