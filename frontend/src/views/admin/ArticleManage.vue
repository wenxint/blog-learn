<script setup>
// 文章管理：列表 + 筛选 + 删除 + 跳转编辑
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchAdminArticles, deleteArticle } from '../../api/admin'
import { formatDate } from '../../utils/format'

const router = useRouter()

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const status = ref('') // '' 全部 / 0 草稿 / 1 已发布
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const data = await fetchAdminArticles({
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value || undefined,
      status: status.value === '' ? undefined : status.value,
    })
    list.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  load()
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(
      `确定删除《${row.title}》？删除后前台将不再显示（软删除，数据库里还留着）。`,
      '删除确认',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }

  try {
    await deleteArticle(row.id)
    ElMessage.success('已删除')
    // 删掉当前页最后一条时，回退一页，避免停在空页
    if (list.value.length === 1 && page.value > 1) {
      page.value -= 1
    }
    load()
  } catch {
    // 拦截器已提示
  }
}

onMounted(load)
</script>

<template>
  <div class="page">
    <div class="toolbar">
      <div class="filters">
        <el-input
          v-model="keyword"
          placeholder="搜索标题"
          clearable
          style="width: 200px"
          @keyup.enter="search"
          @clear="search"
        />
        <el-select v-model="status" placeholder="全部状态" style="width: 130px" @change="search">
          <el-option label="全部状态" value="" />
          <el-option label="已发布" :value="1" />
          <el-option label="草稿" :value="0" />
        </el-select>
        <el-button @click="search">查询</el-button>
      </div>

      <el-button type="primary" @click="router.push('/admin/articles/new')">+ 新建文章</el-button>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="list" border stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="64" />

        <el-table-column label="标题" min-width="260">
          <template #default="{ row }">
            <div class="title-cell">
              <span v-if="row.is_top === 1" class="top">置顶</span>
              <span class="title-text">{{ row.title }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="96">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small" effect="plain">
              {{ row.status === 1 ? '已发布' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="分类" width="110">
          <template #default="{ row }">
            {{ row.category?.name || '未分类' }}
          </template>
        </el-table-column>

        <el-table-column prop="view_count" label="阅读" width="80" />

        <el-table-column label="发布时间" width="120">
          <template #default="{ row }">{{ formatDate(row.published_at) }}</template>
        </el-table-column>

        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="router.push(`/admin/articles/${row.id}/edit`)">
              编辑
            </el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 窄屏用卡片代替表格：表格列太多，压缩后无法阅读 -->
    <ul v-loading="loading" class="mobile-list">
      <li v-for="row in list" :key="row.id" class="m-card">
        <div class="m-head">
          <span class="m-title">{{ row.title }}</span>
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small" effect="plain">
            {{ row.status === 1 ? '已发布' : '草稿' }}
          </el-tag>
        </div>
        <div class="m-meta">
          <span>{{ row.category?.name || '未分类' }}</span>
          <span>{{ row.view_count }} 阅读</span>
          <span>{{ formatDate(row.published_at) }}</span>
        </div>
        <div class="m-actions">
          <el-button link type="primary" @click="router.push(`/admin/articles/${row.id}/edit`)">
            编辑
          </el-button>
          <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
        </div>
      </li>
    </ul>

    <div class="pager">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        background
        @current-change="load"
      />
    </div>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.filters {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

/* 表格里 link 按钮的主色在暗色背景上偏灰，显式提亮，保证「编辑」看得清 */
:deep(.el-button.is-link) {
  color: var(--amber);
}

:deep(.el-button.is-link.el-button--danger) {
  color: #f07575;
}

.mobile-list {
  display: none;
}

.title-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.top {
  font-size: 0.625rem;
  font-weight: 600;
  color: var(--bg);
  background: var(--amber);
  border-radius: 0.1875rem;
  padding: 0 0.3125rem;
  flex-shrink: 0;
}

.title-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

/* ───────── 窄屏 ───────── */
@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }

  .filters {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  /* 覆盖模板里的内联 width: 200px / 130px */
  .filters :deep(.el-input),
  .filters :deep(.el-select) {
    width: 100% !important;
  }

  .filters :deep(.el-button) {
    width: 100%;
  }

  /* 「新建文章」按钮单独占满一行 */
  .toolbar > :deep(.el-button) {
    width: 100%;
  }

  .pager {
    justify-content: center;
    flex-wrap: wrap;
  }

  .pager :deep(.el-pagination) {
    flex-wrap: wrap;
    justify-content: center;
    row-gap: 0.5rem;
  }

  /* 表格横向滚动，避免列被压缩到不可读 */
  /* 手机端直接换成卡片列表，不再往 344px 里硬塞 7 列表格 */
  .table-wrap {
    display: none;
  }

  .mobile-list {
    display: flex;
    flex-direction: column;
    gap: 0.625rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .m-card {
    padding: 0.75rem;
    border: 1px solid var(--line);
    border-radius: 0.5rem;
    background: #0b0f16;
  }

  .m-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.625rem;
  }

  .m-title {
    font-size: 0.875rem;
    line-height: 1.6;
    color: var(--ink);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .m-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0.375rem 0.875rem;
    margin-top: 0.5rem;
    font-size: 0.75rem;
    color: var(--ink-3);
  }

  .m-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.25rem;
    margin-top: 0.5rem;
    padding-top: 0.5rem;
    border-top: 1px solid var(--line);
  }
}
</style>
