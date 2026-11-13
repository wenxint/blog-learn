<script setup>
// 后台布局：左侧菜单 + 顶栏（当前用户 / 退出）+ 内容区
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const menus = [
  { name: 'admin-articles', label: '文章管理', to: '/admin/articles', icon: '📝' },
  { name: 'admin-comments', label: '评论审核', to: '/admin/comments', icon: '💬' },
]

async function handleLogout() {
  try {
    await ElMessageBox.confirm('确定退出登录？', '提示', { type: 'warning' })
  } catch {
    return // 用户点了取消
  }
  userStore.logout()
  ElMessage.success('已退出')
  router.replace('/admin/login')
}

onMounted(() => {
  // 刷新页面后同步一次用户资料（localStorage 里那份可能过期）
  userStore.loadProfile()
})
</script>

<template>
  <div class="admin">
    <aside class="sidebar">
      <div class="brand">
        <span class="brand-mark">B</span>
        <span>博客后台</span>
      </div>

      <nav class="menu">
        <RouterLink
          v-for="item in menus"
          :key="item.name"
          :to="item.to"
          class="menu-item"
          :class="{ 'menu-item--active': route.name === item.name }"
        >
          <span class="icon">{{ item.icon }}</span>
          {{ item.label }}
        </RouterLink>

        <RouterLink to="/" class="menu-item menu-item--back">
          <span class="icon">↩</span>
          返回前台
        </RouterLink>
      </nav>

      <div class="sidebar-foot muted">Go + Vue3 手搓</div>
    </aside>

    <div class="main">
      <header class="topbar">
        <h2 class="page-title">{{ route.meta.title || '后台' }}</h2>

        <div class="user">
          <span class="avatar">{{ (userStore.user?.nickname || 'A').slice(0, 1) }}</span>
          <span class="name">{{ userStore.user?.nickname || '未登录' }}</span>
          <span class="role-tag" :class="`role-tag--${userStore.user?.role || 'author'}`">
            {{ userStore.user?.role === 'admin' ? '管理员' : '作者' }}
          </span>
          <el-button link type="danger" @click="handleLogout">退出</el-button>
        </div>
      </header>

      <main class="content">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped>
.admin {
  display: flex;
  min-height: 100vh;
  background: #f5f6fa;
}

.sidebar {
  width: 208px;
  flex-shrink: 0;
  background: #1e2130;
  color: #c8ccd8;
  display: flex;
  flex-direction: column;
  padding: 20px 14px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #fff;
  font-weight: 600;
  padding: 0 8px 20px;
}

.brand-mark {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  background: linear-gradient(135deg, #4f7cff, #7b5cff);
  display: grid;
  place-items: center;
  font-size: 15px;
}

.menu {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  text-decoration: none;
  color: #c8ccd8;
  font-size: 14px;
  transition: all 0.15s;
}

.menu-item:hover {
  background: rgba(255, 255, 255, 0.06);
  color: #fff;
}

.menu-item--active {
  background: rgba(79, 124, 255, 0.18);
  color: #fff;
}

.menu-item--back {
  margin-top: 16px;
  color: #8b91a3;
  font-size: 13px;
}

.icon {
  font-size: 15px;
}

.sidebar-foot {
  margin-top: auto;
  padding: 0 12px;
  font-size: 12px;
  color: #5f6577;
}

.main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.topbar {
  height: 62px;
  background: #fff;
  border-bottom: 1px solid #e8eaef;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 26px;
}

.page-title {
  margin: 0;
  font-size: 17px;
}

.user {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}

.avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--primary-soft);
  color: var(--primary);
  display: grid;
  place-items: center;
  font-weight: 600;
}

.name {
  color: var(--text);
  font-weight: 600;
}

.role-tag {
  font-size: 11px;
  padding: 1px 7px;
  border-radius: 4px;
}

.role-tag--admin {
  background: #fff1e6;
  color: #d46b08;
}

.role-tag--author {
  background: #eef2ff;
  color: #4f7cff;
}

.content {
  padding: 22px 26px 40px;
  flex: 1;
}
</style>
