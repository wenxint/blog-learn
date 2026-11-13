<script setup>
// 后台登录页：没有布局（不套侧边栏），单独一屏
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../../stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

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
        <span class="mark">B</span>
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
  background: linear-gradient(160deg, #eef2ff 0%, #f7f8fb 40%, #f5f6fa 100%);
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 380px;
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 34px 32px 26px;
  box-shadow: 0 12px 40px rgba(23, 43, 99, 0.08);
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.mark {
  width: 32px;
  height: 32px;
  border-radius: 9px;
  background: linear-gradient(135deg, #4f7cff, #7b5cff);
  color: #fff;
  display: grid;
  place-items: center;
  font-weight: 700;
}

h1 {
  margin: 0;
  font-size: 20px;
}

.submit {
  width: 100%;
  margin-top: 6px;
}

.hint {
  margin-top: 16px;
  text-align: center;
}

.back {
  display: block;
  margin-top: 14px;
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary);
  text-decoration: none;
}

.back:hover {
  color: var(--primary);
}
</style>
