package handler

import (
	"blog/internal/pkg/authctx"
	"blog/internal/pkg/response"
	"blog/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证相关接口
type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
	Email    string `json:"email" binding:"omitempty,email,max=100"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// Register 注册：POST /api/v1/auth/register
// 注册成功后不自动登录，让前端走一次登录流程（顺便验证密码确实记住了）。
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	u, err := h.svc.Register(req.Username, req.Password, req.Nickname, req.Email)
	if err != nil {
		fail(c, err, "注册失败")
		return
	}

	// 注意：这里返回的 model.User 里 PasswordHash 有 json:"-"，不会被序列化出去
	response.OK(c, u)
}

// Login 登录：POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	result, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		fail(c, err, "登录失败")
		return
	}
	response.OK(c, result)
}

// Profile 当前用户资料：GET /api/v1/auth/profile （需要登录）
func (h *AuthHandler) Profile(c *gin.Context) {
	// 用户 id 来自 token，不从前端参数取——否则谁都能查别人资料
	userID := authctx.UserID(c)

	u, err := h.svc.Profile(userID)
	if err != nil {
		fail(c, err, "查询用户资料失败")
		return
	}
	response.OK(c, u)
}

// ChangePassword 修改密码：PUT /api/v1/auth/password （需要登录）
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "请求体错误: "+err.Error())
		return
	}

	if err := h.svc.ChangePassword(authctx.UserID(c), req.OldPassword, req.NewPassword); err != nil {
		fail(c, err, "修改密码失败")
		return
	}
	response.OK(c, gin.H{"changed": true})
}
