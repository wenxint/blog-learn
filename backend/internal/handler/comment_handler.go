package handler

import (
	"blog/internal/pkg/pagination"
	"blog/internal/pkg/response"
	"blog/internal/service"

	"github.com/gin-gonic/gin"
)

// CommentHandler 评论接口
type CommentHandler struct {
	svc *service.CommentService
}

func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

type createCommentRequest struct {
	Nickname string `json:"nickname" binding:"required"`
	Email    string `json:"email" binding:"omitempty,max=100"`
	Content  string `json:"content" binding:"required"`
	ParentID uint64 `json:"parent_id"`
}

type commentStatusRequest struct {
	Status *int8 `json:"status" binding:"required,oneof=0 1"`
}

type commentListQuery struct {
	pagination.Query
	Status *int8 `form:"status" binding:"omitempty,oneof=0 1"`
}

// List 文章评论列表（树形）：GET /api/v1/articles/:id/comments
func (h *CommentHandler) List(c *gin.Context) {
	articleID, ok := parseID(c)
	if !ok {
		return
	}

	list, err := h.svc.ListTree(articleID)
	if err != nil {
		fail(c, err, "查询评论失败")
		return
	}
	response.OK(c, gin.H{"list": list})
}

// Create 发表评论：POST /api/v1/articles/:id/comments
// 默认进入待审核状态，管理员通过后才展示。
func (h *CommentHandler) Create(c *gin.Context) {
	articleID, ok := parseID(c)
	if !ok {
		return
	}

	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	comment, err := h.svc.Create(articleID, service.CreateCommentInput{
		Nickname: req.Nickname,
		Email:    req.Email,
		Content:  req.Content,
		ParentID: req.ParentID,
		// 从连接取真实 IP，不信任请求体里前端自报的 IP
		IP: c.ClientIP(),
	})
	if err != nil {
		fail(c, err, "发表评论失败")
		return
	}

	response.OK(c, gin.H{
		"comment": comment,
		"message": "评论已提交，等待管理员审核后展示",
	})
}

// AdminList 后台评论列表：GET /api/v1/admin/comments?status=0&page=1
func (h *CommentHandler) AdminList(c *gin.Context) {
	var q commentListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.InvalidParam(c, "参数错误: "+err.Error())
		return
	}
	q.Normalize()

	status := service.CommentPending
	if q.Status != nil {
		status = *q.Status
	}

	list, total, err := h.svc.ListByStatus(status, q.Page, q.PageSize)
	if err != nil {
		fail(c, err, "查询评论列表失败")
		return
	}

	pending, _ := h.svc.CountPending()
	response.OK(c, gin.H{
		"list":          list,
		"total":         total,
		"page":          q.Page,
		"page_size":     q.PageSize,
		"pending_count": pending,
	})
}

// SetStatus 审核评论：PUT /api/v1/admin/comments/:id/status
func (h *CommentHandler) SetStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req commentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	if err := h.svc.SetStatus(id, *req.Status); err != nil {
		fail(c, err, "审核评论失败")
		return
	}
	response.OK(c, gin.H{"id": id, "status": *req.Status})
}

// Delete 删除评论：DELETE /api/v1/admin/comments/:id
func (h *CommentHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.svc.Delete(id); err != nil {
		fail(c, err, "删除评论失败")
		return
	}
	response.OK(c, gin.H{"id": id, "deleted": true})
}
