import request from './request'

// 文章列表：params 支持 page / page_size / keyword / category_id / tag_id
export function fetchArticles(params = {}) {
  return request.get('/articles', { params })
}

// 文章详情（阅读数会自动 +1）
export function fetchArticleDetail(id) {
  return request.get(`/articles/${id}`)
}

// 分类列表（含每个分类的文章数）
export function fetchCategories() {
  return request.get('/categories')
}

// 标签列表（含每个标签的文章数）
export function fetchTags() {
  return request.get('/tags')
}

// 归档（按月份统计）
export function fetchArchives() {
  return request.get('/archives')
}

// 某篇文章的评论（树形：一级 + 回复）
export function fetchComments(articleId) {
  return request.get(`/articles/${articleId}/comments`)
}

// 发表评论（提交后进入待审核）
export function createComment(articleId, data) {
  return request.post(`/articles/${articleId}/comments`, data)
}
