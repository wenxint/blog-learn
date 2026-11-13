import { createRouter, createWebHistory } from 'vue-router'

// 路由表组织方式：用「布局」作为父路由，页面作为子路由。
// 这样前台页面自动套 DefaultLayout，后台页面自动套 AdminLayout，
// 页面组件里不用重复写头部和侧边栏。
const routes = [
  {
    path: '/',
    component: () => import('../layouts/DefaultLayout.vue'),
    children: [
      { path: '', name: 'home', component: () => import('../views/Home.vue'), meta: { title: '首页' } },
      {
        path: 'article/:id',
        name: 'article-detail',
        component: () => import('../views/ArticleDetail.vue'),
        meta: { title: '文章详情' },
      },
      { path: 'archive', name: 'archive', component: () => import('../views/Archive.vue'), meta: { title: '归档' } },
      { path: 'tags', name: 'tags', component: () => import('../views/Tags.vue'), meta: { title: '标签' } },
      { path: 'about', name: 'about', component: () => import('../views/About.vue'), meta: { title: '关于' } },
    ],
  },
  {
    path: '/admin/login',
    name: 'admin-login',
    component: () => import('../views/admin/Login.vue'),
    meta: { title: '后台登录' },
  },
  {
    path: '/admin',
    component: () => import('../layouts/AdminLayout.vue'),
    meta: { requiresAuth: true }, // 整组需要登录
    children: [
      { path: '', redirect: '/admin/articles' },
      {
        path: 'articles',
        name: 'admin-articles',
        component: () => import('../views/admin/ArticleManage.vue'),
        meta: { title: '文章管理' },
      },
      {
        path: 'articles/new',
        name: 'admin-article-new',
        component: () => import('../views/admin/ArticleEdit.vue'),
        meta: { title: '新建文章' },
      },
      {
        path: 'articles/:id/edit',
        name: 'admin-article-edit',
        component: () => import('../views/admin/ArticleEdit.vue'),
        meta: { title: '编辑文章' },
      },
      {
        path: 'comments',
        name: 'admin-comments',
        component: () => import('../views/admin/CommentManage.vue'),
        meta: { title: '评论审核' },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('../views/NotFound.vue'),
    meta: { title: '页面不存在' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

// 全局前置守卫：进入需要登录的页面时检查 token。
// 注意这只是「前端体验」层面的拦截，真正的安全靠后端每个接口校验 token——
// 前端守卫能被绕过（改 localStorage 就行），后端鉴权不能。
router.beforeEach((to) => {
  if (!to.meta.requiresAuth) return true

  const token = localStorage.getItem('token')
  if (token) return true

  // 记住原本想去的地址，登录后跳回去
  return { name: 'admin-login', query: { redirect: to.fullPath } }
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · 我的博客` : '我的博客'
})

export default router
