<script setup>
// 首页：文章列表 + 分类筛选 + 搜索 + 分页
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchArticles, fetchCategories } from '../api/article'

const route = useRoute()
const router = useRouter()

const articles = ref([])
const categories = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const categoryId = ref(0)
const tagId = ref(0)
const month = ref('')
const tagName = ref('')
const loading = ref(false)

// 列表左侧的日期列：日 + 月，做成真实的时间序列索引
function dayOf(value) {
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? '--' : String(d.getDate()).padStart(2, '0')
}

function monthOf(value) {
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? '' : `${d.getMonth() + 1} 月`
}

// 从地址栏读取筛选条件。
// 归档页/标签页跳过来时会带 ?month=2026-11 或 ?tag_id=2，
// 这样筛选状态可以分享、刷新不丢。
function syncFromQuery() {
  const q = route.query
  categoryId.value = Number(q.category_id) || 0
  tagId.value = Number(q.tag_id) || 0
  month.value = q.month || ''
  keyword.value = q.keyword || ''
  tagName.value = q.tag_name || ''
}

// 把当前筛选同步回地址栏（用 replace，避免每次筛选都往历史里塞一条）
function syncToQuery() {
  const query = {}
  if (categoryId.value) query.category_id = categoryId.value
  if (tagId.value) query.tag_id = tagId.value
  if (month.value) query.month = month.value
  if (keyword.value) query.keyword = keyword.value
  if (tagName.value) query.tag_name = tagName.value
  router.replace({ path: '/', query })
}

