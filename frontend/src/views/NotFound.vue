<script setup>
import { useRouter } from 'vue-router'
const router = useRouter()
</script>

<template>
  <div class="container notfound">
    <p class="code">404</p>
    <h1>页面不存在</h1>
    <p class="desc">地址可能写错了，或者这篇文章已经被删掉。</p>
    <ElButton type="primary" @click="router.push('/')">回到首页</ElButton>
  </div>
</template>

<style scoped>
.notfound {
  max-width: 26rem;
  margin: 0 auto;
  padding: 5rem 0 7rem;
  text-align: center;
}

.code {
  margin: 0 0 1rem;
  font-family: var(--font-mono);
  font-size: clamp(4rem, 15vw, 7rem);
  line-height: 1;
  letter-spacing: -0.05em;
  color: var(--ink-3);
}

h1 {
  margin: 0 0 0.75rem;
  font-size: 1.25rem;
  color: var(--ink);
}

.desc {
  margin: 0 0 2rem;
  font-size: 0.9375rem;
  color: var(--ink-2);
}
</style>
