import axios from 'axios'
import { ElMessage } from 'element-plus'

// 统一的后端地址。
// 开发期 baseURL 是 /api/v1，由 vite.config.js 的 proxy 转发到 8090 端口；
// 生产期 Nginx 做同样的转发，前端代码一行都不用改。
const request = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// ---------- 请求拦截器 ----------
// 每次请求前执行：把 token 自动塞进 Authorization 头。
// 这样每个 api 函数都不用自己记着带 token。
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

// ---------- 响应拦截器 ----------
// 后端统一返回 { code, message, data }：
//   成功 → 这里直接返回 data，页面里拿到的就是干净的业务数据
//   失败 → 统一弹提示，页面只处理成功分支
request.interceptors.response.use(
  (response) => {
    const body = response.data

    // HTTP 200 但业务码不为 0，也算失败
    if (body && body.code !== 0) {
      ElMessage.error(body.message || '请求失败')
      return Promise.reject(new Error(body.message || '请求失败'))
    }
    return body.data
  },
  (error) => {
    const status = error.response?.status
    const body = error.response?.data
    const message = body?.message || error.message || '网络错误'

    if (status === 401) {
      // token 失效：清掉本地登录态，提示重新登录
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      ElMessage.error('登录已过期，请重新登录')
    } else if (status === 403) {
      ElMessage.error(message || '没有权限')
    } else {
      ElMessage.error(message)
    }

    return Promise.reject(error)
  },
)

export default request
