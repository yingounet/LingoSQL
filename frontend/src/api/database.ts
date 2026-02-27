/**
 * 数据库管理 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

/**
 * 获取指定连接的数据库列表
 */
export async function getDatabases(connectionId: number): Promise<string[]> {
  const response = await request.get<ApiResponse<string[]>>('/databases', {
    params: { connection_id: connectionId }
  })
  return response.data
}
