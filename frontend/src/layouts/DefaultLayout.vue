<script setup>
// 前台通用布局：头部导航 + 内容区 + 页脚。
// 注意这里的 RouterView：子路由（首页/详情/归档…）会渲染到 main 里。
import { RouterLink, RouterView } from 'vue-router'
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
          <RouterLink to="/archive" class="nav-link">归档</RouterLink>
          <RouterLink to="/tags" class="nav-link">标签</RouterLink>
          <RouterLink to="/about" class="nav-link">关于</RouterLink>
          <RouterLink to="/admin" class="nav-link nav-link--admin">后台</RouterLink>
        </nav>
      </div>
    </header>

    <main class="site-main">
      <RouterView />
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
  align-items: center;
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

.nav-link--admin {
  padding: 4px 12px;
  border: 1px solid var(--border);
  border-radius: 999px;
  font-size: 13px;
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
