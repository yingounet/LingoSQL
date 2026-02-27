/**
 * 数据库管理 API
 * 注意：这些API需要后端实现，当前为占位实现
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { AdminPermission, DatabaseInfo, CreateDatabaseRequest, RenameDatabaseRequest } from '@/types/databaseAdmin'

/**
 * 检查当前连接用户的管理权限
 */
export async function checkAdminPermissions(connectionId: number): Promise<AdminPermission> {
  const response = await request.get<ApiResponse<AdminPermission>>('/admin/permissions/check', {
    params: { connection_id: connectionId }
  })
  return response.data
}

/**
 * 获取数据库详细信息列表
 */
export async function getDatabaseList(connectionId: number): Promise<DatabaseInfo[]> {
  const response = await request.get<ApiResponse<DatabaseInfo[]>>('/admin/databases', {
    params: { connection_id: connectionId }
  })
  return response.data
}

/**
 * 获取单个数据库详情
 */
export async function getDatabaseInfo(connectionId: number, databaseName: string): Promise<DatabaseInfo> {
  const response = await request.get<ApiResponse<DatabaseInfo>>(`/admin/databases/${databaseName}/info`, {
    params: { connection_id: connectionId }
  })
  return response.data
}

/**
 * 创建数据库
 */
export async function createDatabase(connectionId: number, data: CreateDatabaseRequest): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/databases', data, {
    params: { connection_id: connectionId }
  })
}

/**
 * 删除数据库
 */
export async function deleteDatabase(connectionId: number, databaseName: string): Promise<void> {
  await request.delete<ApiResponse<void>>(`/admin/databases/${databaseName}`, {
    params: { connection_id: connectionId }
  })
}

/**
 * 重命名数据库（仅PostgreSQL）
 */
export async function renameDatabase(connectionId: number, databaseName: string, data: RenameDatabaseRequest): Promise<void> {
  await request.post<ApiResponse<void>>(`/admin/databases/${databaseName}/rename`, data, {
    params: { connection_id: connectionId }
  })
}
