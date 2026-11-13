package repository

import (
	"blog/internal/model"

	"gorm.io/gorm"
)

// TagWithCount 标签 + 使用它的已发布文章数
type TagWithCount struct {
	model.Tag
	ArticleCount int64 `json:"article_count"`
}

// TagRepository 标签数据访问
type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) WithTx(tx *gorm.DB) *TagRepository {
	return &TagRepository{db: tx}
}

// ListWithArticleCount 标签列表带文章数
func (r *TagRepository) ListWithArticleCount() ([]TagWithCount, error) {
	var list []TagWithCount
	err := r.db.Model(&model.Tag{}).
		Select(`tags.*,
		        (SELECT COUNT(*) FROM article_tags at
		           JOIN articles a ON a.id = at.article_id
		          WHERE at.tag_id = tags.id
		            AND a.status = 1
		            AND a.deleted_at IS NULL) AS article_count`).
		Order("tags.id ASC").
		Scan(&list).Error
	return list, err
}

func (r *TagRepository) FindByID(id uint64) (*model.Tag, error) {
	var t model.Tag
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TagRepository) Create(t *model.Tag) error {
	return r.db.Omit("Articles").Create(t).Error
}

func (r *TagRepository) Save(t *model.Tag) error {
	return r.db.Omit("Articles").Save(t).Error
}

// DeleteWithLinks 删除标签并清理关联表。
// 必须放在事务里：只删标签会留下指向不存在标签的 article_tags 脏数据。
func (r *TagRepository) DeleteWithLinks(tagID uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tag_id = ?", tagID).Delete(&model.ArticleTag{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Tag{}, tagID).Error
	})
}

// ExistsNameOrSlug 名称或标识是否被别的标签占用
func (r *TagRepository) ExistsNameOrSlug(name, slug string, excludeID uint64) (bool, error) {
	q := r.db.Model(&model.Tag{}).Where("(name = ? OR slug = ?)", name, slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}
