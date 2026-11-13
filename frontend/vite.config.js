import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 开发期把 /api 和 /uploads 代理到后端，浏览器看起来是同源，
// 这样就不需要 CORS（虽然后端也配了，但代理更接近生产 Nginx 的形态）。
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8090',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://127.0.0.1:8090',
        changeOrigin: true,
      },
    },
  },
})
