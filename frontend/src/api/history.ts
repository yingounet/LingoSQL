/**
 * 历史记录 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { QueryHistory, QueryHistoryListResponse } from '@/types/history'

/**
 * 获取用户执行历史记录列表
 */
export async function getUserHistory(params: {
  connection_id?: number
  page?: number
  page_size?: number
}): Promise<QueryHistoryListResponse> {
  const queryParams = new URLSearchParams()
  if (params.connection_id) {
    queryParams.append('connection_id', String(params.connection_id))
  }
  if (params.page) {
    queryParams.append('page', String(params.page))
  }
  if (params.page_size) {
    queryParams.append('page_size', String(params.page_size))
  }

  const response = await request.get<ApiResponse<QueryHistoryListResponse>>(
    `/history?${queryParams.toString()}`
  )
  return response.data
}

/**
 * 获取系统执行历史记录列表
 */
export async function getSystemHistory(params: {
  connection_id: number
  page?: number
  page_size?: number
}): Promise<QueryHistoryListResponse> {
  const queryParams = new URLSearchParams()
  queryParams.append('connection_id', String(params.connection_id))
  if (params.page) {
    queryParams.append('page', String(params.page))
  }
  if (params.page_size) {
    queryParams.append('page_size', String(params.page_size))
  }

  const response = await request.get<ApiResponse<QueryHistoryListResponse>>(
    `/system-history?${queryParams.toString()}`
  )
  return response.data
}

/**
 * 获取单条历史记录（用户执行）
 */
export async function getHistoryById(id: number): Promise<QueryHistory> {
  const response = await request.get<ApiResponse<QueryHistory>>(`/history/${id}`)
  return response.data
}

/**
 * 删除用户历史记录
 */
export async function deleteHistory(id: number): Promise<void> {
  await request.delete<ApiResponse<void>>(`/history/${id}`)
}

/**
 * 删除系统历史记录
 */
export async function deleteSystemHistory(id: number): Promise<void> {
  await request.delete<ApiResponse<void>>(`/system-history/${id}`)
}
