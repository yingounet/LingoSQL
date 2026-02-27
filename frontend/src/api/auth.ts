import request from '@/utils/request'
import type { ApiResponse, User } from '@/types/api'
import type { LoginRequest, RegisterRequest, UpdateProfileRequest, ChangePasswordRequest } from '@/types/model'

export function login(data: LoginRequest) {
  return request.post<ApiResponse<{ user: User; token: string }>>('/auth/login', data)
}

export function register(data: RegisterRequest) {
  return request.post<ApiResponse<{ user: User; token: string }>>('/auth/register', data)
}

export function logout() {
  return request.post<ApiResponse>('/auth/logout')
}

export function getMe() {
  return request.get<ApiResponse<User>>('/auth/me')
}

/** 更新当前用户信息。请求体: { username, email }；响应与 GET /auth/me 一致 */
export function updateProfile(data: UpdateProfileRequest) {
  return request.patch<ApiResponse<User>>('/auth/me', data)
}

/** 修改登录密码。请求体: { current_password, new_password }；成功后保持当前 token 有效 */
export function changePassword(data: ChangePasswordRequest) {
  return request.post<ApiResponse<void>>('/auth/change-password', data)
}
