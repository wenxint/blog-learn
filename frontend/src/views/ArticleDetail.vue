<script setup>
// 文章详情页：正文 + 评论
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createComment, fetchArticleDetail, fetchArticles, fetchComments } from '../api/article'
import { formatDate } from '../utils/format'
import { renderMarkdown } from '../utils/markdown'
// 代码高亮主题（github 风格浅色）
// 深色主题的高亮样式，和页面底色配套（浅色主题会让关键字在黑底上发暗）
import 'highlight.js/styles/github-dark.css'

const route = useRoute()
const router = useRouter()

const article = ref(null)
const comments = ref([])
// 同分类下的文章，用来在文末拼出「系列目录」。
// 不额外加接口：列表接口本来就支持按 category_id 筛选。
const series = ref([])
const loading = ref(true)
const submitting = ref(false)
const form = ref({ nickname: '', email: '', content: '' })

// Markdown → 净化后的 HTML。
// 用 computed：article 变化时自动重算，不用手动维护一份 html 状态。
const contentHtml = computed(() => renderMarkdown(article.value?.content || ''))

async function load() {
  loading.value = true
  try {
    article.value = await fetchArticleDetail(route.params.id)
    const data = await fetchComments(route.params.id)
    comments.value = data.list || []

    // 系列目录失败不影响正文阅读，所以单独 try/catch，不让整个页面 404
    if (article.value?.category_id) {
      try {
        const list = await fetchArticles({
          category_id: article.value.category_id,
          page_size: 100,
        })
        series.value = list.list || []
      } catch {
        series.value = []
      }
    }
  } catch {
    // 文章不存在或已删除时，回首页而不是停在白屏
    router.replace('/')
  } finally {
    loading.value = false
  }
}

async function submitComment() {
  if (!form.value.nickname.trim()) {
    ElMessage.warning('请填写昵称')
    return
  }
  if (!form.value.content.trim()) {
    ElMessage.warning('请填写评论内容')
    return
  }

  submitting.value = true
  try {
    await createComment(route.params.id, {
      nickname: form.value.nickname.trim(),
      email: form.value.email.trim(),
      content: form.value.content.trim(),
    })
    ElMessage.success('评论已提交，等待审核后展示')
    form.value.content = ''
    // 注意：不会立刻出现在列表里，因为后端默认进入待审核状态
  } finally {
    submitting.value = false
  }
}

// 阅读进度：随滚动更新顶部那条琥珀色细线。
// 只服务于「这篇有多长、读到哪了」，不承载装饰功能。
const progress = ref(0)

function updateProgress() {
  const el = document.documentElement
  const max = el.scrollHeight - el.clientHeight
  progress.value = max > 0 ? Math.min(1, el.scrollTop / max) : 0
}

