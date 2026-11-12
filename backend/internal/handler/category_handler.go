package handler

import (
	"blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// CategoryItem 分类列表项（第 5 课会替换成 model.Category）
type CategoryItem struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Description  string `json:"description"`
	ArticleCount int    `json:"article_count"`
}

// mockCategories 假数据，article_count 与数据库种子数据一致
var mockCategories = []CategoryItem{
	{1, "技术笔记", "tech", "后端、前端、数据库相关的学习记录", 3},
	{2, "生活随笔", "life", "写点日常和想法", 1},
}

// ListCategories 分类列表：GET /api/v1/categories
func ListCategories(c *gin.Context) {
	response.OK(c, gin.H{"list": mockCategories})
}
