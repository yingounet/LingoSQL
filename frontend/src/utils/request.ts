import axios from 'axios'
import { useAuthStore } from '@/store/auth'
import router from '@/router'

const request = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    // 如果响应是ApiResponse格式，检查code
    const data = response.data
    if (data && typeof data === 'object' && 'code' in data) {
      // 如果code不是200，抛出错误
      if (data.code !== 200) {
        return Promise.reject(new Error(data.message || '请求失败'))
      }
      // 返回data字段
      return data
    }
    // 否则直接返回响应数据
    return response.data
  },
  (error) => {
    if (error.response) {
      if (error.response.status === 401) {
        const authStore = useAuthStore()
        authStore.logout()
        router.push('/login')
      }
      // 处理业务错误（code不为200的情况）
      if (error.response.data && error.response.data.code !== 200) {
        return Promise.reject(new Error(error.response.data.message || '请求失败'))
      }
    }
    return Promise.reject(error)
  }
)

export default request
