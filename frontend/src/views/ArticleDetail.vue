<script setup>
// 文章详情页：正文 + 评论
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElTag, ElInput, ElButton, ElMessage, ElSkeleton, ElEmpty } from 'element-plus'
import { createComment, fetchArticleDetail, fetchComments } from '../api/article'
import { formatDate } from '../utils/format'

const route = useRoute()
const router = useRouter()

const article = ref(null)
const comments = ref([])
const loading = ref(true)
const submitting = ref(false)
const form = ref({ nickname: '', email: '', content: '' })

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
            第 18 课会换成 marked + highlight.js 渲染真正的 Markdown。
            现在先用 pre-wrap 保留换行，让内容可读即可。
          -->
          <pre class="article-content">{{ article.content }}</pre>
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

.article-content {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: inherit;
  font-size: 15.5px;
  line-height: 1.85;
  color: var(--text);
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
