package repository

import (
	"blog/internal/model"

	"gorm.io/gorm"
)

// CategoryWithCount 分类 + 该分类下已发布文章数
type CategoryWithCount struct {
	model.Category
	ArticleCount int64 `json:"article_count"`
}

// CategoryRepository 分类数据访问
type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) WithTx(tx *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: tx}
}

// ListWithArticleCount 列表带文章数。
// 用相关子查询而不是 JOIN + GROUP BY：不受 ONLY_FULL_GROUP_BY 影响，也更好读。
func (r *CategoryRepository) ListWithArticleCount() ([]CategoryWithCount, error) {
	var list []CategoryWithCount
	err := r.db.Model(&model.Category{}).
		Select(`categories.*,
		        (SELECT COUNT(*) FROM articles a
		         WHERE a.category_id = categories.id
		           AND a.status = 1
		           AND a.deleted_at IS NULL) AS article_count`).
		Order("categories.sort ASC, categories.id ASC").
		Scan(&list).Error
	return list, err
}

func (r *CategoryRepository) FindByID(id uint64) (*model.Category, error) {
	var c model.Category
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) Create(c *model.Category) error {
	return r.db.Omit("Articles").Create(c).Error
}

func (r *CategoryRepository) Save(c *model.Category) error {
	return r.db.Omit("Articles").Save(c).Error
}

func (r *CategoryRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Category{}, id).Error
}

// CountArticles 统计该分类下（未删除的）文章数量，用于删除前的业务判断
func (r *CategoryRepository) CountArticles(categoryID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.Article{}).Where("category_id = ?", categoryID).Count(&n).Error
	return n, err
}

// ExistsNameOrSlug 判断名称或标识是否被别的分类占用（excludeID>0 时排除自己）
func (r *CategoryRepository) ExistsNameOrSlug(name, slug string, excludeID uint64) (bool, error) {
	q := r.db.Model(&model.Category{}).Where("(name = ? OR slug = ?)", name, slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}
