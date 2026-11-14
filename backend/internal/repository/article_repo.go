package repository

import (
	"blog/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ArticleRepository 只负责数据存取，不含业务判断。
// 它不 import service，保证依赖单向（service -> repository -> gorm）。
type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// WithTx 返回一个绑定到事务上的新 Repository。
// 用法：s.repo.WithTx(tx).Create(...)，不要用 tx 去污染原对象。
func (r *ArticleRepository) WithTx(tx *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: tx}
}

// ArticleFilter 列表查询条件。用指针表达「可选」：
// Status 为 nil 表示不限状态（后台看全部），指向 1 表示只看已发布。
type ArticleFilter struct {
	Status     *int8
	Keyword    string
	CategoryID uint64
	TagID      uint64
	// PublishedMonth 形如 "2026-11"，用于归档页按月筛选
	PublishedMonth string
	Page           int
	PageSize       int
	// OrderAsc 为 true 时按发布时间正序（系列文章 01→13），默认倒序（最新在前）
	OrderAsc bool
}

// build 每次返回一个干净的链式对象。
// *gorm.DB 是有状态的，复用同一个对象会让条件互相叠加。
func (r *ArticleRepository) build(f ArticleFilter) *gorm.DB {
	d := r.db.Model(&model.Article{})

	if f.Status != nil {
		d = d.Where("status = ?", *f.Status)
	}
	if f.Keyword != "" {
		d = d.Where("title LIKE ?", "%"+f.Keyword+"%")
	}
	if f.CategoryID > 0 {
		d = d.Where("category_id = ?", f.CategoryID)
	}
	if f.PublishedMonth != "" {
		// 用 DATE_FORMAT 匹配年月。注意这会让 idx_status_published 的
		// published_at 部分失效（函数包住列无法走索引），
		// 数据量大时应改成 published_at >= 月初 AND published_at < 下月初
		d = d.Where("DATE_FORMAT(published_at, '%Y-%m') = ?", f.PublishedMonth)
	}
	if f.TagID > 0 {
		// 子查询：先找出打了这个标签的文章 id
		d = d.Where("id IN (?)", r.db.Model(&model.Tag{}).
			Select("article_id").
			Table("article_tags").
			Where("tag_id = ?", f.TagID))
	}
	return d
}

// List 分页查询文章列表（不含正文）
func (r *ArticleRepository) List(f ArticleFilter) ([]model.Article, int64, error) {
	var total int64
	if err := r.build(f).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "is_top DESC, published_at DESC"
	if f.OrderAsc {
		order = "published_at ASC"
	}

	var list []model.Article
	err := r.build(f).
		Select("id", "title", "summary", "cover", "status", "is_top",
			"view_count", "like_count", "author_id", "category_id",
			"published_at", "created_at").
		Preload("Category").
		Preload("Author").
		Preload("Tags").
		Order(order).
		Limit(f.PageSize).
		Offset((f.Page - 1) * f.PageSize).
		Find(&list).Error
	return list, total, err
}

// FindByID 查单篇（含正文与关联）。找不到返回 gorm.ErrRecordNotFound，由 service 翻译。
func (r *ArticleRepository) FindByID(id uint64) (*model.Article, error) {
	var a model.Article
	err := r.db.
		Preload("Category").
		Preload("Author").
		Preload("Tags").
		First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// IncrementViewCount 阅读数自增。
// 必须用 SQL 表达式 view_count = view_count + 1，不能先读再写，否则并发下会丢计数。
func (r *ArticleRepository) IncrementViewCount(id uint64) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// Create 新建文章（写成功后 a.ID 会被填上）
//
// Omit(clause.Associations) 非常关键：不写的话 GORM 会把 a 上挂的
// Author / Category / Tags 一起 upsert，可能覆盖别的表的数据。
func (r *ArticleRepository) Create(a *model.Article) error {
	return r.db.Omit(clause.Associations).Create(a).Error
}

// Save 保存整篇（更新场景：先查出来改字段，再整体保存）
// 同样必须 Omit 关联，否则预加载进来的 Author / Category 会被一起写回
func (r *ArticleRepository) Save(a *model.Article) error {
	return r.db.Omit(clause.Associations).Save(a).Error
}

// Delete 软删除：因为模型里有 gorm.DeletedAt，Delete 实际是 UPDATE deleted_at
func (r *ArticleRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Article{}, id).Error
}

// ReplaceTags 全量替换文章的标签关联：先删旧的，再插新的。
//
// 为什么不用 Association("Tags").Replace(...)？
// 实测它会顺带 upsert tags 表（INSERT INTO tags ... ON DUPLICATE KEY），
// 多一次没必要的写，还可能被关联对象上的脏字段带回数据库。
// 关联表这种「只有两个外键」的场景，手写两行 SQL 更可控。
func (r *ArticleRepository) ReplaceTags(articleID uint64, tagIDs []uint64) error {
	if err := r.db.Where("article_id = ?", articleID).Delete(&model.ArticleTag{}).Error; err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}

	rows := make([]model.ArticleTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		rows = append(rows, model.ArticleTag{ArticleID: articleID, TagID: tagID})
	}
	return r.db.Create(&rows).Error
}

// CountByTitle 统计标题完全相同的文章数（用于示例/测试）
func (r *ArticleRepository) CountByTitle(title string) (int64, error) {
	var n int64
	err := r.db.Model(&model.Article{}).Where("title = ?", title).Count(&n).Error
	return n, err
}

// ArchiveItem 归档条目：某个月有多少篇已发布文章
type ArchiveItem struct {
	Month string `json:"month"` // 形如 2026-11
	Count int64  `json:"count"`
}

// Archives 按月份统计已发布文章数量，倒序返回。
//
// DATE_FORMAT 把 published_at 截成 YYYY-MM 再分组。
// 真实项目里这种聚合通常走缓存（数据不常变、查询却频繁）。
func (r *ArticleRepository) Archives() ([]ArchiveItem, error) {
	var list []ArchiveItem
	err := r.db.Model(&model.Article{}).
		Select("DATE_FORMAT(published_at, '%Y-%m') AS month, COUNT(*) AS count").
		Where("status = ? AND published_at IS NOT NULL", 1).
		Group("month").
		Order("month DESC").
		Scan(&list).Error
	return list, err
}
