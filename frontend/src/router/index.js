import { createRouter, createWebHistory } from 'vue-router'

// 路由表：一个文件看完全站有哪些页面。
// 全部用动态 import()：首屏只加载首页代码，其他页面按需下载（代码分割）。
const routes = [
  {
    path: '/',
    name: 'home',
    component: () => import('../views/Home.vue'),
    meta: { title: '首页' },
  },
  {
    path: '/article/:id',
    name: 'article-detail',
    component: () => import('../views/ArticleDetail.vue'),
    meta: { title: '文章详情' },
  },
  {
    // 兜底：未匹配的地址显示 404，而不是白屏
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('../views/NotFound.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  // 切换页面时滚回顶部，否则从长文章点回首页会停在中间
  scrollBehavior() {
    return { top: 0 }
  },
})

// 全局后置钩子：把页面标题同步到浏览器标签
router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · 我的博客` : '我的博客'
})

export default router
