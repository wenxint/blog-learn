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

.timeline {
  display: flex;
  flex-direction: column;
  gap: 3rem;
}

.year {
  margin: 0 0 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--line);
  font-size: 1.5rem;
  color: var(--amber);
  letter-spacing: 0.02em;
}

.months {
  list-style: none;
  margin: 0;
  padding: 0;
}

/* 月份行：整行可点，hover 时轻微右移，代替卡片 */
.month-row {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 0;
  border: none;
  border-bottom: 1px solid var(--line);
  background: none;
  cursor: pointer;
  text-align: left;
  font-family: var(--font-sans);
  font-size: 0.9375rem;
  color: var(--ink-2);
  transition: color 0.18s ease, padding-left 0.2s ease;
}

.month-row:hover {
  color: var(--amber);
  padding-left: 0.5rem;
}

.month-label {
  font-family: var(--font-mono);
  color: var(--ink);
}

.month-dots {
  flex: 1;
  height: 1px;
  border-bottom: 1px dashed var(--line);
}

.month-count {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--ink-3);
}
</style>
