package handler

import (
	"blog/internal/pkg/response"
	"blog/internal/service"

	"github.com/gin-gonic/gin"
)

// TagHandler 标签接口
type TagHandler struct {
	svc *service.TagService
}

func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

type tagRequest struct {
	Name string `json:"name" binding:"required,max=50"`
	Slug string `json:"slug" binding:"required,max=50"`
}

// List 标签列表：GET /api/v1/tags
func (h *TagHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		fail(c, err, "查询标签列表失败")
		return
	}
	response.OK(c, gin.H{"list": list})
}

// Create 新建标签：POST /api/v1/admin/tags
func (h *TagHandler) Create(c *gin.Context) {
	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	t, err := h.svc.Create(service.TagInput{Name: req.Name, Slug: req.Slug})
	if err != nil {
		fail(c, err, "创建标签失败")
		return
	}
	response.OK(c, t)
}

// Update 编辑标签：PUT /api/v1/admin/tags/:id
func (h *TagHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	t, err := h.svc.Update(id, service.TagInput{Name: req.Name, Slug: req.Slug})
	if err != nil {
		fail(c, err, "更新标签失败")
		return
	}
	response.OK(c, t)
}

// Delete 删除标签：DELETE /api/v1/admin/tags/:id
func (h *TagHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.svc.Delete(id); err != nil {
		fail(c, err, "删除标签失败")
		return
	}
	response.OK(c, gin.H{"id": id, "deleted": true})
}
