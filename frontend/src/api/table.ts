/**
 * 表管理 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

/**
 * 表信息
 */
export interface TableInfo {
  name: string
  engine?: string
  rows?: number
  data_length?: number
  index_length?: number
  column_count?: number
}

/**
 * 表详细信息
 */
export interface TableDetailInfo {
  name: string
  engine: string
  rows: number
  data_length: number
  index_length: number
  collation: string
  auto_increment: number | null
  comment: string
  create_time: string
  update_time: string
}

/**
 * 获取指定数据库的表列表
 */
export async function getTables(connectionId: number, database: string): Promise<TableInfo[]> {
  const response = await request.get<ApiResponse<TableInfo[]>>('/tables', {
    params: {
      connection_id: connectionId,
      database: database
    },
    timeout: 10000
  })
  return response.data
}

/**
 * 获取表详细信息
 */
export async function getTableInfo(connectionId: number, database: string, table: string): Promise<TableDetailInfo> {
  const response = await request.get<ApiResponse<TableDetailInfo>>('/tables/info', {
    params: {
      connection_id: connectionId,
      database: database,
      table: table
    }
  })
  return response.data
}
