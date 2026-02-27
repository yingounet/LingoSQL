/**
 * 用户管理相关类型定义
 */

export interface DatabaseUser {
  username: string
  host?: string // MySQL only
  role?: string // PostgreSQL only
  privileges?: string
  created_at?: string
}

export interface CreateUserRequest {
  username: string
  host?: string // MySQL only
  password: string
  is_superuser?: boolean // PostgreSQL only
}

export interface ChangePasswordRequest {
  username: string
  host?: string // MySQL only
  new_password: string
}

export interface DeleteUserRequest {
  username: string
  host?: string // MySQL only
}
