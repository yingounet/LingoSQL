/**
 * 主题切换组合式函数
 */

import { ref, onMounted } from 'vue'
import { cssVariables, darkCssVariables } from '@/utils/theme'

export type Theme = 'light' | 'dark'

const theme = ref<Theme>('light')

export function useTheme() {
  // 从本地存储加载主题
  function loadTheme() {
    const saved = localStorage.getItem('theme') as Theme | null
    if (saved && (saved === 'light' || saved === 'dark')) {
      theme.value = saved
    } else {
      // 检测系统主题
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      theme.value = prefersDark ? 'dark' : 'light'
    }
    applyTheme()
  }

  // 应用主题
  function applyTheme() {
    document.documentElement.setAttribute('data-theme', theme.value)
    const variables = theme.value === 'dark' ? darkCssVariables : cssVariables
    const root = document.documentElement
    Object.entries(variables).forEach(([key, value]) => {
      root.style.setProperty(key, value)
    })
    if (theme.value === 'dark') {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  // 切换主题
  function toggleTheme() {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
    localStorage.setItem('theme', theme.value)
    applyTheme()
  }

  // 设置主题
  function setTheme(newTheme: Theme) {
    theme.value = newTheme
    localStorage.setItem('theme', newTheme)
    applyTheme()
  }

  // 监听系统主题变化
  if (typeof window !== 'undefined') {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
      if (!localStorage.getItem('theme')) {
        theme.value = e.matches ? 'dark' : 'light'
        applyTheme()
      }
    })
  }

  onMounted(() => {
    loadTheme()
  })

  return {
    theme,
    toggleTheme,
    setTheme
  }
}
