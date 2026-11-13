<script setup>
// 归档页：按年份分组的时间轴，点月份跳到首页并带上筛选条件
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchArchives } from '../api/article'

const router = useRouter()
const archives = ref([])
const loading = ref(true)

// 把 [{month:'2026-11',count:4}] 变成 [{year:'2026', items:[...]}]
const grouped = computed(() => {
  const map = new Map()
  for (const item of archives.value) {
    const year = item.month.slice(0, 4)
    if (!map.has(year)) map.set(year, [])
    map.get(year).push(item)
  }
  return [...map.entries()].map(([year, items]) => ({ year, items }))
})

const totalCount = computed(() => archives.value.reduce((sum, a) => sum + a.count, 0))
const archiveMonths = computed(() => archives.value.length)

function goMonth(month) {
  router.push({ path: '/', query: { month } })
}

onMounted(async () => {
  try {
    const data = await fetchArchives()
    archives.value = data.list || []
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="container">
    <header class="page-head">
      <h1>归档</h1>
      <p class="muted">共 {{ archiveMonths }} 个月 · {{ totalCount }} 篇文章</p>
    </header>

    <ElSkeleton v-if="loading" :rows="3" animated />

    <ElEmpty v-else-if="archives.length === 0" description="还没有发布文章" :image-size="90" />

    <div v-else class="timeline">
      <section v-for="group in grouped" :key="group.year" class="year-block">
        <h2 class="year">{{ group.year }}</h2>
        <ul class="months">
          <li v-for="item in group.items" :key="item.month">
            <button class="month-row" @click="goMonth(item.month)">
              <span class="month-label">{{ item.month.slice(5) }} 月</span>
              <span class="month-dots"></span>
              <span class="month-count">{{ item.count }} 篇</span>
            </button>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>


<style scoped>
.page-head {
  margin-bottom: 26px;
}

.page-head h1 {
  margin: 0 0 6px;
  font-size: 26px;
}

.timeline {
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.year-block {
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 22px 26px;
}

.year {
  margin: 0 0 14px;
  font-size: 20px;
  letter-spacing: 1px;
  color: var(--primary);
}

.months {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.month-row {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 6px;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 8px;
  font-size: 14px;
  color: var(--text-secondary);
  transition: background 0.15s;
}

.month-row:hover {
  background: var(--primary-soft);
  color: var(--primary);
}

.month-label {
  font-weight: 600;
  color: var(--text);
}

.month-dots {
  flex: 1;
  border-bottom: 1px dashed var(--border);
  height: 1px;
}

.month-count {
  font-size: 13px;
  color: var(--text-tertiary);
}
</style>
