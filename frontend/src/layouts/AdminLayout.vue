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
        <span class="brand-mark">~/</span>
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
  /* 前台有一层 fixed 的网格纹理（body::before），后台表格需要干净底色，
     用 z-index + 实色背景把纹理挡在后面，避免格线和表格线交错。 */
  position: relative;
  z-index: 1;
  display: flex;
  min-height: 100vh;
  background: var(--bg);
}

.sidebar {
  width: 13rem;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  padding: 1.5rem 1rem;
  background: var(--panel);
  border-right: 1px solid var(--line);
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  margin-bottom: 1rem;
  padding: 0 0.5rem 1.25rem;
  border-bottom: 1px solid var(--line);
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--ink);
}

.brand-mark {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 0.375rem;
  background: var(--amber);
  color: var(--bg);
  display: grid;
  place-items: center;
  font-size: 0.75rem;
  font-weight: 700;
}

.menu {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.625rem 0.75rem;
  border-radius: 0.375rem;
  font-family: var(--font-sans);
  font-size: 0.8125rem;
  color: var(--ink-2);
  transition: background 0.18s ease, color 0.18s ease;
}

.menu-item:hover {
  background: rgba(255, 255, 255, 0.04);
  color: var(--ink);
}

.menu-item--active {
  background: var(--amber-soft);
  color: var(--amber);
}

.menu-item--back {
  margin-top: 1rem;
  font-size: 0.75rem;
  color: var(--ink-3);
}

.icon {
  font-size: 0.875rem;
}

.sidebar-foot {
  margin-top: auto;
  padding: 0 0.75rem;
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  color: var(--ink-3);
}

.main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.topbar {
  height: 3.75rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.75rem;
  background: rgba(11, 15, 22, 0.6);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--line);
}

.page-title {
  margin: 0;
  font-size: 1rem;
}

.user {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  font-family: var(--font-sans);
  font-size: 0.75rem;
}

.avatar {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: var(--amber-soft);
  color: var(--amber);
  font-family: var(--font-mono);
  font-weight: 600;
}

.name {
  color: var(--ink);
  font-weight: 600;
}

.role-tag {
  padding: 0.0625rem 0.4375rem;
  border-radius: 0.1875rem;
  font-family: var(--font-mono);
  font-size: 0.625rem;
}

.role-tag--admin {
  background: var(--amber-soft);
  color: var(--amber);
}

.role-tag--author {
  background: rgba(88, 199, 192, 0.12);
  color: var(--cyan);
}

.content {
  flex: 1;
  min-width: 0;
  padding: 1.5rem 1.75rem 3rem;
}

/* 后台列表页统一放进一块实色面板：表格、工具栏和分页有明确边界 */
.content :deep(.page) {
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 0.75rem;
  padding: 1rem 1.125rem;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
}

/* ───────── 窄屏：侧栏收成顶部横条 ───────── */
@media (max-width: 48rem) {
  .admin {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
    flex-direction: row;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 0.875rem;
    border-right: none;
    border-bottom: 1px solid var(--line);
    position: sticky;
    top: 0;
    z-index: 20;
    background: var(--bg);
  }

  .brand {
    margin: 0;
    padding: 0;
    border-bottom: none;
    flex-shrink: 0;
  }

  .brand > span:last-child {
    display: none;
  }

  .menu {
    flex-direction: row;
    gap: 0.375rem;
    flex: 1;
    min-width: 0;
    overflow-x: auto;
    scrollbar-width: none;
  }

  .menu::-webkit-scrollbar {
    display: none;
  }

  .menu-item {
    padding: 0.5rem 0.75rem;
    font-size: 0.75rem;
    white-space: nowrap;
  }

  .menu-item--back {
    margin-top: 0;
    border: 1px solid var(--line);
  }

  .sidebar-foot {
    display: none;
  }

  .topbar {
    height: auto;
    padding: 0.625rem 0.875rem;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .page-title {
    font-size: 0.9375rem;
  }

  .user {
    min-width: 0;
  }

  .name {
    max-width: 5rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .content {
    padding: 0.875rem 0.75rem 2rem;
  }

  .content :deep(.page) {
    padding: 0.75rem;
  }
}
</style>
