<script setup>
import { useRouter } from 'vue-router'
import { ElButton } from 'element-plus'

const router = useRouter()
</script>

<template>
  <div class="container notfound">
    <h1>404</h1>
    <p class="muted">这个页面不存在，可能地址写错了。</p>
    <ElButton type="primary" @click="router.push('/')">回到首页</ElButton>
  </div>
</template>

<style scoped>
.notfound {
  text-align: center;
  padding: 80px 0;
}

h1 {
  font-size: 64px;
  margin: 0;
  color: var(--text-tertiary);
  letter-spacing: -2px;
}
</style>
