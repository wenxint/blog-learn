import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

// 开发期把 /api 和 /uploads 代理到后端，浏览器看起来是同源，
// 这样就不需要 CORS（虽然后端也配了，但代理更接近生产 Nginx 的形态）。
export default defineConfig({
  plugins: [
    vue(),

    // Element Plus 按需引入（两个插件分工不同）：
    //   AutoImport  -> 处理 JS 里用到的 ElMessage / ElMessageBox 这类「函数」
    //   Components  -> 处理 template 里写的 <el-table> 这类「组件」
    // 两者都会顺便把对应的 CSS 按需引入。
    //
    // 不配这个的话，main.js 里 app.use(ElementPlus) 会把 60+ 个组件和
    // 360KB 的完整 CSS 全打进包里（主包 830KB）。
    AutoImport({
      resolvers: [ElementPlusResolver()],
      dts: false, // 纯 JS 项目，不生成类型声明文件
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: false,
    }),
  ],

  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },

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

  build: {
    // 关掉 500KB 的警告阈值噪音（按需引入后主包已显著变小）
    chunkSizeWarningLimit: 600,
  },
})
