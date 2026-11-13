package handler

import (
	"errors"
	"log"
	"strconv"

	"blog/internal/pkg/response"
	"blog/internal/repository"
	"blog/internal/service"

	"github.com/gin-gonic/gin"
)

// ArticleHandler 只做 HTTP 相关的事：解析参数、调 service、组装响应、翻译错误。
type ArticleHandler struct {
	svc *service.ArticleService
}

func NewArticleHandler(svc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

// ListArticleQuery 列表查询参数
type ListArticleQuery struct {
	Page       int    `form:"page" binding:"omitempty,gte=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
	Keyword    string `form:"keyword" binding:"omitempty,max=50"`
	CategoryID uint64 `form:"category_id" binding:"omitempty,gte=1"`
	TagID      uint64 `form:"tag_id" binding:"omitempty,gte=1"`
}

// articleRequest 创建/编辑的请求体（DTO）。
// 特意不直接用 model.Article：避免前端能改 id、created_at、view_count 这些字段。
type articleRequest struct {
	Title      string   `json:"title" binding:"required,max=200"`
	Summary    string   `json:"summary" binding:"omitempty,max=500"`
	Content    string   `json:"content" binding:"required"`
	Cover      string   `json:"cover" binding:"omitempty,max=255"`
	Status     *int8    `json:"status" binding:"omitempty,oneof=0 1"`
	IsTop      *int8    `json:"is_top" binding:"omitempty,oneof=0 1"`
	AuthorID   uint64   `json:"author_id"`
	CategoryID uint64   `json:"category_id"`
	TagIDs     []uint64 `json:"tag_ids"`
}

// PageResult 分页结果
type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

const defaultPageSize = 20

// List 文章列表：GET /api/v1/articles
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

	// 前台只展示已发布
	published := service.StatusPublished
	list, total, err := h.svc.List(repository.ArticleFilter{
		Status:     &published,
		Keyword:    q.Keyword,
		CategoryID: q.CategoryID,
		TagID:      q.TagID,
		Page:       q.Page,
		PageSize:   q.PageSize,
	})
	if err != nil {
		h.fail(c, err, "查询文章列表失败")
		return
	}

	response.OK(c, PageResult{List: list, Total: total, Page: q.Page, PageSize: q.PageSize})
}

// Detail 文章详情：GET /api/v1/articles/:id
func (h *ArticleHandler) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	a, err := h.svc.Detail(id, true) // 前台：草稿不可见
	if err != nil {
		h.fail(c, err, "查询文章详情失败")
		return
	}
	response.OK(c, a)
}

// Create 新建文章：POST /api/v1/admin/articles
func (h *ArticleHandler) Create(c *gin.Context) {
	var req articleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	a, err := h.svc.Create(toInput(req))
	if err != nil {
		h.fail(c, err, "创建文章失败")
		return
	}
	response.OK(c, a)
}

// Update 编辑文章：PUT /api/v1/admin/articles/:id
func (h *ArticleHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req articleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	a, err := h.svc.Update(id, toInput(req))
	if err != nil {
		h.fail(c, err, "更新文章失败")
		return
	}
	response.OK(c, a)
}

// Delete 删除文章（软删除）：DELETE /api/v1/admin/articles/:id
func (h *ArticleHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.svc.Delete(id); err != nil {
		h.fail(c, err, "删除文章失败")
		return
	}
	response.OK(c, gin.H{"id": id, "deleted": true})
}

// ---------- 内部辅助 ----------

// toInput DTO -> service 入参
func toInput(req articleRequest) service.ArticleInput {
	status := service.StatusDraft
	if req.Status != nil {
		status = *req.Status
	}
	var isTop int8
	if req.IsTop != nil {
		isTop = *req.IsTop
	}
	authorID := req.AuthorID
	if authorID == 0 {
		authorID = 1 // 第 10 课接入 JWT 后会改成「当前登录用户」
	}

	return service.ArticleInput{
		Title:      req.Title,
		Summary:    req.Summary,
		Content:    req.Content,
		Cover:      req.Cover,
		Status:     status,
		IsTop:      isTop,
		AuthorID:   authorID,
		CategoryID: req.CategoryID,
		TagIDs:     req.TagIDs,
	}
}

// parseID 解析路径参数 :id，失败时直接写 400 响应
func parseID(c *gin.Context) (uint64, bool) {
	raw := c.Param("id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		response.InvalidParam(c, "id 必须是正整数: "+raw)
		return 0, false
	}
	return id, true
}

// fail 把 service 的错误翻译成 HTTP 响应。
// 业务错误（NotFound / InvalidInput）原样返回给前端；
// 其他错误（数据库挂了等）只回笼统信息，细节写日志，避免把内部结构暴露给外部。
func (h *ArticleHandler) fail(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrInvalidInput):
		response.InvalidParam(c, err.Error())
	default:
		log.Printf("[ERROR] %s: %v", fallback, err)
		response.ServerError(c, fallback)
	}
}
