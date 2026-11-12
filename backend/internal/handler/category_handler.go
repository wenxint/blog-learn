package handler

import (
	"blog/internal/model"
	"blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CategoryHandler 分类接口
type CategoryHandler struct {
	db *gorm.DB
}

func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

// CategoryWithCount 分类 + 该分类下已发布文章数
type CategoryWithCount struct {
	model.Category
	ArticleCount int64 `json:"article_count"`
}

// List 分类列表：GET /api/v1/categories
func (h *CategoryHandler) List(c *gin.Context) {
	var list []CategoryWithCount

	// 用相关子查询统计文章数，避免 GROUP BY 与 ONLY_FULL_GROUP_BY 的坑
	err := h.db.Model(&model.Category{}).
		Select(`categories.*,
		        (SELECT COUNT(*) FROM articles a
		         WHERE a.category_id = categories.id
		           AND a.status = 1
		           AND a.deleted_at IS NULL) AS article_count`).
		Order("categories.sort ASC, categories.id ASC").
		Scan(&list).Error
	if err != nil {
		response.ServerError(c, "查询分类列表失败")
		return
	}

	response.OK(c, gin.H{"list": list})
}
