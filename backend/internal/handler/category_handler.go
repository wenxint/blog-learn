package handler

import (
	"blog/internal/model"
	"blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CategoryHandler 分类接口。和 ArticleHandler 一样，db 由 router 注入。
type CategoryHandler struct {
	db *gorm.DB
}

func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

// CategoryWithCount 在分类的全部字段之外，多一个文章数。
// 第一行没有字段名、直接写 model.Category，叫嵌入：外层会带上 Category 的所有字段，
// JSON 里不会多包一层，看起来仍是 {id, name, slug, ..., article_count}。
type CategoryWithCount struct {
	model.Category
	ArticleCount int64 `json:"article_count"`
}

// List 分类列表：GET /api/v1/categories
func (h *CategoryHandler) List(c *gin.Context) {
	var list []CategoryWithCount

	// 相关子查询：每一行分类，都去 articles 里数一次「属于我、已发布、未软删除」的文章。
	// 用子查询而不是 JOIN + GROUP BY，是为了避开 MySQL ONLY_FULL_GROUP_BY 对分组字段的限制。
	// Scan 把查询结果按列名填进 list，不走 Model 的自动关联。
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
