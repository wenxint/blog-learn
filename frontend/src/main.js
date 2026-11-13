import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'

import App from './App.vue'
import router from './router'
import './styles/main.css'

const app = createApp(App)

// Pinia 必须在 router 之前注册：路由守卫里可能会用到 store
app.use(createPinia())
app.use(router)

// 组件库用中文语言包（分页器会显示「共 x 条」而不是英文）
app.use(ElementPlus, { locale: zhCn })

app.mount('#app')