onMounted(() => {
  load()
  updateProgress()
  window.addEventListener('scroll', updateProgress, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', updateProgress)
})
</script>

<template>
  <div class="read-progress" :style="{ transform: `scaleX(${progress})` }" aria-hidden="true"></div>
  <div class="container">
    <ElSkeleton v-if="loading" :rows="6" animated />

    <template v-else-if="article">
      <article class="article">
        <div class="article-head">
          <span v-if="article.is_top === 1" class="top-badge">置顶</span>
          <h1 class="article-title">{{ article.title }}</h1>
        </div>

        <div class="article-meta">
          <span v-if="article.author" class="author">{{ article.author.nickname }}</span>
          <span>{{ formatDate(article.published_at) }}</span>
          <span>{{ article.view_count }} 次阅读</span>
          <span v-if="article.category" class="meta-category">{{ article.category.name }}</span>
        </div>

        <div class="article-body">
          <!--
            v-html 会把字符串当成 HTML 插入，所以内容必须先过 DOMPurify 净化
            （见 utils/markdown.js），否则作者可以注入 <script> 偷用户 token。
          -->
          <div class="markdown-body" v-html="contentHtml"></div>
        </div>

        <div v-if="(article.tags || []).length" class="article-tags">
          <ElTag v-for="tag in article.tags" :key="tag.id" size="small" effect="plain">
            {{ tag.name }}
          </ElTag>
        </div>
      </article>

      <!-- 系列目录：同一分类下的全部文章，按发布时间倒序（这里恰好是 01→13 的阅读顺序） -->
      <nav v-if="series.length > 1" class="series-nav" aria-label="同系列文章">
        <h3 class="series-title">
          系列目录
          <span class="series-count">{{ series.length }} 篇</span>
        </h3>
        <ol class="series-list">
          <li v-for="item in series" :key="item.id">
            <RouterLink
              :to="`/article/${item.id}`"
              class="series-link"
              :class="{ 'series-link--current': item.id === article.id }"
              :aria-current="item.id === article.id ? 'page' : undefined"
            >
              <span class="series-name">{{ item.title }}</span>
              <span v-if="item.id === article.id" class="series-now">当前</span>
            </RouterLink>
          </li>
        </ol>
      </nav>

      <!-- 评论区 -->
      <section class="comments">
        <h3 class="comments-title">
          评论 <span class="count">{{ comments.length }}</span>
        </h3>

        <ElEmpty v-if="comments.length === 0" description="还没有评论，来说两句" :image-size="70" />

        <div v-else class="comment-list">
          <div v-for="item in comments" :key="item.id" class="comment">
            <div class="comment-head">
              <span class="avatar">{{ item.nickname.slice(0, 1) }}</span>
              <span class="nickname">{{ item.nickname }}</span>
              <span class="comment-time">{{ formatDate(item.created_at) }}</span>
            </div>
            <p class="comment-content">{{ item.content }}</p>

            <div v-if="item.replies && item.replies.length" class="replies">
              <div v-for="reply in item.replies" :key="reply.id" class="reply">
                <span class="nickname">{{ reply.nickname }}</span>
                <span class="comment-time">{{ formatDate(reply.created_at) }}</span>
                <p class="comment-content">{{ reply.content }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- 发表评论 -->
        <div class="comment-form card">
          <h4>发表评论</h4>
          <div class="form-row">
            <ElInput v-model="form.nickname" placeholder="昵称（必填）" maxlength="50" />
            <ElInput v-model="form.email" placeholder="邮箱（选填，不会公开）" maxlength="100" />
          </div>
          <ElInput
            v-model="form.content"
            type="textarea"
            :rows="4"
            maxlength="1000"
            show-word-limit
            placeholder="说点什么…（提交后需要管理员审核）"
          />
          <div class="form-actions">
            <ElButton type="primary" :loading="submitting" @click="submitComment">
              提交评论
            </ElButton>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.read-progress {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  z-index: 60;
  background: var(--amber);
  transform-origin: left;
  transform: scaleX(0);
  transition: transform 0.1s linear;
}

.article {
  max-width: var(--measure);
  margin: 0 auto;
}

.article-head {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.top-badge {
  flex-shrink: 0;
  margin-top: 0.6rem;
  font-family: var(--font-mono);
  font-size: 0.625rem;
  font-weight: 600;
  color: var(--bg);
  background: var(--amber);
  border-radius: 0.125rem;
  padding: 0.0625rem 0.3125rem;
}

.article-title {
  margin: 0;
  font-size: clamp(1.75rem, 3.6vw, 2.5rem);
  line-height: 1.24;
  letter-spacing: -0.02em;
}

.article-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--line);
  font-family: var(--font-sans);
  font-size: 0.75rem;
  color: var(--ink-3);
}

.author {
  color: var(--ink-2);
}

.meta-category {
  margin-left: auto;
  color: var(--amber);
}

.dot {
  opacity: 0.4;
}

.article-body {
  margin: 2.25rem 0;
}

/* 正文用衬线字：三千字的技术文需要可读性，不是展示性 */
.markdown-body {
  font-size: 1.0625rem;
  line-height: 1.9;
  color: #d3dae6;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  font-family: var(--font-mono);
  color: var(--ink);
  margin: 2.25em 0 0.7em;
  line-height: 1.32;
  letter-spacing: -0.01em;
}

.markdown-body :deep(h1) {
  font-size: 1.625rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--line);
}

.markdown-body :deep(h2) {
  font-size: 1.375rem;
}

.markdown-body :deep(h3) {
  font-size: 1.125rem;
}

.markdown-body :deep(p) {
  margin: 0 0 1.25em;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 0 0 1.25em;
  padding-left: 1.5em;
}

.markdown-body :deep(li) {
  margin: 0.35em 0;
}

.markdown-body :deep(li)::marker {
  color: var(--ink-3);
}

.markdown-body :deep(a) {
  color: var(--amber);
  text-decoration: underline;
  text-underline-offset: 0.2em;
  text-decoration-thickness: 1px;
  text-decoration-color: rgba(245, 185, 66, 0.35);
}

.markdown-body :deep(a:hover) {
  text-decoration-color: var(--amber);
}

.markdown-body :deep(blockquote) {
  margin: 1.5em 0;
  padding: 0.25em 0 0.25em 1.25em;
  border-left: 2px solid var(--amber);
  color: var(--ink-2);
  font-style: italic;
}

.markdown-body :deep(code) {
  font-family: var(--font-mono);
  font-size: 0.875em;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--line);
  border-radius: 0.25rem;
  padding: 0.1em 0.35em;
  color: #e9c98d;
}

.markdown-body :deep(pre) {
  margin: 1.5em 0;
  padding: 1.25rem 1.375rem;
  background: #0d131e;
  border: 1px solid var(--line-strong);
  border-radius: 0.625rem;
  overflow-x: auto;
  line-height: 1.7;
}

.markdown-body :deep(pre code) {
  background: none;
  border: none;
  padding: 0;
  font-size: 0.8125rem;
  line-height: 1.7;
  color: #ccd6e5;
}

.markdown-body :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1.5em 0;
  font-family: var(--font-sans);
  font-size: 0.875rem;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  border-bottom: 1px solid var(--line);
  padding: 0.625rem 0.75rem;
  text-align: left;
}

.markdown-body :deep(th) {
  color: var(--ink);
  font-weight: 600;
  background: rgba(255, 255, 255, 0.03);
}

