import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import './styles/main.css'

// 注意这里不再有：
//   import ElementPlus from 'element-plus'
//   import 'element-plus/dist/index.css'
//   app.use(ElementPlus)
// 组件和样式改为由 vite.config.js 里的两个插件按需引入，
// 只为真正用到的组件加载 JS 与 CSS。

const app = createApp(App)

// Pinia 必须在 router 之前注册：路由守卫里可能会用到 store
app.use(createPinia())
app.use(router)

app.mount('#app')
