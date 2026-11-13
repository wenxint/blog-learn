<script setup>
// 首页：文章列表 + 分类筛选 + 搜索 + 分页
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElSkeleton, ElPagination, ElInput, ElTag, ElEmpty, ElButton } from 'element-plus'
import { fetchArticles, fetchCategories } from '../api/article'
import { formatDate } from '../utils/format'

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
// 这里不直接 syncToQuery：watch 里改 URL 再触发 watch 会绕圈
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

// 清掉所有筛选（标签页/归档页跳过来时用得上）
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
    <section class="hero">
      <h1 class="hero-title">记录学习后端的过程</h1>
      <p class="hero-sub">Go · MySQL · Vue3 —— 从零手搓一个博客，边做边写。</p>
    </section>

    <!-- 当前生效的筛选条件（从归档/标签页跳过来时显示） -->
    <div v-if="hasFilter" class="active-filters">
      <span class="muted">当前筛选：</span>
      <ElTag v-if="month" type="warning" effect="plain" size="small">{{ month }}</ElTag>
      <ElTag v-if="tagName || tagId" type="info" effect="plain" size="small">
        标签：{{ tagName || `#${tagId}` }}
      </ElTag>
      <ElTag v-if="keyword" effect="plain" size="small">关键词：{{ keyword }}</ElTag>
      <ElButton link type="primary" size="small" @click="clearFilters">清除</ElButton>
    </div>

    <section class="filters">
      <div class="filter-group">
        <button class="chip" :class="{ 'chip--active': categoryId === 0 }" @click="categoryId = 0">
          全部
        </button>
        <button
          v-for="cat in categories"
          :key="cat.id"
          class="chip"
          :class="{ 'chip--active': categoryId === cat.id }"
          @click="categoryId = cat.id"
        >
          {{ cat.name }}
          <span class="chip-count">{{ cat.article_count }}</span>
        </button>
      </div>

      <ElInput v-model="keyword" placeholder="搜索文章标题…" clearable class="search" />
    </section>

    <ElSkeleton v-if="loading" :rows="4" animated class="skeleton" />

    <template v-else>
      <ElEmpty v-if="articles.length === 0" description="没有找到文章" :image-size="90" />

      <div v-else class="article-list">
        <RouterLink
          v-for="item in articles"
          :key="item.id"
          :to="`/article/${item.id}`"
          class="article-card card"
        >
          <div class="article-head">
            <span v-if="item.is_top === 1" class="top-badge">置顶</span>
            <h2 class="article-title">{{ item.title }}</h2>
          </div>

          <p class="article-summary">{{ item.summary || '（暂无摘要）' }}</p>

          <div class="article-meta">
            <span v-if="item.category" class="meta-category">{{ item.category.name }}</span>
            <ElTag
              v-for="tag in item.tags || []"
              :key="tag.id"
              size="small"
              type="info"
              effect="plain"
            >
              {{ tag.name }}
            </ElTag>

            <span class="meta-right">
              <span>{{ formatDate(item.published_at) }}</span>
              <span class="dot">·</span>
              <span>{{ item.view_count }} 次阅读</span>
            </span>
          </div>
        </RouterLink>
      </div>

      <div v-if="total > pageSize" class="pagination">
        <ElPagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          background
        />
      </div>
    </template>
  </div>
</template>

<style scoped>
.hero {
  padding: 8px 0 28px;
}

.hero-title {
  margin: 0 0 8px;
  font-size: 30px;
  letter-spacing: -0.5px;
}

.hero-sub {
  margin: 0;
  color: var(--text-secondary);
}

.active-filters {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}

.filters {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 22px;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.chip {
  border: 1px solid var(--border);
  background: #fff;
  color: var(--text-secondary);
  border-radius: 999px;
  padding: 6px 14px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.chip:hover {
  border-color: #c9d3ea;
}

.chip--active {
  background: var(--primary-soft);
  border-color: #b9c8ff;
  color: var(--primary);
  font-weight: 600;
}

.chip-count {
  opacity: 0.6;
  margin-left: 2px;
  font-size: 12px;
}

.search {
  width: 220px;
}

.skeleton {
  background: #fff;
  border-radius: var(--radius);
  padding: 24px;
}

.article-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.article-card {
  display: block;
  padding: 20px 22px;
  text-decoration: none;
}

.article-head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.top-badge {
  flex-shrink: 0;
  font-size: 11px;
  color: #fff;
  background: #ff7a45;
  border-radius: 4px;
  padding: 1px 6px;
}

.article-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  line-height: 1.45;
  color: var(--text);
}

.article-card:hover .article-title {
  color: var(--primary);
}

.article-summary {
  margin: 8px 0 14px;
  color: var(--text-secondary);
  font-size: 14px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.article-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  font-size: 12px;
  color: var(--text-tertiary);
}

.meta-category {
  color: var(--primary);
  background: var(--primary-soft);
  border-radius: 4px;
  padding: 2px 8px;
}

.meta-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 4px;
}

.dot {
  opacity: 0.5;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 32px;
}
</style>