.markdown-body :deep(img) {
  border-radius: 0.5rem;
}

/* ───────── 讲解图 ─────────
   文章里的图不是装饰，是「把结构讲清楚」的一部分。
   统一样式：整幅宽度、细边框、等宽字图注；点击可看原始大图。
   SVG 本身带深色底，边框让它在正文里有一块明确的「图版」区域。 */
.markdown-body :deep(.md-figure) {
  margin: 2.25em 0;
}

/* 桌面端让讲解图突破正文 42rem 的行宽，按 56rem 展示。
   图里的字号是按 960 宽画的，缩到 672 会小到看不清；
   正文仍然保持 42rem 行宽，只有图「破格」加宽，阅读节奏不变。 */
@media (min-width: 64rem) {
  .markdown-body :deep(.md-figure) {
    width: min(56rem, calc(100vw - 4rem));
    max-width: none;
    margin-left: 50%;
    transform: translateX(-50%);
  }
}

.markdown-body :deep(.md-figure-link) {
  display: block;
  text-decoration: none;
  border-radius: 0.625rem;
}

.markdown-body :deep(.md-figure img) {
  display: block;
  width: 100%;
  height: auto;
  border: 1px solid var(--line-strong);
  border-radius: 0.625rem;
  background: #0d131e;
  cursor: zoom-in;
  transition: border-color 0.15s ease;
}

.markdown-body :deep(.md-figure-link:hover img) {
  border-color: var(--amber);
}

.markdown-body :deep(.md-figure figcaption) {
  margin-top: 0.75rem;
  font-family: var(--font-mono);
  font-size: 0.75rem;
  line-height: 1.65;
  color: var(--ink-3);
  text-align: center;
  text-wrap: balance;
}

.markdown-body :deep(hr) {
  border: none;
  border-top: 1px solid var(--line);
  margin: 2.5em 0;
}

.article-tags {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  padding-top: 1.5rem;
  border-top: 1px solid var(--line);
}

/* ───────── 系列目录 ───────── */
.series-nav {
  max-width: var(--measure);
  margin: 3rem auto 0;
  padding-top: 1.5rem;
  border-top: 1px solid var(--line);
}

.series-title {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  margin: 0 0 1rem;
  font-size: 1.125rem;
}

.series-count {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  font-weight: 400;
  color: var(--ink-3);
}

.series-list {
  margin: 0;
  padding: 0;
  list-style: none;
  counter-reset: series;
}

.series-list li {
  counter-increment: series;
}

.series-link {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--line);
  color: var(--ink-2);
  text-decoration: none;
  font-size: 0.9375rem;
  transition: color 0.15s ease;
}

.series-link::before {
  content: counter(series, decimal-leading-zero);
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  color: var(--ink-3);
}

.series-link:hover {
  color: var(--ink);
}

.series-link--current {
  color: var(--amber);
}

.series-link--current::before {
  color: var(--amber);
}

.series-now {
  margin-left: auto;
  font-family: var(--font-mono);
  font-size: 0.625rem;
  color: var(--amber);
}

/* ───────── 评论 ───────── */
.comments {
  max-width: var(--measure);
  margin: 3rem auto 0;
}

.comments-title {
  margin: 0 0 1.25rem;
  font-size: 1.125rem;
}

.count {
  color: var(--ink-3);
  font-weight: 400;
}

.comment-list {
  display: flex;
  flex-direction: column;
}

.comment {
  padding: 1.25rem 0;
  border-bottom: 1px solid var(--line);
}

.comment-head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.avatar {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 50%;
  background: var(--amber-soft);
  color: var(--amber);
  display: grid;
  place-items: center;
  font-family: var(--font-mono);
  font-size: 0.75rem;
  font-weight: 600;
}

.nickname {
  font-family: var(--font-sans);
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--ink);
}

.comment-time {
  margin-left: auto;
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  color: var(--ink-3);
}

.comment-content {
  margin: 0.5rem 0 0;
  font-size: 0.9375rem;
  color: var(--ink-2);
  white-space: pre-wrap;
}

.replies {
  margin-top: 0.875rem;
  padding-left: 1rem;
  border-left: 1px solid var(--line);
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
}

/* 回复没有独立头部容器，用 flex 让昵称和时间分行对齐 */
.reply {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0 0.5rem;
}

.reply .comment-time {
  margin-left: 0;
}

.reply .comment-content {
  width: 100%;
  margin-top: 0.25rem;
}

.comment-form {
  margin-top: 2rem;
  padding: 1.5rem;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: var(--radius);
}

.comment-form h4 {
  margin: 0 0 1rem;
  font-family: var(--font-sans);
  font-size: 0.9375rem;
  color: var(--ink);
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 0.875rem;
}

@media (max-width: 48rem) {
  .markdown-body {
    font-size: 1rem;
    line-height: 1.85;
  }

  .markdown-body :deep(table) {
    display: block;
    overflow-x: auto;
    white-space: nowrap;
  }

  .meta-category {
    margin-left: 0;
  }

  .comment-form {
    padding: 1.125rem;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .comments {
    margin-top: 2rem;
  }
}
</style>
