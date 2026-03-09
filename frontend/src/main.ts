import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import './styles/element-plus-override.css'
import App from './App.vue'
import router from './router'
import { i18n } from './i18n'
import { useTheme } from './composables/useTheme'
import { loadElementLocale } from './composables/useLocale'
import { getInitialLocale } from './i18n'

const app = createApp(App)

app.use(createPinia())
app.use(i18n)
app.use(router)
app.use(ElementPlus, {
  size: 'default',
})

// 初始化主题
useTheme()

// 初始化 Element Plus 语言包
loadElementLocale(getInitialLocale()).then(() => {
  app.mount('#app')
})
