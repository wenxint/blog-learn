import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { login as loginApi, fetchProfile } from '../api/auth'

// 登录态 store。
//
// token 存在 localStorage 而不是只在内存里：
// 刷新页面后还能保持登录。代价是能被 XSS 偷走——
// 所以正文渲染必须过 DOMPurify（见 utils/markdown.js）。
export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function login(username, password) {
    const data = await loginApi(username, password)
    token.value = data.token
    user.value = data.user
    localStorage.setItem('token', data.token)
    localStorage.setItem('user', JSON.stringify(data.user))
    return data
  }

  // 刷新页面后那一份 user 可能是旧的（改了昵称），这里静默同步一次
  async function loadProfile() {
    if (!token.value) return
    try {
      const u = await fetchProfile()
      user.value = u
      localStorage.setItem('user', JSON.stringify(u))
    } catch (err) {
      // 只有「明确未授权」才清登录态。
      // 网络抖动、后端重启、超时这类临时错误不能把用户踢出去——
      // 否则后端一重启，所有正在操作的人都被登出，体验很差。
      if (err?.response?.status === 401) {
        logout()
      }
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return { token, user, isLoggedIn, isAdmin, login, logout, loadProfile }
})
