// Package handler 是 HTTP 入口：读参数、做校验、调用 service、用 response 写出 JSON。
// 从第 6 课起这里不再直接查数据库：
//
//	「查数据库」下沉到 repository（internal/repository）
//	「业务规则 / 事务」下沉到 service（internal/service）
//	handler 只保留三件事：解析请求参数、调用 service、把结果或错误翻译成 HTTP 响应。
//
// 好处：想在别处（定时任务、命令行工具）复用同一套业务逻辑时，不用连着 Gin 一起搬。
package handler

import (
	"strconv"

	"blog/internal/pkg/authctx"
	"blog/internal/pkg/response"
	"blog/internal/repository"
	"blog/internal/service"

	"github.com/gin-gonic/gin"
)

// ArticleHandler 文章接口。
// 小写的 svc 包外看不见，只能通过 NewArticleHandler 把业务层对象交进来。
type ArticleHandler struct {
	svc *service.ArticleService
}

// NewArticleHandler 构造函数。Go 没有 class，约定用 NewXxx 返回指针。
// router 启动时调用一次，之后每个请求都复用同一份 handler（service 内部的 db 是连接池，可以并发用）。
func NewArticleHandler(svc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

// ListArticleQuery 是网址问号后面的参数，例如 ?page=1&page_size=20&keyword=Go。
// form:"page" 告诉 Gin：把 query 里的 page 填进 Page。
// binding 是校验规则：omitempty 表示可以不传；gte=1 表示传了就必须 ≥ 1；lte=100 是上限。
type ListArticleQuery struct {
	Page       int    `form:"page" binding:"omitempty,gte=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
	Keyword    string `form:"keyword" binding:"omitempty,max=50"`
	CategoryID uint64 `form:"category_id" binding:"omitempty,gte=1"`
	TagID      uint64 `form:"tag_id" binding:"omitempty,gte=1"`
}

// articleRequest 是「新建/编辑文章」的请求体，给前端传的 JSON 用的。
// 为什么不直接用 model.Article 接收？
//  1. 防越权：前端不该能指定 id、view_count、created_at 这些字段
//  2. 防耦合：数据库加个内部字段，不会自动暴露成接口参数
//
// json 标签写的是前端要传的字段名；binding 是校验规则，required 表示必填。
// Status / IsTop 用指针 *int8：nil 表示「前端没传」，这样才能区分「传了 0」和「没传」。
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

// PageResult 分页结果（第 11 课会抽到 pkg/pagination）
type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

const defaultPageSize = 20

// List 文章列表：GET /api/v1/articles?page=1&page_size=20&keyword=Go&category_id=1
// (h *ArticleHandler) 表示这是 ArticleHandler 的方法。router 里写 articleHandler.List 就能注册它。
// h 是这份 handler，c 是这一次请求。
func (h *ArticleHandler) List(c *gin.Context) {
	// var 声明一个零值结构体。ShouldBindQuery 按 form 标签把网址参数填进去。
	// 传 &q 是因为函数要修改 q 本身，必须给它地址。校验失败直接 400，不再查库。
	var q ListArticleQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.InvalidParam(c, "参数错误: "+err.Error())
		return
	}
	// 不传就给默认值。第 3 课的作业要求默认每页 20 条。
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = defaultPageSize
	}

	// 前台只展示已发布：把 status=1 装进指针传给 service。
	// 用指针是为了表达「限定已发布」和「不限状态」的区别——后台列表就不限状态。
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
		fail(c, err, "查询文章列表失败")
		return
	}

	response.OK(c, PageResult{List: list, Total: total, Page: q.Page, PageSize: q.PageSize})
}

// Detail 文章详情：GET /api/v1/articles/1
// c.Param("id") 取的是路由里 :id 那一段，取到的永远是字符串，要自己转数字。
func (h *ArticleHandler) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	// 第二个参数 true：前台访问，草稿返回 404（不告诉调用方「存在但没发布」）
	a, err := h.svc.Detail(id, true)
	if err != nil {
		fail(c, err, "查询文章详情失败")
		return
	}
	response.OK(c, a)
}

// Create 新建文章：POST /api/v1/admin/articles
// ShouldBindJSON 会把请求体里的 JSON 解析进结构体，并按 binding 规则校验。
func (h *ArticleHandler) Create(c *gin.Context) {
	var req articleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	a, err := h.svc.Create(toInput(req, authctx.UserID(c)))
	if err != nil {
		fail(c, err, "创建文章失败")
		return
	}
	response.OK(c, a)
}

// Update 编辑文章：PUT /api/v1/admin/articles/:id
// 这里用的是「全量覆盖」语义：没传的字段会被清空成零值。
// 如果要「只改传了的字段」，就得用 PATCH 语义 + 全字段指针。
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

	a, err := h.svc.Update(id, toInput(req, authctx.UserID(c)))
	if err != nil {
		fail(c, err, "更新文章失败")
		return
	}
	response.OK(c, a)
}

// Delete 删除文章：DELETE /api/v1/admin/articles/:id
// 注意是「软删除」：数据还在表里，只是 deleted_at 被填上，查询时自动被过滤掉。
func (h *ArticleHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.svc.Delete(id); err != nil {
		fail(c, err, "删除文章失败")
		return
	}
	response.OK(c, gin.H{"id": id, "deleted": true})
}

// ---------- 下面是本文件内部的小工具，不对外暴露 ----------

// toInput 把 HTTP 请求体（DTO）转成 service 的入参。
// 这层转换看着多余，其实是在隔离「接口契约」和「业务模型」：以后前端字段改名，只改这里。
func toInput(req articleRequest, currentUserID uint64) service.ArticleInput {
	status := service.StatusDraft
	if req.Status != nil {
		status = *req.Status
	}
	var isTop int8
	if req.IsTop != nil {
		isTop = *req.IsTop
	}
	// 作者优先取 token 里的当前用户，防止前端伪造 author_id 冒充别人发文
	authorID := currentUserID
	if authorID == 0 {
		authorID = req.AuthorID // 未登录时（理论上已被中间件拦住）兜底
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

// parseID 解析路径参数 :id。
// 返回 (id, true) 表示成功；失败时它已经自己写好 400 响应，调用方直接 return 即可。
func parseID(c *gin.Context) (uint64, bool) {
	raw := c.Param("id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		response.InvalidParam(c, "id 必须是正整数: "+raw)
		return 0, false
	}
	return id, true
}
