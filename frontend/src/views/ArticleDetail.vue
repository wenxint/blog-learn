<script setup>
// 文章详情页：正文 + 评论
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElTag, ElInput, ElButton, ElMessage, ElSkeleton, ElEmpty } from 'element-plus'
import { createComment, fetchArticleDetail, fetchComments } from '../api/article'
import { formatDate } from '../utils/format'
import { renderMarkdown } from '../utils/markdown'
// 代码高亮主题（github 风格浅色）
import 'highlight.js/styles/github.css'

const route = useRoute()
const router = useRouter()

const article = ref(null)
const comments = ref([])
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

onMounted(load)
</script>

<template>
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
          <span class="dot">·</span>
          <span>{{ formatDate(article.published_at) }}</span>
          <span class="dot">·</span>
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
.article {
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 32px 36px;
}

.article-head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.top-badge {
  font-size: 11px;
  color: #fff;
  background: #ff7a45;
  border-radius: 4px;
  padding: 2px 7px;
  flex-shrink: 0;
}

.article-title {
  margin: 0;
  font-size: 28px;
  line-height: 1.35;
  letter-spacing: -0.4px;
}

.article-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 14px;
  font-size: 13px;
  color: var(--text-tertiary);
  flex-wrap: wrap;
}

.author {
  color: var(--text-secondary);
  font-weight: 600;
}

.meta-category {
  color: var(--primary);
  background: var(--primary-soft);
  border-radius: 4px;
  padding: 2px 8px;
  margin-left: auto;
}

.article-body {
  margin: 26px 0;
}

.markdown-body {
  font-size: 15.5px;
  line-height: 1.85;
  color: var(--text);
  word-break: break-word;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  margin: 1.6em 0 0.6em;
  line-height: 1.4;
  font-weight: 600;
}

.markdown-body :deep(h1) {
  font-size: 24px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border);
}

.markdown-body :deep(h2) {
  font-size: 20px;
}

.markdown-body :deep(h3) {
  font-size: 17px;
}

.markdown-body :deep(p) {
  margin: 0 0 1em;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 1.4em;
  margin: 0 0 1em;
}

.markdown-body :deep(li) {
  margin: 0.25em 0;
}

.markdown-body :deep(a) {
  color: var(--primary);
  text-decoration: none;
}

.markdown-body :deep(a:hover) {
  text-decoration: underline;
}

.markdown-body :deep(blockquote) {
  margin: 1em 0;
  padding: 10px 16px;
  border-left: 3px solid #c9d3ea;
  background: #f8f9fc;
  color: var(--text-secondary);
  border-radius: 0 8px 8px 0;
}

.markdown-body :deep(code) {
  background: #f2f4f8;
  border-radius: 4px;
  padding: 2px 6px;
  font-size: 0.9em;
  font-family: 'SF Mono', Menlo, Consolas, monospace;
}

.markdown-body :deep(pre) {
  margin: 1.2em 0;
  padding: 16px 18px;
  background: #f8f9fc;
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow-x: auto;
}

.markdown-body :deep(pre code) {
  background: transparent;
  padding: 0;
  font-size: 13.5px;
  line-height: 1.65;
}

.markdown-body :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1em 0;
  font-size: 14px;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid var(--border);
  padding: 8px 12px;
  text-align: left;
}

.markdown-body :deep(th) {
  background: #f8f9fc;
  font-weight: 600;
}

.markdown-body :deep(img) {
  max-width: 100%;
  border-radius: 8px;
}

.markdown-body :deep(hr) {
  border: none;
  border-top: 1px solid var(--border);
  margin: 2em 0;
}

.article-tags {
  display: flex;
  gap: 8px;
  padding-top: 18px;
  border-top: 1px solid var(--border);
}

.comments {
  margin-top: 30px;
}

.comments-title {
  margin: 0 0 16px;
  font-size: 18px;
}

.count {
  color: var(--text-tertiary);
  font-weight: 400;
}

.comment-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.comment {
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 16px 18px;
}

.comment-head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: var(--primary-soft);
  color: var(--primary);
  display: grid;
  place-items: center;
  font-size: 13px;
  font-weight: 600;
}

.nickname {
  font-weight: 600;
  font-size: 14px;
}

.comment-time {
  color: var(--text-tertiary);
  font-size: 12px;
}

.comment-content {
  margin: 8px 0 0;
  color: var(--text-secondary);
  white-space: pre-wrap;
}

.replies {
  margin-top: 12px;
  padding-left: 14px;
  border-left: 2px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.reply .comment-content {
  margin-top: 4px;
}

.comment-form {
  margin-top: 22px;
  padding: 20px;
}

.comment-form h4 {
  margin: 0 0 14px;
  font-size: 15px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 10px;
}

.form-actions {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

.dot {
  opacity: 0.5;
}
</style>
