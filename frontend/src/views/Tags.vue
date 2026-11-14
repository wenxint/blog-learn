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

.tag-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 0.75rem;
  padding: 1.5rem 0;
  border-top: 1px solid var(--line);
}

.tag-chip {
  border: 1px solid var(--line);
  background: none;
  color: var(--ink-2);
  border-radius: 999px;
  padding: 0.375rem 0.875rem;
  cursor: pointer;
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  transition: border-color 0.18s ease, color 0.18s ease;
}

.tag-chip:hover {
  border-color: var(--amber);
  color: var(--amber);
}

.count {
  font-size: 0.6875rem;
  opacity: 0.5;
  margin-left: 0.25rem;
}
</style>
