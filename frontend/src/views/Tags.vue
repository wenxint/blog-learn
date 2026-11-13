<script setup>
// 标签云页：点标签跳到首页并带上 tag_id 筛选
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchTags } from '../api/article'

const router = useRouter()
const tags = ref([])
const loading = ref(true)

// 按文章数从多到少排，最常用的标签先看到
const sorted = computed(() => [...tags.value].sort((a, b) => b.article_count - a.article_count))

// 标签字号随文章数变化：用得越多越大，形成"标签云"的视觉效果
function fontOf(tag) {
  if (tag.article_count >= 3) return 20
  if (tag.article_count === 2) return 17
  return 15
}

function goTag(tag) {
  router.push({ path: '/', query: { tag_id: tag.id, tag_name: tag.name } })
}

onMounted(async () => {
  try {
    const data = await fetchTags()
    tags.value = data.list || []
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="container">
    <header class="page-head">
      <h1>标签</h1>
      <p class="muted">共 {{ sorted.length }} 个标签</p>
    </header>

    <ElSkeleton v-if="loading" :rows="2" animated />
    <ElEmpty v-else-if="sorted.length === 0" description="还没有标签" :image-size="90" />

    <div v-else class="tag-cloud">
      <button
        v-for="tag in sorted"
        :key="tag.id"
        class="tag-chip"
        :style="{ fontSize: fontOf(tag) + 'px' }"
        @click="goTag(tag)"
      >
        {{ tag.name }}
        <span class="count">{{ tag.article_count }}</span>
      </button>
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

.tag-cloud {
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 28px;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.tag-chip {
  border: 1px solid var(--border);
  background: #fff;
  color: var(--text-secondary);
  border-radius: 999px;
  padding: 6px 16px;
  cursor: pointer;
  transition: all 0.15s;
  line-height: 1.4;
}

.tag-chip:hover {
  background: var(--primary-soft);
  border-color: #b9c8ff;
  color: var(--primary);
  transform: translateY(-1px);
}

.count {
  font-size: 12px;
  opacity: 0.55;
  margin-left: 2px;
}
</style>
