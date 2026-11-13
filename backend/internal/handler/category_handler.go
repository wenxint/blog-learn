package handler

import (
	"blog/internal/pkg/response"
	"blog/internal/service"

	"github.com/gin-gonic/gin"
)

// CategoryHandler 分类接口
type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

type categoryRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Slug        string `json:"slug" binding:"required,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
	Sort        int    `json:"sort"`
}

// List 分类列表：GET /api/v1/categories
func (h *CategoryHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		fail(c, err, "查询分类列表失败")
		return
	}
	response.OK(c, gin.H{"list": list})
}

// Create 新建分类：POST /api/v1/admin/categories （需 admin）
func (h *CategoryHandler) Create(c *gin.Context) {
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	cat, err := h.svc.Create(service.CategoryInput{
		Name: req.Name, Slug: req.Slug, Description: req.Description, Sort: req.Sort,
	})
	if err != nil {
		fail(c, err, "创建分类失败")
		return
	}
	response.OK(c, cat)
}

// Update 编辑分类：PUT /api/v1/admin/categories/:id
func (h *CategoryHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	cat, err := h.svc.Update(id, service.CategoryInput{
		Name: req.Name, Slug: req.Slug, Description: req.Description, Sort: req.Sort,
	})
	if err != nil {
		fail(c, err, "更新分类失败")
		return
	}
	response.OK(c, cat)
}

// Delete 删除分类：DELETE /api/v1/admin/categories/:id
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.svc.Delete(id); err != nil {
		fail(c, err, "删除分类失败")
		return
	}
	response.OK(c, gin.H{"id": id, "deleted": true})
}
