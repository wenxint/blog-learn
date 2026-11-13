package handler

import (
	"blog/internal/pkg/response"
	"blog/internal/service"

	"github.com/gin-gonic/gin"
)

// UploadHandler 图片上传
type UploadHandler struct {
	svc *service.UploadService
}

func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

// Upload 上传图片：POST /api/v1/admin/upload
//
// 请求必须是 multipart/form-data，字段名固定为 file：
//
//	curl -F "file=@/path/to/pic.png" -H "Authorization: Bearer <token>" .../admin/upload
func (h *UploadHandler) Upload(c *gin.Context) {
	// FormFile 同时受 router 里设置的 MaxMultipartMemory 和 service 里的 maxMB 约束
	file, err := c.FormFile("file")
	if err != nil {
		response.InvalidParam(c, "缺少 file 字段，或请求不是 multipart/form-data: "+err.Error())
		return
	}

	result, err := h.svc.Save(file)
	if err != nil {
		fail(c, err, "上传图片失败")
		return
	}

	response.OK(c, result)
}
