// Package pagination 统一分页参数与返回结构。
//
// 为什么要抽出来：文章列表、评论列表、后台列表都要分页，
// 如果每个 handler 各写一套「默认 20、上限 100」，改规则时就得改十几个地方，
// 而且前端拿到的字段名还可能不一致。
package pagination

const (
	// DefaultPageSize 不传 page_size 时的默认条数
	DefaultPageSize = 20
	// MaxPageSize 单页上限，防止有人传 page_size=999999 把数据库拖死
	MaxPageSize = 100
)

// Query 是每个列表接口都有的两个查询参数。
// 用匿名嵌入的方式放进各自的 Query 结构体里，Gin 会一起绑定和校验。
type Query struct {
	Page     int `form:"page" binding:"omitempty,gte=1"`
	PageSize int `form:"page_size" binding:"omitempty,gte=1,lte=100"`
}

// Normalize 补默认值并夹紧上限。
// 放在这里而不是各 handler 里，保证所有接口行为一致。
func (q *Query) Normalize() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = DefaultPageSize
	}
	if q.PageSize > MaxPageSize {
		q.PageSize = MaxPageSize
	}
}

// Offset 计算 SQL 的 OFFSET
func (q Query) Offset() int {
	return (q.Page - 1) * q.PageSize
}

// Result 是所有分页接口统一的返回结构
type Result struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// NewResult 组装分页返回
func NewResult(list any, total int64, q Query) Result {
	return Result{List: list, Total: total, Page: q.Page, PageSize: q.PageSize}
}
