<script setup>
// 后台登录页：没有布局（不套侧边栏），单独一屏
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../../stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 生产构建里不显示默认账号提示，避免把后台入口信息暴露出去
const isDev = import.meta.env.DEV

const form = ref({ username: 'admin', password: '' })
const loading = ref(false)

async function handleSubmit() {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('请填写用户名和密码')
    return
  }

  loading.value = true
  try {
    const data = await userStore.login(form.value.username, form.value.password)
    ElMessage.success(`欢迎回来，${data.user.nickname || data.user.username}`)

    // 登录成功回到原本想去的页面（守卫里记下来的），否则进文章管理
    const redirect = route.query.redirect || '/admin/articles'
    router.replace(redirect)
  } catch {
    // 错误提示已经由 axios 响应拦截器统一弹了，这里不用重复处理
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="brand">
        <span class="mark">~/</span>
        <h1>博客后台</h1>
      </div>
      <p class="muted">登录后可以管理文章、审核评论</p>

      <el-form :model="form" label-position="top" @submit.prevent="handleSubmit">
        <el-form-item label="用户名">
          <el-input
            v-model="form.username"
            placeholder="admin"
            size="large"
            @keyup.enter="handleSubmit"
          />
        </el-form-item>
        <el-form-item label="密码">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            size="large"
            show-password
            @keyup.enter="handleSubmit"
          />
        </el-form-item>

        <el-button
          type="primary"
          size="large"
          class="submit"
          :loading="loading"
          @click="handleSubmit"
        >
          登录
        </el-button>
      </el-form>

      <div class="hint">
        <span class="muted">本地开发账号：admin / 123456</span>
      </div>
      <RouterLink to="/" class="back">← 回到博客首页</RouterLink>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  background: var(--bg);
}

.login-card {
  width: 100%;
  max-width: 24rem;
  padding: 2.25rem 2rem 1.75rem;
  background: var(--panel);
  border: 1px solid var(--line-strong);
  border-radius: 0.75rem;
  box-shadow: 0 2rem 4rem rgba(0, 0, 0, 0.45);
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  margin-bottom: 0.5rem;
  font-family: var(--font-mono);
}

.mark {
  width: 2rem;
  height: 2rem;
  border-radius: 0.5rem;
  display: grid;
  place-items: center;
  background: var(--amber);
  color: var(--bg);
  font-family: var(--font-mono);
  font-weight: 700;
}

h1 {
  margin: 0;
  font-size: 1.25rem;
}

.submit {
  width: 100%;
  margin-top: 0.375rem;
}

.hint {
  margin-top: 1rem;
  text-align: center;
}

.back {
  display: block;
  margin-top: 0.875rem;
  text-align: center;
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--ink-3);
}

.back:hover {
  color: var(--amber);
}

@media (max-width: 30rem) {
  .login-card {
    padding: 1.75rem 1.25rem 1.25rem;
  }
}
</style>
