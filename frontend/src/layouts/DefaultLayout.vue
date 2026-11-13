<script setup>
// 前台通用布局：头部导航 + 内容区 + 页脚。
// 想加「后台布局」就再写一个 AdminLayout，两条路由树各用各的。
import { RouterLink } from 'vue-router'
</script>

<template>
  <div class="layout">
    <header class="site-header">
      <div class="container header-inner">
        <RouterLink to="/" class="logo">
          <span class="logo-mark">B</span>
          <span class="logo-text">我的博客</span>
        </RouterLink>

        <nav class="nav">
          <RouterLink to="/" class="nav-link">首页</RouterLink>
          <!-- 下面这些页面后续课程再做，先占位，避免导航看起来空 -->
          <span class="nav-link nav-link--muted">归档</span>
          <span class="nav-link nav-link--muted">标签</span>
          <span class="nav-link nav-link--muted">关于</span>
        </nav>
      </div>
    </header>

    <main class="site-main">
      <!-- slot 是「占位插槽」：App.vue 里放进去的 RouterView 会填到这里 -->
      <slot />
    </main>

    <footer class="site-footer">
      <div class="container">
        <p>© 2026 我的博客 · 用 Go + Vue3 手搓</p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.site-header {
  position: sticky;
  top: 0;
  z-index: 10;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--border);
}

.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: var(--text);
  font-weight: 600;
  font-size: 17px;
}

.logo-mark {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  background: linear-gradient(135deg, #4f7cff, #7b5cff);
  color: #fff;
  display: grid;
  place-items: center;
  font-size: 16px;
}

.nav {
  display: flex;
  gap: 22px;
}

.nav-link {
  text-decoration: none;
  color: var(--text-secondary);
  font-size: 14px;
  transition: color 0.15s;
}

.nav-link:hover {
  color: var(--primary);
}

.nav-link.router-link-active {
  color: var(--primary);
  font-weight: 600;
}

.nav-link--muted {
  cursor: not-allowed;
  opacity: 0.45;
}

.site-main {
  flex: 1;
  padding: 32px 0 64px;
}

.site-footer {
  border-top: 1px solid var(--border);
  padding: 24px 0;
  color: var(--text-tertiary);
  font-size: 13px;
  text-align: center;
}
</style>
