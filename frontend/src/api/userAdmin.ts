/**
 * 用户管理 API
 * 注意：这些API需要后端实现，当前为占位实现
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { DatabaseUser, CreateUserRequest, ChangePasswordRequest, DeleteUserRequest } from '@/types/userAdmin'

/**
 * 获取用户列表
 */
export async function getUserList(connectionId: number): Promise<DatabaseUser[]> {
  const response = await request.get<ApiResponse<DatabaseUser[]>>('/admin/users', {
    params: { connection_id: connectionId }
  })
  return response.data
}

/**
 * 创建用户
 */
export async function createUser(connectionId: number, data: CreateUserRequest): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/users', data, {
    params: { connection_id: connectionId }
  })
}

/**
 * 删除用户
 */
export async function deleteUser(connectionId: number, data: DeleteUserRequest): Promise<void> {
  await request.delete<ApiResponse<void>>('/admin/users', {
    params: { connection_id: connectionId, ...data }
  })
}

/**
 * 修改用户密码
 */
export async function changeUserPassword(connectionId: number, data: ChangePasswordRequest): Promise<void> {
  await request.put<ApiResponse<void>>('/admin/users/password', data, {
    params: { connection_id: connectionId }
  })
}

/**
 * 获取用户权限
 */
export async function getUserGrants(connectionId: number, username: string, host?: string): Promise<string[]> {
  const response = await request.get<ApiResponse<string[]>>('/admin/users/grants', {
    params: { 
      connection_id: connectionId,
      username,
      host
    }
  })
  return response.data
}