async function loadArticles() {
  loading.value = true
  try {
    const data = await fetchArticles({
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value || undefined,
      category_id: categoryId.value || undefined,
      tag_id: tagId.value || undefined,
      month: month.value || undefined,
    })
    articles.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

async function loadCategories() {
  const data = await fetchCategories()
  categories.value = data.list || []
}

// 切换筛选条件时回到第 1 页，否则会出现「第 5 页没有数据」
function resetAndLoad() {
  page.value = 1
  loadArticles()
}

watch([keyword, categoryId], () => {
  tagId.value = 0
  tagName.value = ''
  month.value = ''
  resetAndLoad()
  syncToQuery()
})

watch(page, loadArticles)

function clearFilters() {
  categoryId.value = 0
  tagId.value = 0
  month.value = ''
  keyword.value = ''
  tagName.value = ''
  resetAndLoad()
  syncToQuery()
}

const hasFilter = computed(
  () => categoryId.value || tagId.value || month.value || keyword.value,
)

onMounted(() => {
  syncFromQuery()
  loadArticles()
  loadCategories()
})
</script>

<template>
  <div class="container">
    <!-- Hero：全站唯一「大胆」的地方，其余部分一律克制 -->
    <section class="hero">
      <div class="hero-copy">
        <h1 class="hero-title rise">
          <span>写代码</span>
          <span class="hero-title-2">也写明白</span>
        </h1>
        <p class="hero-sub rise rise-2">
          用 Go 和 Vue3 从零搭的这个博客。学到的、踩过的坑，都留在这里。
        </p>
        <div class="hero-stats rise rise-3">
          <span class="mono-tag">{{ total }} 篇笔记</span>
          <span class="mono-tag">{{ categories.length }} 个分类</span>
        </div>
      </div>

      <!-- 终端窗口：装饰性的，对读屏隐藏 -->
      <div class="terminal rise rise-2" aria-hidden="true">
        <div class="terminal-bar">
          <span class="tdot"></span>
          <span class="tdot"></span>
          <span class="tdot"></span>
          <span class="terminal-path">~/blog</span>
        </div>
        <div class="terminal-body">
          <p><span class="prompt">$</span> whoami</p>
          <p class="out">后端学习中</p>
          <p><span class="prompt">$</span> ls notes/</p>
          <p class="out">01-architecture.md</p>
          <p class="out">05-auth-jwt.md</p>
          <p class="out">10-deploy.md</p>
          <p><span class="prompt">$</span><span class="caret">▍</span></p>
        </div>
      </div>
    </section>

    <!-- 筛选 -->
    <section class="filters">
      <div class="filter-tabs">
        <button class="tab" :class="{ 'tab--active': categoryId === 0 }" @click="categoryId = 0">
          全部
        </button>
        <button
          v-for="cat in categories"
          :key="cat.id"
          class="tab"
          :class="{ 'tab--active': categoryId === cat.id }"
          @click="categoryId = cat.id"
        >
          {{ cat.name }}
          <span class="tab-count">{{ cat.article_count }}</span>
        </button>
      </div>

      <ElInput v-model="keyword" placeholder="搜索标题…" clearable class="search" />
    </section>

    <div v-if="hasFilter" class="active-filters">
      <span class="muted">筛选</span>
      <span v-if="month" class="filter-pill">{{ month }}</span>
      <span v-if="tagName || tagId" class="filter-pill">标签 {{ tagName || `#${tagId}` }}</span>
      <span v-if="keyword" class="filter-pill">关键词 {{ keyword }}</span>
      <button class="clear-btn" @click="clearFilters">清除</button>
    </div>

    <ElSkeleton v-if="loading" :rows="5" animated class="skeleton" />

    <template v-else>
      <ElEmpty v-if="articles.length === 0" description="没有找到文章" :image-size="80" />

      <ul v-else class="post-list">
        <li v-for="item in articles" :key="item.id" class="post-item">
          <RouterLink :to="`/article/${item.id}`" class="post-link">
            <div class="post-date" aria-hidden="true">
              <span class="post-day">{{ dayOf(item.published_at) }}</span>
              <span class="post-month">{{ monthOf(item.published_at) }}</span>
            </div>

            <div class="post-main">
              <h2 class="post-title">
                <span v-if="item.is_top === 1" class="pin">置顶</span>
                {{ item.title }}
              </h2>
              <p class="post-summary">{{ item.summary || '（暂无摘要）' }}</p>

              <div class="post-meta">
                <span v-if="item.category" class="post-cat">{{ item.category.name }}</span>
                <span v-for="tag in item.tags || []" :key="tag.id" class="post-tag">
                  {{ tag.name }}
                </span>
                <span class="post-views">{{ item.view_count }} 次阅读</span>
              </div>
            </div>
          </RouterLink>
        </li>
      </ul>

      <div v-if="total > pageSize" class="pagination">
        <ElPagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          :pager-count="5"
          layout="prev, pager, next"
          background
        />
      </div>
    </template>
  </div>
</template>

<style scoped>
/* ───────── Hero ───────── */
.hero {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(0, 0.85fr);
  gap: 3.5rem;
  align-items: center;
  padding: 1.5rem 0 4.5rem;
}

.hero-title {
  margin: 0 0 1.25rem;
  font-size: clamp(2.25rem, 5.2vw, 3.75rem);
  line-height: 1.08;
  letter-spacing: -0.03em;
  display: flex;
  flex-direction: column;
}

/* 第二行用点缀色，把「写明白」这三个字变成主张 */
.hero-title-2 {
  color: var(--amber);
}

.hero-sub {
  margin: 0 0 1.75rem;
  max-width: 30rem;
  color: var(--ink-2);
  font-size: 1.0625rem;
  line-height: 1.8;
}

.hero-stats {
  display: flex;
  gap: 1.5rem;
}

.hero-stats .mono-tag {
  color: var(--ink-3);
}

/* ───────── 终端窗口 ───────── */
.terminal {
  background: var(--panel);
  border: 1px solid var(--line-strong);
  border-radius: 0.75rem;
  overflow: hidden;
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  box-shadow: 0 1.5rem 4rem rgba(0, 0, 0, 0.45);
}

.terminal-bar {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.625rem 0.875rem;
  border-bottom: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.02);
}

.tdot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  background: var(--ink-3);
  opacity: 0.5;
}

.terminal-path {
  margin-left: auto;
  color: var(--ink-3);
  font-size: 0.6875rem;
}

.terminal-body {
  padding: 1rem 1.125rem 1.25rem;
  line-height: 1.9;
}

.terminal-body p {
  margin: 0;
  color: var(--ink-2);
}

.prompt {
  color: var(--amber);
  margin-right: 0.5rem;
}

.out {
  color: var(--ink-3);
  padding-left: 0.875rem;
}

.caret {
  color: var(--ink);
  animation: blink 1.2s steps(1) infinite;
}

@keyframes blink {
  50% {
    opacity: 0;
  }
}

/* ───────── 筛选 ───────── */
.filters {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1.5rem;
  padding-bottom: 0.875rem;
  border-bottom: 1px solid var(--line);
}

.filter-tabs {
  display: flex;
  gap: 1.5rem;
  overflow-x: auto;
  scrollbar-width: none;
}

.filter-tabs::-webkit-scrollbar {
  display: none;
}

