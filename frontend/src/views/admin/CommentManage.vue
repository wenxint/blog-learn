<script setup>
// 评论审核：默认展示待审核，可切换到已通过
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { deleteComment, fetchAdminComments, setCommentStatus } from '../../api/admin'
import { formatDate } from '../../utils/format'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const status = ref(0) // 0 待审核 / 1 已通过
const pendingCount = ref(0)
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const data = await fetchAdminComments({
      page: page.value,
      page_size: pageSize.value,
      status: status.value,
    })
    list.value = data.list || []
    total.value = data.total || 0
    pendingCount.value = data.pending_count || 0
  } finally {
    loading.value = false
  }
}

function switchStatus(value) {
  status.value = value
  page.value = 1
  load()
}

async function approve(row) {
  await setCommentStatus(row.id, 1)
  ElMessage.success('已通过，前台可见')
  load()
}

async function reject(row) {
  await setCommentStatus(row.id, 0)
  ElMessage.success('已打回待审核')
  load()
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(
      `确定删除这条评论？它下面的回复会一起删除。`,
      '删除确认',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }

  await deleteComment(row.id)
  ElMessage.success('已删除')
  if (list.value.length === 1 && page.value > 1) page.value -= 1
  load()
}

onMounted(load)
</script>

<template>
  <div class="page">
    <div class="toolbar">
      <el-radio-group :model-value="status" @change="switchStatus">
        <el-radio-button :value="0">
          待审核
          <el-badge v-if="pendingCount > 0" :value="pendingCount" class="badge" />
        </el-radio-button>
        <el-radio-button :value="1">已通过</el-radio-button>
      </el-radio-group>

      <span class="muted">共 {{ total }} 条</span>
    </div>

    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="id" label="ID" width="64" />

      <el-table-column label="文章" width="90">
        <template #default="{ row }">
          <RouterLink :to="`/article/${row.article_id}`" class="link">
            #{{ row.article_id }}
          </RouterLink>
        </template>
      </el-table-column>

      <el-table-column label="昵称" width="120">
        <template #default="{ row }">
          <span class="nickname">{{ row.nickname }}</span>
          <span v-if="row.parent_id" class="reply-tag">回复#{{ row.parent_id }}</span>
        </template>
      </el-table-column>

      <el-table-column prop="content" label="内容" min-width="300" show-overflow-tooltip />

      <el-table-column prop="ip" label="IP" width="130" />

      <el-table-column label="时间" width="120">
        <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
      </el-table-column>

      <el-table-column label="操作" width="170" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status === 0" link type="success" @click="approve(row)">
            通过
          </el-button>
          <el-button v-else link type="warning" @click="reject(row)">打回</el-button>
          <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

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
  margin-bottom: 16px;
}

.badge {
  margin-left: 4px;
  transform: translateY(-2px);
}

.nickname {
  font-weight: 600;
  font-size: 13px;
}

.reply-tag {
  display: block;
  font-size: 11px;
  color: var(--text-tertiary);
}

.link {
  color: var(--primary);
  text-decoration: none;
  font-size: 13px;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
