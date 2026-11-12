package handler

import (
	"strings"

	"blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListArticleQuery 列表查询参数，binding tag 由 Gin 自动校验
type ListArticleQuery struct {
	Page     int    `form:"page" binding:"omitempty,gte=1"`
	PageSize int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
	Keyword  string `form:"keyword" binding:"omitempty,max=50"`
}

// ArticleItem 列表项（第 5 课会替换成 model.Article）
type ArticleItem struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	ViewCount   int    `json:"view_count"`
	PublishedAt string `json:"published_at"`
}

// PageResult 分页结果（第 11 课会抽到 pkg/pagination）
type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// mockArticles 假数据，字段与数据库种子数据一致
var mockArticles = []ArticleItem{
	{1, "Go 从零开始：第一个 HTTP 服务", "用标准库 net/http 起一个服务，理解 Handler 到底是什么。", 320, "2026-11-01 10:00:00"},
	{2, "MySQL 索引到底怎么建", "联合索引为什么要讲究顺序？最左前缀原则讲清楚。", 210, "2026-11-05 14:30:00"},
	{3, "Vue3 组合式 API 上手笔记", "script setup 和以前 Options API 的差别。", 156, "2026-11-08 09:15:00"},
	{4, "GORM 会用但不懂的那些事", "AutoMigrate 到底做了什么，为什么生产不建议直接用。", 89, "2026-11-10 20:00:00"},
}

// List 文章列表：GET /api/v1/articles?page=1&page_size=10&keyword=Go
func List(c *gin.Context) {
	var q ListArticleQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.InvalidParam(c, "参数错误: "+err.Error())
		return
	}

	// 给默认值
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}

	// 第 3 课先返回假数据，第 7 课换成真实数据库查询
	list := mockArticles
	if q.Keyword != "" {
		filtered := make([]ArticleItem, 0, len(mockArticles))
		for _, a := range mockArticles {
			// 简易版模糊匹配，真实实现交给 SQL 的 LIKE
			if strings.Contains(a.Title, q.Keyword) {
				filtered = append(filtered, a)
			}
		}
		list = filtered
	}

	// 模拟内存分页
	start := (q.Page - 1) * q.PageSize
	if start > len(list) {
		start = len(list)
	}
	end := start + q.PageSize
	if end > len(list) {
		end = len(list)
	}

	response.OK(c, PageResult{
		List:     list[start:end],
		Total:    int64(len(list)),
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}