.tab {
  position: relative;
  flex-shrink: 0;
  padding: 0.25rem 0 0.625rem;
  background: none;
  border: none;
  cursor: pointer;
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  color: var(--ink-3);
  transition: color 0.18s ease;
}

.tab:hover {
  color: var(--ink);
}

.tab--active {
  color: var(--amber);
}

.tab--active::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: -1px;
  height: 1px;
  background: var(--amber);
}

.tab-count {
  margin-left: 0.375rem;
  font-size: 0.6875rem;
  opacity: 0.55;
}

.search {
  width: 13rem;
  flex-shrink: 0;
}

/* ───────── 筛选状态 ───────── */
.active-filters {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  flex-wrap: wrap;
  padding: 0.875rem 0;
  font-family: var(--font-sans);
}

.filter-pill {
  font-size: 0.75rem;
  color: var(--amber);
  background: var(--amber-soft);
  border-radius: 999px;
  padding: 0.125rem 0.625rem;
}

.clear-btn {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  font-size: 0.75rem;
  color: var(--ink-3);
  text-decoration: underline;
  text-underline-offset: 0.125rem;
}

.clear-btn:hover {
  color: var(--ink);
}

.skeleton {
  padding: 1.5rem 0;
}

/* ───────── 文章清单 ───────── */
.post-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.post-item {
  border-bottom: 1px solid var(--line);
}

.post-link {
  display: grid;
  grid-template-columns: 4.5rem minmax(0, 1fr);
  gap: 1.5rem;
  padding: 1.75rem 0;
  align-items: start;
  transition: opacity 0.18s ease;
}

/* 日期列：真实的序列索引，不是装饰性编号 */
.post-date {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  font-family: var(--font-mono);
  padding-top: 0.25rem;
  border-left: 1px solid transparent;
  padding-left: 0.75rem;
  transition: border-color 0.2s ease;
}

.post-day {
  font-size: 1.375rem;
  line-height: 1;
  color: var(--ink-2);
  transition: color 0.2s ease;
}

.post-month {
  font-size: 0.6875rem;
  color: var(--ink-3);
}

.post-link:hover .post-date {
  border-left-color: var(--amber);
}

.post-link:hover .post-day {
  color: var(--amber);
}

.post-title {
  margin: 0 0 0.5rem;
  font-size: 1.25rem;
  line-height: 1.45;
  color: var(--ink);
  transition: color 0.18s ease;
}

.post-link:hover .post-title {
  color: var(--amber);
}

.pin {
  font-size: 0.625rem;
  color: var(--bg);
  background: var(--amber);
  border-radius: 0.125rem;
  padding: 0.0625rem 0.3125rem;
  margin-right: 0.375rem;
  vertical-align: middle;
  font-weight: 600;
}

.post-summary {
  margin: 0 0 0.875rem;
  color: var(--ink-2);
  font-size: 0.9375rem;
  line-height: 1.75;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.post-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  font-family: var(--font-sans);
  font-size: 0.75rem;
  color: var(--ink-3);
}

.post-cat {
  color: var(--amber);
}

.post-tag {
  color: var(--ink-3);
}

.post-tag::before {
  content: '#';
  opacity: 0.5;
  margin-right: 0.125rem;
}

.post-views {
  margin-left: auto;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 2.5rem;
}

/* ───────── 窄屏 ───────── */
@media (max-width: 60rem) {
  .hero {
    grid-template-columns: 1fr;
    gap: 2.5rem;
    padding: 0.5rem 0 3rem;
  }

  /* 手机上没有空间，终端窗口降级为一条摘要行 */
  .terminal {
    display: none;
  }
}

@media (max-width: 48rem) {
  .filters {
    flex-direction: column;
    align-items: stretch;
    gap: 0.875rem;
  }

  .search {
    width: 100%;
  }

  .post-link {
    grid-template-columns: 1fr;
    gap: 0.625rem;
    padding: 1.375rem 0;
  }

  /* 窄屏日期改为行内小字，不再占一列 */
  .post-date {
    flex-direction: row;
    align-items: baseline;
    gap: 0.375rem;
    padding-left: 0;
    border-left: none;
  }

  .post-day {
    font-size: 0.875rem;
    color: var(--ink-3);
  }

  .post-month {
    font-size: 0.6875rem;
  }

  .post-title {
    font-size: 1.125rem;
  }

  .post-views {
    margin-left: 0;
  }
}
</style>
