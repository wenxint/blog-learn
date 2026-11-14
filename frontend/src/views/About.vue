<script setup>
// 关于页：静态介绍。真实项目里这块内容通常从后端 /site/info 读取，
// 我们这里先写死，等需要在线编辑时再对接。
const stack = [
  { name: 'Go 1.26', desc: '后端语言' },
  { name: 'Gin', desc: 'HTTP 框架' },
  { name: 'GORM', desc: 'ORM / 数据访问' },
  { name: 'MySQL 8.4', desc: '关系型数据库' },
  { name: 'JWT', desc: '无状态认证' },
  { name: 'Vue 3', desc: '前端框架' },
  { name: 'Vite', desc: '构建工具' },
  { name: 'Element Plus', desc: 'UI 组件库' },
]

const lessons = [
  'Go 标准库起 HTTP 服务',
  'MySQL 建库建表与索引',
  'Gin 路由 / 中间件 / 统一响应',
  '配置加载与连接池',
  'GORM 模型与真实查询',
  'CRUD 与事务',
  'handler / service / repository 分层',
  'bcrypt 密码哈希',
  'JWT 签发与鉴权中间件',
  '分页 / 分类 / 标签 / 归档',
  '评论楼中楼与审核',
  '图片上传与静态资源',
  'Vue3 路由与请求封装',
  'Markdown 渲染与后台管理',
]
</script>

<template>
  <div class="container">
    <header class="page-head">
      <h1>关于</h1>
      <p class="muted">一个边学边做的博客系统</p>
    </header>

    <section class="card block">
      <h2>这是什么</h2>
      <p>
        这是一个从零手写的前后端分离博客：后端用 Go（Gin + GORM + MySQL），
        前端用 Vue3 + Vite。写着学习笔记，也顺便把后端该会的东西都过了一遍。
      </p>
      <p>
        后端不是照着模板生成的，每一层（配置、连接池、模型、仓储、业务、接口、中间件、鉴权）
        都手写过一遍，踩的坑记在 <code>docs/notes/踩坑记录.md</code> 里。
      </p>
    </section>

    <section class="card block">
      <h2>技术栈</h2>
      <div class="stack">
        <div v-for="item in stack" :key="item.name" class="stack-item">
          <span class="stack-name">{{ item.name }}</span>
          <span class="muted">{{ item.desc }}</span>
        </div>
      </div>
    </section>

    <section class="card block">
      <h2>做过的内容</h2>
      <ul class="lessons">
        <li v-for="(lesson, i) in lessons" :key="i">
          <span class="index">{{ String(i + 1).padStart(2, '0') }}</span>
          {{ lesson }}
        </li>
      </ul>
    </section>

    <section class="card block">
      <h2>联系</h2>
      <p class="muted">
        这个站点主要用于自己学习，欢迎交流。
        源码结构见项目根目录 README。
      </p>
    </section>
  </div>
</template>

<style scoped>
.page-head {
  margin-bottom: 2.5rem;
}

.page-head h1 {
  margin: 0 0 0.5rem;
  font-size: clamp(1.75rem, 3.5vw, 2.5rem);
  letter-spacing: -0.02em;
}

.page-head p {
  margin: 0;
  color: var(--ink-2);
}

.block {
  padding: 1.75rem 0;
  border-bottom: 1px solid var(--line);
  /* 覆盖 .card 带来的面板底色与边框，保持「手记」的无卡片风格 */
  background: none;
  border-left: none;
  border-right: none;
  border-top: none;
  border-radius: 0;
}

.block:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.block h2 {
  margin: 0 0 0.875rem;
  font-size: 1.125rem;
}

.block p {
  max-width: 46rem;
  margin: 0 0 0.75rem;
  color: var(--ink-2);
  line-height: 1.85;
}

code {
  font-family: var(--font-mono);
  font-size: 0.875em;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--line);
  border-radius: 0.25rem;
  padding: 0.1em 0.35em;
  color: #e9c98d;
}

.stack {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(11rem, 1fr));
  gap: 0.75rem;
}

.stack-item {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  padding: 0.75rem 0.875rem;
  border: 1px solid var(--line);
  border-radius: 0.375rem;
}

.stack-item > span:last-child {
  font-family: var(--font-sans);
  font-size: 0.75rem;
  color: var(--ink-3);
}

.stack-name {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--ink);
}

.lessons {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(15rem, 1fr));
  gap: 0.5rem 1.5rem;
}

.lessons li {
  display: flex;
  gap: 0.5rem;
  font-size: 0.9375rem;
  color: var(--ink-2);
}

.index {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--ink-3);
  padding-top: 0.15rem;
}

@media (max-width: 48rem) {
  .stack {
    grid-template-columns: 1fr;
  }

  .lessons {
    grid-template-columns: 1fr;
  }
}
</style>
