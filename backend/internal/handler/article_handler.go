package handler

import (
	"blog/internal/model"
	"blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ArticleHandler 文章接口
type ArticleHandler struct {
	db *gorm.DB
}

func NewArticleHandler(db *gorm.DB) *ArticleHandler {
	return &ArticleHandler{db: db}
}

// ListArticleQuery 列表查询参数
type ListArticleQuery struct {
	Page       int    `form:"page" binding:"omitempty,gte=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
	Keyword    string `form:"keyword" binding:"omitempty,max=50"`
	CategoryID uint64 `form:"category_id" binding:"omitempty,gte=1"`
}

// PageResult 分页结果（第 11 课会抽到 pkg/pagination）
type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

const defaultPageSize = 20

// List 文章列表：GET /api/v1/articles?page=1&page_size=20&keyword=Go&category_id=1
func (h *ArticleHandler) List(c *gin.Context) {
	var q ListArticleQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.InvalidParam(c, "参数错误: "+err.Error())
		return
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = defaultPageSize
	}

	// 把查询条件封装成函数，每次调用返回一个干净的链式对象。
	// 重要：*gorm.DB 的链式调用是有状态的，同一个对象复用会把条件串在一起
	//（比如 Count 之后再 Find，条件会叠加两次），所以必须每次重建。
	base := func() *gorm.DB {
		d := h.db.Model(&model.Article{}).Where("status = ?", 1) // 只查已发布
		if q.Keyword != "" {
			d = d.Where("title LIKE ?", "%"+q.Keyword+"%")
		}
		if q.CategoryID > 0 {
			d = d.Where("category_id = ?", q.CategoryID)
		}
		return d
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		response.ServerError(c, "查询文章总数失败")
		return
	}

	var list []model.Article
	err := base().
		// 列表不返回正文（longtext 很大），详情接口才查
		Select("id", "title", "summary", "cover", "status", "is_top",
			"view_count", "like_count", "author_id", "category_id",
					"published_at", "created_at").
		Preload("Category"). // 预加载关联：另外发两条 SQL，避免 N+1
		Preload("Author").
		Preload("Tags").
		Order("is_top DESC, published_at DESC").
		Limit(q.PageSize).
		Offset((q.Page - 1) * q.PageSize).
		Find(&list).Error
	if err != nil {
		response.ServerError(c, "查询文章列表失败")
		return
	}

	response.OK(c, PageResult{
		List:     list,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}
