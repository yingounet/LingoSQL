/**
 * 表数据 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

/**
 * 筛选条件
 */
export interface FilterCondition {
  field: string
  operator: string
  value: string
}

/**
 * 分页参数
 */
export interface PaginationParams {
  page: number
  pageSize: number
}

/**
 * 列定义（前端）
 */
export interface ColumnDefFrontend {
  name: string
  type: string
  isPrimary: boolean
  isIndex: boolean
}

/**
 * 行数据结果
 */
export interface RowDataResult {
  columns: ColumnDefFrontend[]
  rows: Record<string, unknown>[]
  total: number
  page: number
  pageSize: number
}

/**
 * 获取表数据
 */
export async function getTableData(
  connectionId: number,
  database: string,
  table: string,
  filters: FilterCondition[] = [],
  pagination: PaginationParams = { page: 1, pageSize: 100 }
): Promise<RowDataResult> {
  const response = await request.post<ApiResponse<RowDataResult>>('/tables/data', {
    connection_id: connectionId,
    database,
    table,
    filters,
    ...pagination
  })
  return response.data
}

/**
 * 更新表行数据
 */
export async function updateTableRow(
  connectionId: number,
  database: string,
  table: string,
  primaryKey: Record<string, unknown>,
  updateData: Record<string, unknown>
): Promise<number> {
  const response = await request.post<ApiResponse<{ affected_rows: number }>>('/tables/data/update', {
    connection_id: connectionId,
    database,
    table,
    primary_key: primaryKey,
    update_data: updateData
  })
  return response.data.affected_rows
}

/**
 * 批量插入数据
 */
export interface BatchInsertRequest {
  connection_id: number
  database: string
  table: string
  data: unknown[][]
  columns: string[]
}

export async function batchInsertData(params: BatchInsertRequest): Promise<{ inserted_rows: number }> {
  const response = await request.post<ApiResponse<{ inserted_rows: number }>>('/tables/data/batch-insert', params)
  return response.data
}

/**
 * 批量更新数据
 */
export interface BatchUpdateRequest {
  connection_id: number
  database: string
  table: string
  filters: FilterCondition[]
  update_data: Record<string, unknown>
}

export async function batchUpdateData(params: BatchUpdateRequest): Promise<{ affected_rows: number }> {
  const response = await request.post<ApiResponse<{ affected_rows: number }>>('/tables/data/batch-update', params)
  return response.data
}

/**
 * 批量删除数据
 */
export interface BatchDeleteRequest {
  connection_id: number
  database: string
  table: string
  filters: FilterCondition[]
}

export async function batchDeleteData(params: BatchDeleteRequest): Promise<{ affected_rows: number }> {
  const response = await request.post<ApiResponse<{ affected_rows: number }>>('/tables/data/batch-delete', params)
  return response.data
}

/**
 * 数据对比请求
 */
export interface CompareDataRequest {
  connection_id: number
  database1: string
  table1: string
  database2?: string
  table2: string
  key_columns: string[]
}

export interface CompareDataResponse {
  only_in_table1: Record<string, unknown>[]
  only_in_table2: Record<string, unknown>[]
  different: Array<{
    key: Record<string, unknown>
    table1_data: Record<string, unknown>
    table2_data: Record<string, unknown>
    differences: string[]
  }>
  same_count: number
}

export async function compareData(params: CompareDataRequest): Promise<CompareDataResponse> {
  const response = await request.post<ApiResponse<CompareDataResponse>>('/tables/data/compare', params)
  return response.data
}

/**
 * 查找替换请求
 */
export interface FindReplaceRequest {
  connection_id: number
  database: string
  table: string
  column: string
  find_value: string
  replace_value: string
  filters?: FilterCondition[]
}

export interface FindReplaceResponse {
  affected_rows: number
  matched_rows: number
}

export async function findReplaceData(params: FindReplaceRequest): Promise<FindReplaceResponse> {
  const response = await request.post<ApiResponse<FindReplaceResponse>>('/tables/data/find-replace', params)
  return response.data
}

/**
 * 从索引信息中提取索引字段
 */
export function extractIndexFields(indexes: Array<{ columns: string[] }>): string[] {
  const fields = new Set<string>()
  indexes.forEach(index => {
    index.columns.forEach(col => fields.add(col))
  })
  return Array.from(fields)
}
