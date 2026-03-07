/**
 * 数据库备份 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

/**
 * 创建备份请求
 */
export interface CreateBackupRequest {
  connection_id: number
  database: string
  tables?: string[]
  compress?: boolean
  schema_only?: boolean
  max_file_size_mb?: number
}

/**
 * 备份响应（同步或任务）
 */
export interface BackupResponse {
  backup_id?: string
  task_id?: number
  download_url?: string
  file_size?: number
}

/**
 * 任务详情
 */
export interface TaskInfo {
  id: number
  user_id: number
  type: string
  status: 'pending' | 'running' | 'success' | 'failed' | 'canceled'
  progress: number
  payload: string
  result: string
  error_message: string
  created_at: string
  updated_at: string
  started_at?: string
  finished_at?: string
}

/**
 * 任务列表响应
 */
export interface TaskListResponse {
  list: TaskInfo[]
  total: number
}

/**
 * 备份文件信息
 */
export interface BackupFileInfo {
  id: string
  name: string
  database: string
  size: number
  file_count: number
  created_at: string
  connection_id?: number
}

/**
 * 备份列表响应
 */
export interface BackupListResponse {
  list: BackupFileInfo[]
  total: number
}

/**
 * 创建备份（调用现有 /admin/backup，后端可能返回同步结果或 task_id）
 */
export async function createBackup(params: CreateBackupRequest): Promise<BackupResponse> {
  const response = await request.post<ApiResponse<BackupResponse>>('/admin/backup', {
    connection_id: params.connection_id,
    database: params.database,
    tables: params.tables,
    format: 'sql',
    compress: params.compress ?? true,
    schema_only: params.schema_only ?? false,
    max_file_size_mb: params.max_file_size_mb ?? 0
  })
  return response.data
}

/**
 * 获取任务详情
 */
export async function getTask(taskId: number): Promise<TaskInfo> {
  const response = await request.get<ApiResponse<TaskInfo>>(`/tasks/${taskId}`)
  return response.data
}

/**
 * 获取任务列表（含 type 过滤，前端可筛选 BACKUP_DATABASE）
 */
export async function getTasks(page = 1, pageSize = 20): Promise<TaskListResponse> {
  const response = await request.get<ApiResponse<TaskListResponse>>('/tasks', {
    params: { page, page_size: pageSize }
  })
  return response.data
}

/**
 * 取消任务
 */
export async function cancelTask(taskId: number): Promise<void> {
  await request.post<ApiResponse<void>>(`/tasks/${taskId}/cancel`)
}

/**
 * 获取备份文件列表（后端实现后启用）
 */
export async function listBackups(connectionId?: number, page = 1, pageSize = 20): Promise<BackupListResponse> {
  const response = await request.get<ApiResponse<BackupListResponse>>('/admin/backups', {
    params: { connection_id: connectionId, page, page_size: pageSize }
  })
  return response.data
}

/**
 * 下载备份（相对路径，需与 request baseURL 一致）
 */
export function getBackupDownloadUrl(backupId: string): string {
  return `/api/admin/backups/${encodeURIComponent(backupId)}/download`
}

/**
 * 删除备份
 */
export async function deleteBackup(backupId: string): Promise<void> {
  await request.delete<ApiResponse<void>>(`/admin/backups/${backupId}`)
}

/**
 * 从备份恢复
 */
export interface RestoreBackupRequest {
  connection_id: number
  database: string
  backup_id?: string
  file_path?: string
}

export async function restoreBackup(params: RestoreBackupRequest): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/restore', params)
}
