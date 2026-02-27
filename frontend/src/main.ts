import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import './styles/element-plus-override.css'
import App from './App.vue'
import router from './router'
import { useTheme } from './composables/useTheme'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus, {
  // 配置 Element Plus 主题以匹配设计规范
  size: 'default',
})

// 初始化主题
useTheme()

app.mount('#app')
