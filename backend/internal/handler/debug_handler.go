package handler

import (
	"time"

	"blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DebugHandler 诊断用途的接口，只在 dev 环境注册。
type DebugHandler struct {
	db *gorm.DB
}

// NewDebugHandler 依赖注入：把 *gorm.DB 传进来，而不是用全局变量
func NewDebugHandler(db *gorm.DB) *DebugHandler {
	return &DebugHandler{db: db}
}

// DB 数据库健康检查：GET /api/v1/debug/db
func (h *DebugHandler) DB(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		response.ServerError(c, "获取底层连接池失败: "+err.Error())
		return
	}

	// 实测一次 Ping，拿到真实往返耗时
	start := time.Now()
	pingErr := sqlDB.Ping()
	pingCost := time.Since(start)

	if pingErr != nil {
		response.ServerError(c, "数据库不可用: "+pingErr.Error())
		return
	}

	var version string
	var currentDB string
	h.db.Raw("SELECT VERSION()").Scan(&version)
	h.db.Raw("SELECT DATABASE()").Scan(&currentDB)

	var tables []string
	h.db.Raw(`SELECT table_name FROM information_schema.tables
	          WHERE table_schema = DATABASE() ORDER BY table_name`).Scan(&tables)

	stats := sqlDB.Stats()

	response.OK(c, gin.H{
		"version":  version,
		"database": currentDB,
		"tables":   tables,
		"ping_ms":  float64(pingCost.Microseconds()) / 1000.0,
		"pool": gin.H{
			"max_open":         stats.MaxOpenConnections,
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"wait_count":       stats.WaitCount,
		},
	})
}
