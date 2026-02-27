import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { storage } from '@/utils/storage'
import { login, register, getMe } from '@/api/auth'
import type { User, LoginRequest, RegisterRequest } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(storage.get<string>('token'))
  const user = ref<User | null>(storage.get<User>('user'))

  const isAuthenticated = computed(() => !!token.value)

  function setAuth(newToken: string, newUser: User) {
    token.value = newToken
    user.value = newUser
    storage.set('token', newToken)
    storage.set('user', newUser)
  }

  function logout() {
    token.value = null
    user.value = null
    storage.remove('token')
    storage.remove('user')
  }

  async function loginUser(credentials: LoginRequest) {
    const response = await login(credentials)
    setAuth(response.data.token, response.data.user)
    return response
  }

  async function registerUser(credentials: RegisterRequest) {
    const response = await register(credentials)
    setAuth(response.data.token, response.data.user)
    return response
  }

  async function fetchUser() {
    if (!token.value) return
    try {
      const response = await getMe()
      user.value = response.data
      storage.set('user', response.data)
    } catch (error) {
      logout()
    }
  }

  return {
    token,
    user,
    isAuthenticated,
    setAuth,
    logout,
    loginUser,
    registerUser,
    fetchUser,
  }
})
