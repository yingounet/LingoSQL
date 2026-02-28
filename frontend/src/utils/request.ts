import axios from 'axios'
import { useAuthStore } from '@/store/auth'
import router from '@/router'

type ApiError = Error & {
  status?: number
  code?: number
  requestId?: string
  data?: unknown
}

function buildApiError(message: string, status?: number, code?: number, requestId?: string, data?: unknown): ApiError {
  const err = new Error(message) as ApiError
  err.status = status
  err.code = code
  err.requestId = requestId
  err.data = data
  return err
}

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
        return Promise.reject(
          buildApiError(
            data.message || '请求失败',
            response.status,
            data.code,
            data.request_id,
            data,
          ),
        )
      }
      // 返回data字段
      return data
    }
    // 否则直接返回响应数据
    return response.data
  },
  (error) => {
    if (error.response) {
      const responseData = error.response.data
      const status = error.response.status
      if (error.response.status === 401) {
        const authStore = useAuthStore()
        authStore.logout()
        router.push('/login')
      }
      // 处理业务错误（code不为200的情况）
      if (responseData && responseData.code !== 200) {
        return Promise.reject(
          buildApiError(
            responseData.message || '请求失败',
            status,
            responseData.code,
            responseData.request_id,
            responseData,
          ),
        )
      }
      if (status) {
        return Promise.reject(
          buildApiError(
            error.message || '请求失败',
            status,
            responseData?.code,
            responseData?.request_id,
            responseData,
          ),
        )
      }
    }
    return Promise.reject(error)
  }
)

export default request
