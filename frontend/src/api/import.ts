/**
 * 数据导入 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

/**
 * 导入数据请求参数
 */
export interface ImportDataRequest {
  connection_id: number
  database: string
  table: string
  data: unknown[][]
  headers: string[]
  field_mapping?: Record<string, string> // 字段映射：文件列名 -> 表字段名
  skip_first_row?: boolean // 是否跳过第一行（表头）
  on_duplicate?: 'ignore' | 'update' | 'error' // 重复数据处理方式
}

/**
 * 导入数据响应
 */
export interface ImportDataResponse {
  total_rows: number
  inserted_rows: number
  updated_rows: number
  error_rows: number
  errors?: Array<{ row: number; error: string }>
}

/**
 * 导入数据到表
 */
export async function importData(params: ImportDataRequest): Promise<ImportDataResponse> {
  const response = await request.post<ApiResponse<ImportDataResponse>>('/import/data', params)
  return response.data
}

/**
 * 执行 SQL 文件
 */
export interface ExecuteSQLFileRequest {
  connection_id: number
  database?: string
  sql: string
  transaction?: boolean
}

export interface ExecuteSQLFileResponse {
  executed_statements: number
  success_count: number
  error_count: number
  errors?: Array<{ statement: number; error: string }>
}

export async function executeSQLFile(params: ExecuteSQLFileRequest): Promise<ExecuteSQLFileResponse> {
  const response = await request.post<ApiResponse<ExecuteSQLFileResponse>>('/import/sql', params)
  return response.data
}
