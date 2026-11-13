import request from './request'

// ---------- 文章 ----------

export function fetchAdminArticles(params = {}) {
  return request.get('/admin/articles', { params })
}

export function fetchAdminArticle(id) {
  return request.get(`/admin/articles/${id}`)
}

export function createArticle(data) {
  return request.post('/admin/articles', data)
}

export function updateArticle(id, data) {
  return request.put(`/admin/articles/${id}`, data)
}

export function deleteArticle(id) {
  return request.delete(`/admin/articles/${id}`)
}

// ---------- 分类 / 标签 ----------

export function createCategory(data) {
  return request.post('/admin/categories', data)
}

export function updateCategory(id, data) {
  return request.put(`/admin/categories/${id}`, data)
}

export function deleteCategory(id) {
  return request.delete(`/admin/categories/${id}`)
}

export function createTag(data) {
  return request.post('/admin/tags', data)
}

export function deleteTag(id) {
  return request.delete(`/admin/tags/${id}`)
}

// ---------- 评论 ----------

export function fetchAdminComments(params = {}) {
  return request.get('/admin/comments', { params })
}

export function setCommentStatus(id, status) {
  return request.put(`/admin/comments/${id}/status`, { status })
}

export function deleteComment(id) {
  return request.delete(`/admin/comments/${id}`)
}

// ---------- 上传 ----------

/**
 * 上传图片。必须用 FormData 传，不能直接 JSON——
 * 文件是二进制，只能走 multipart/form-data。
 */
export function uploadImage(file) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/admin/upload', formData)
}
