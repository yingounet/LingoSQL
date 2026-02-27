/**
 * 表管理 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { ColumnInfo, IndexInfo } from '@/api/schema'

/**
 * 创建表请求
 */
export interface CreateTableRequest {
  database: string
  table_name: string
  create_ddl?: string
}

/**
 * 创建表响应
 */
export interface CreateTableResponse {
  table_name: string
}

/**
 * 修改表结构请求
 */
export interface AlterTableRequest {
  database: string
  table: string
  operations: Array<{
    type: 'add_column' | 'drop_column' | 'modify_column' | 'rename_column'
    column?: ColumnInfo
    old_column_name?: string
    new_column_name?: string
  }>
}

/**
 * 创建索引请求
 */
export interface CreateIndexRequest {
  database: string
  table: string
  index: IndexInfo
}

/**
 * 删除索引请求
 */
export interface DropIndexRequest {
  database: string
  table: string
  index_name: string
}

/**
 * 重命名表请求
 */
export interface RenameTableRequest {
  database: string
  old_name: string
  new_name: string
}

/**
 * 创建表
 */
export async function createTable(
  connectionId: number,
  data: CreateTableRequest
): Promise<CreateTableResponse> {
  const response = await request.post<ApiResponse<CreateTableResponse>>('/admin/tables', data, {
    params: { connection_id: connectionId }
  })
  return response.data
}

/**
 * 删除表
 */
export async function dropTable(
  connectionId: number,
  database: string,
  table: string
): Promise<void> {
  await request.delete<ApiResponse<void>>(`/admin/tables/${table}`, {
    params: {
      connection_id: connectionId,
      database
    }
  })
}

/**
 * 修改表结构
 */
export async function alterTable(
  connectionId: number,
  data: AlterTableRequest
): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/tables/alter', data, {
    params: { connection_id: connectionId }
  })
}

/**
 * 创建索引
 */
export async function createIndex(
  connectionId: number,
  data: CreateIndexRequest
): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/tables/indexes', data, {
    params: { connection_id: connectionId }
  })
}

/**
 * 删除索引
 */
export async function dropIndex(
  connectionId: number,
  data: DropIndexRequest
): Promise<void> {
  await request.delete<ApiResponse<void>>(`/admin/tables/indexes/${data.index_name}`, {
    params: {
      connection_id: connectionId,
      database: data.database,
      table: data.table
    }
  })
}

/**
 * 重命名表
 */
export async function renameTable(
  connectionId: number,
  data: RenameTableRequest
): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/tables/rename', data, {
    params: { connection_id: connectionId }
  })
}
