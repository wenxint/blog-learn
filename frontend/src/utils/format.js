// 小工具函数：日期格式化。

/**
 * 把后端返回的 RFC3339 时间（2026-11-13T14:07:13+08:00）截成 2026-11-13
 * 后端用 time.Time 序列化时带时区，前端只需要展示日期部分。
 */
export function formatDate(value) {
  if (!value) return '未发布'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/**
 * 相对时间：刚刚 / 3 分钟前 / 2 天前。列表页用它比绝对时间更友好。
 */
export function formatRelative(value) {
  if (!value) return ''
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value

  const diff = Date.now() - d.getTime()
  const minute = 60 * 1000
  const hour = 60 * minute
  const day = 24 * hour

  if (diff < minute) return '刚刚'
  if (diff < hour) return `${Math.floor(diff / minute)} 分钟前`
  if (diff < day) return `${Math.floor(diff / hour)} 小时前`
  if (diff < 30 * day) return `${Math.floor(diff / day)} 天前`
  return formatDate(value)
}
