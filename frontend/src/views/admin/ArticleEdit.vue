<script setup>
// 文章编辑页：新建与编辑共用（靠路由里有没有 id 区分）
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  createArticle,
  fetchAdminArticle,
  updateArticle,
  uploadImage,
} from '../../api/admin'
import { fetchCategories, fetchTags } from '../../api/article'
import { renderMarkdown } from '../../utils/markdown'
import 'highlight.js/styles/github.css'

const route = useRoute()
const router = useRouter()

const isEdit = computed(() => !!route.params.id)
const loading = ref(false)
const saving = ref(false)
const uploading = ref(false)
const showPreview = ref(false)
const categories = ref([])
const tags = ref([])

const form = reactive({
  title: '',
  summary: '',
  content: '',
  cover: '',
  status: 0,
  is_top: 0,
  category_id: undefined,
  tag_ids: [],
})

const previewHtml = computed(() => renderMarkdown(form.content))

async function loadOptions() {
  const [catData, tagData] = await Promise.all([fetchCategories(), fetchTags()])
  categories.value = catData.list || []
  tags.value = tagData.list || []
}

async function loadArticle() {
  if (!isEdit.value) return
  loading.value = true
  try {
    const a = await fetchAdminArticle(route.params.id)
    form.title = a.title
    form.summary = a.summary
    form.content = a.content
    form.cover = a.cover
    form.status = a.status
    form.is_top = a.is_top
    form.category_id = a.category_id || undefined
    form.tag_ids = (a.tags || []).map((t) => t.id)
  } finally {
    loading.value = false
  }
}

// 封面上传：走我们自己的 /admin/upload，不用 Element 的默认上传
async function handleCoverChange(file) {
  uploading.value = true
  try {
    const data = await uploadImage(file.raw)
    form.cover = data.url
    ElMessage.success('封面上传成功')
  } finally {
    uploading.value = false
  }
  return false // 阻止 el-upload 自动提交到自己的 action
}

async function handleSave() {
  if (!form.title.trim()) {
    ElMessage.warning('标题不能为空')
    return
  }
  if (!form.content.trim()) {
    ElMessage.warning('正文不能为空')
    return
  }

  saving.value = true
  try {
    const payload = {
      title: form.title.trim(),
      summary: form.summary.trim(),
      content: form.content,
      cover: form.cover,
      status: form.status,
      is_top: form.is_top,
      category_id: form.category_id || 0,
      tag_ids: form.tag_ids,
    }

    if (isEdit.value) {
      await updateArticle(route.params.id, payload)
      ElMessage.success('已保存')
    } else {
      await createArticle(payload)
      ElMessage.success('已创建')
    }
    router.push('/admin/articles')
  } catch {
    // 拦截器已提示
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadOptions()
  await loadArticle()
})
</script>

<template>
  <div v-loading="loading" class="edit-page">
    <div class="grid">
      <!-- 左：编辑区 -->
      <div class="editor">
        <el-input v-model="form.title" placeholder="文章标题" size="large" class="title-input" />

        <el-input
          v-model="form.summary"
          placeholder="摘要（列表页展示，可不填）"
          maxlength="500"
          show-word-limit
        />

        <div class="toolbar">
          <el-button size="small" @click="showPreview = !showPreview">
            {{ showPreview ? '编辑' : '预览' }}
          </el-button>
          <span class="muted">支持 Markdown：标题、列表、代码块、表格</span>
        </div>

        <el-input
          v-show="!showPreview"
          v-model="form.content"
          type="textarea"
          :rows="22"
          placeholder="# 标题&#10;&#10;正文用 Markdown 写…"
          class="content-input"
        />

        <!-- 预览时同样要经过 DOMPurify 净化 -->
        <div v-show="showPreview" class="markdown-body card preview" v-html="previewHtml"></div>
      </div>

      <!-- 右：属性面板 -->
      <aside class="side">
        <div class="side-block">
          <h4>发布设置</h4>
          <el-form label-position="top">
            <el-form-item label="状态">
              <el-radio-group v-model="form.status">
                <el-radio :value="0">草稿</el-radio>
                <el-radio :value="1">发布</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="置顶">
              <el-switch v-model="form.is_top" :active-value="1" :inactive-value="0" />
            </el-form-item>

            <el-form-item label="分类">
              <el-select v-model="form.category_id" placeholder="选择分类" clearable style="width: 100%">
                <el-option
                  v-for="c in categories"
                  :key="c.id"
                  :label="c.name"
                  :value="c.id"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="标签">
              <el-select
                v-model="form.tag_ids"
                multiple
                placeholder="选择标签"
                style="width: 100%"
              >
                <el-option v-for="t in tags" :key="t.id" :label="t.name" :value="t.id" />
              </el-select>
            </el-form-item>
          </el-form>
        </div>

        <div class="side-block">
          <h4>封面图</h4>
          <div class="cover-box">
            <img v-if="form.cover" :src="form.cover" alt="封面" class="cover-img" />
            <div v-else class="cover-empty muted">暂无封面</div>
          </div>

          <el-upload
            :show-file-list="false"
            :before-upload="handleCoverChange"
            accept="image/png,image/jpeg,image/gif,image/webp"
          >
            <el-button size="small" :loading="uploading">上传封面</el-button>
          </el-upload>

          <el-input
            v-model="form.cover"
            placeholder="或直接填图片地址"
            size="small"
            class="cover-url"
          />
        </div>

        <div class="actions">
          <el-button @click="router.push('/admin/articles')">取消</el-button>
          <el-button type="primary" :loading="saving" @click="handleSave">
            {{ isEdit ? '保存修改' : '创建文章' }}
          </el-button>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 18px;
  align-items: start;
}

.editor {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.title-input :deep(input) {
  font-size: 18px;
  font-weight: 600;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.content-input :deep(textarea) {
  font-family: 'SF Mono', Menlo, Consolas, monospace;
  font-size: 13.5px;
  line-height: 1.7;
}

.preview {
  min-height: 200px;
  padding: 4px 6px;
  border: 1px dashed var(--border);
  border-radius: 8px;
  font-size: 15px;
  line-height: 1.8;
}

.side {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.side-block {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 16px;
}

.side-block h4 {
  margin: 0 0 12px;
  font-size: 14px;
}

.cover-box {
  margin-bottom: 10px;
}

.cover-img {
  width: 100%;
  border-radius: 8px;
  display: block;
}

.cover-empty {
  height: 90px;
  border: 1px dashed var(--border);
  border-radius: 8px;
  display: grid;
  place-items: center;
  font-size: 13px;
}

.cover-url {
  margin-top: 8px;
}

.actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

/* Markdown 预览样式（与详情页一致的简化版） */
.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  margin: 1.2em 0 0.5em;
}

.markdown-body :deep(pre) {
  background: #f8f9fc;
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 14px;
  overflow-x: auto;
}

.markdown-body :deep(code) {
  background: #f2f4f8;
  border-radius: 4px;
  padding: 2px 5px;
  font-size: 0.9em;
}

.markdown-body :deep(pre code) {
  background: transparent;
  padding: 0;
}

.markdown-body :deep(blockquote) {
  margin: 1em 0;
  padding: 8px 14px;
  border-left: 3px solid #c9d3ea;
  background: #f8f9fc;
  color: var(--text-secondary);
}

.markdown-body :deep(table) {
  width: 100%;
  border-collapse: collapse;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid var(--border);
  padding: 7px 10px;
}
</style>
