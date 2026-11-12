package handler

import (
	"sort"
	"time"

	"blog/internal/model"
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

// ModelCheck 单个模型的检查结果
type ModelCheck struct {
	Model        string   `json:"model"`
	Table        string   `json:"table"`
	TableExists  bool     `json:"table_exists"`
	ModelColumns []string `json:"model_columns"`
	DBColumns    []string `json:"db_columns"`
	MissingInDB  []string `json:"missing_in_db"` // 模型有、表里没有
	ExtraInDB    []string `json:"extra_in_db"`   // 表里有、模型没映射
}

// Models 校验 GORM 模型与真实表结构是否一致：GET /api/v1/debug/models
//
// 这是第 5 课的核心工具：写 Model 之后先让它告诉你「模型和表对不对得上」，
// 而不是等到查询报 Unknown column 才发现。
func (h *DebugHandler) Models(c *gin.Context) {
	samples := []any{
		&model.User{},
		&model.Category{},
		&model.Tag{},
		&model.Article{},
		&model.Comment{},
	}

	out := make([]ModelCheck, 0, len(samples))
	for _, s := range samples {
		stmt := &gorm.Statement{DB: h.db}
		if err := stmt.Parse(s); err != nil {
			response.ServerError(c, "解析模型失败: "+err.Error())
			return
		}

		check := ModelCheck{
			Model:        stmt.Schema.Name,
			Table:        stmt.Schema.Table,
			ModelColumns: stmt.Schema.DBNames,
		}

		check.TableExists = h.db.Migrator().HasTable(check.Table)
		if check.TableExists {
			cols, err := h.db.Migrator().ColumnTypes(check.Table)
			if err != nil {
				response.ServerError(c, "读取表结构失败: "+err.Error())
				return
			}
			for _, col := range cols {
				check.DBColumns = append(check.DBColumns, col.Name())
			}
		}
		sort.Strings(check.DBColumns)

		check.MissingInDB = diffSet(check.ModelColumns, check.DBColumns)
		check.ExtraInDB = diffSet(check.DBColumns, check.ModelColumns)
		out = append(out, check)
	}

	response.OK(c, gin.H{"models": out})
}

// diffSet 返回 a 里有、b 里没有的元素
func diffSet(a, b []string) []string {
	set := make(map[string]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}
	res := make([]string, 0)
	for _, v := range a {
		if _, ok := set[v]; !ok {
			res = append(res, v)
		}
	}
	sort.Strings(res)
	return res
}
