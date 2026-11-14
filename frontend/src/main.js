import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
// Element Plus 的暗色变量。全站是深色主题，后台的表格/表单/弹窗
// 靠这套变量自动跟随，否则会出现白底控件突兀地压在深色页面上。
// 注意顺序：必须先引入它，再引入 main.css，
// 否则 main.css 里对 --el-color-primary 的覆盖会被它盖掉。
import 'element-plus/theme-chalk/dark/css-vars.css'
import './styles/main.css'

// 组件和样式由 vite.config.js 里的两个插件按需引入，
// 只为真正用到的组件加载 JS 与 CSS。
document.documentElement.classList.add('dark')

const app = createApp(App)

// Pinia 必须在 router 之前注册：路由守卫里可能会用到 store
app.use(createPinia())
app.use(router)

app.mount('#app')
