/**
 * 数据库维护 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

/**
 * 备份请求
 */
export interface BackupRequest {
  connection_id: number
  database?: string
  tables?: string[]
  format?: 'sql' | 'csv'
}

/**
 * 备份响应
 */
export interface BackupResponse {
  backup_id: string
  download_url: string
  file_size: number
}

/**
 * 恢复请求
 */
export interface RestoreRequest {
  connection_id: number
  database?: string
  sql_file: string // SQL 文件内容
}

/**
 * 表维护操作
 */
export interface TableMaintenanceRequest {
  connection_id: number
  database: string
  table: string
  operation: 'optimize' | 'repair' | 'analyze'
}

/**
 * 备份数据库
 */
export async function backupDatabase(params: BackupRequest): Promise<BackupResponse> {
  const response = await request.post<ApiResponse<BackupResponse>>('/admin/backup', params)
  return response.data
}

/**
 * 恢复数据库
 */
export async function restoreDatabase(params: RestoreRequest): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/restore', params)
}

/**
 * 表优化
 */
export async function optimizeTable(params: TableMaintenanceRequest): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/tables/optimize', {
    database: params.database,
    table: params.table
  }, {
    params: { connection_id: params.connection_id }
  })
}

/**
 * 表修复
 */
export async function repairTable(params: TableMaintenanceRequest): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/tables/repair', {
    database: params.database,
    table: params.table
  }, {
    params: { connection_id: params.connection_id }
  })
}

/**
 * 表分析
 */
export async function analyzeTable(params: TableMaintenanceRequest): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/tables/analyze', {
    database: params.database,
    table: params.table
  }, {
    params: { connection_id: params.connection_id }
  })
}
