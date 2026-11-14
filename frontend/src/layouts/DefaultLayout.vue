<script setup>
// 前台通用布局：终端风格的头部 + 内容区 + 页脚。
// 头部的品牌用 ~/ 作为提示符，和「工程手记」的概念呼应。
import { ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

const menuOpen = ref(false)
const route = useRoute()

// 路由变化后收起移动端菜单，避免点完链接还挡着内容
watch(
  () => route.fullPath,
  () => {
    menuOpen.value = false
  },
)
</script>

<template>
  <div class="layout">
    <header class="site-header">
      <div class="container header-inner">
        <RouterLink to="/" class="brand" aria-label="返回首页">
          <span class="brand-prompt">~/</span>
          <span class="brand-name">youzhichong</span>
        </RouterLink>

        <button
          class="menu-btn"
          type="button"
          :aria-expanded="menuOpen"
          aria-label="切换导航菜单"
          @click="menuOpen = !menuOpen"
        >
          <span class="menu-icon" :class="{ 'menu-icon--open': menuOpen }"></span>
        </button>

        <nav class="nav" :class="{ 'nav--open': menuOpen }">
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
      <div class="container footer-inner">
        <span class="mono-tag">陶文鑫 · youzhichong.xyz</span>
        <span class="mono-tag">Go + Vue3 手搓</span>
        <span class="mono-tag">© 2026</span>
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
  z-index: 50;
  /* 不用半透明：滚动时正文会从导航栏底下透出来，读起来很脏。
     纯色底 + 一条细线，滚动时内容被完全挡住。 */
  background: var(--bg);
  border-bottom: 1px solid var(--line);
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.35);
}

.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 4rem;
}

.brand {
  display: flex;
  align-items: baseline;
  gap: 0.25rem;
  font-family: var(--font-mono);
}

.brand-prompt {
  color: var(--amber);
  font-size: 0.9375rem;
  font-weight: 600;
}

.brand-name {
  color: var(--ink);
  font-size: 0.9375rem;
  letter-spacing: -0.01em;
}

.nav {
  display: flex;
  align-items: center;
  gap: 1.75rem;
}

.nav-link {
  position: relative;
  padding: 0.25rem 0;
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  color: var(--ink-2);
  transition: color 0.18s ease;
}

/* 下划线从左侧展开，只在 hover / 当前项出现 */
.nav-link::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: -0.125rem;
  height: 1px;
  background: var(--amber);
  transform: scaleX(0);
  transform-origin: left;
  transition: transform 0.22s cubic-bezier(0.2, 0.7, 0.2, 1);
}

.nav-link:hover {
  color: var(--ink);
}

.nav-link:hover::after,
.nav-link.router-link-exact-active::after {
  transform: scaleX(1);
}

.nav-link.router-link-exact-active {
  color: var(--amber);
}

.nav-link--admin {
  padding: 0.3rem 0.75rem;
  border: 1px solid var(--line-strong);
  border-radius: 999px;
}

.nav-link--admin::after {
  display: none;
}

.nav-link--admin:hover {
  border-color: var(--amber);
  color: var(--amber);
}

.site-main {
  flex: 1;
  padding: 4rem 0 6rem;
  position: relative;
  z-index: 1;
}

.site-footer {
  border-top: 1px solid var(--line);
  padding: 2rem 0;
  position: relative;
  z-index: 1;
}

.footer-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}

.footer-inner .mono-tag {
  color: var(--ink-3);
}

/* 汉堡按钮：两条线，打开时交叉成 × */
.menu-btn {
  display: none;
  width: 2.5rem;
  height: 2.5rem;
  padding: 0;
  background: transparent;
  border: 1px solid var(--line-strong);
  border-radius: var(--radius);
  cursor: pointer;
  position: relative;
}

.menu-icon,
.menu-icon::before {
  position: absolute;
  left: 50%;
  width: 1.125rem;
  height: 1px;
  background: var(--ink);
  transition: transform 0.22s ease;
  transform: translateX(-50%);
}

.menu-icon {
  top: calc(50% - 0.1875rem);
}

.menu-icon::before {
  content: '';
  top: 0.375rem;
  left: 0;
  transform: none;
}

.menu-icon--open {
  transform: translateX(-50%) rotate(45deg);
  top: 50%;
}

.menu-icon--open::before {
  transform: rotate(-90deg);
  top: 0;
}

/* ───────── 窄屏：导航改为整屏覆盖菜单 ───────── */
@media (max-width: 48rem) {
  .header-inner {
    height: 3.5rem;
  }

  .menu-btn {
    display: block;
  }

  /* 注意：.site-header 有 backdrop-filter，它会让 fixed 子元素相对它定位。
     所以这里必须用 absolute（相对 header 的 bottom），不能用 fixed，
     否则菜单高度会被压成 0、内容溢出到正文上。
     动画也不能用 opacity，否则菜单背景会半透明、正文透出来。 */
  .nav {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    max-height: calc(100vh - 3.5rem);
    overflow-y: auto;
    flex-direction: column;
    align-items: stretch;
    gap: 0;
    padding: 1.5rem 1.125rem 2rem;
    background: var(--bg);
    border-top: 1px solid var(--line);
    box-shadow: 0 1.5rem 3rem rgba(0, 0, 0, 0.6);
    visibility: hidden;
    transform: translateY(-0.5rem);
    transition: transform 0.24s ease, visibility 0.24s ease;
  }

  .nav--open {
    visibility: visible;
    transform: none;
  }

  .nav-link {
    font-size: 1.125rem;
    padding: 1rem 0;
    width: 100%;
    border-bottom: 1px solid var(--line);
  }

  .nav-link::after {
    display: none;
  }

  .nav-link--admin {
    margin-top: 1.5rem;
    text-align: center;
    padding: 0.75rem 0;
    border-radius: 999px;
  }

  .site-main {
    padding: 2.5rem 0 4rem;
  }

  .footer-inner {
    justify-content: flex-start;
    gap: 0.5rem 1rem;
  }
}
</style>
