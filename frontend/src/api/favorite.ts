/**
 * 收藏 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { Favorite, FavoriteListParams, FavoriteCreateRequest, FavoriteUpdateRequest } from '@/types/favorite'

/**
 * 获取收藏列表
 */
export async function getFavorites(params?: FavoriteListParams): Promise<Favorite[]> {
  const queryParams = new URLSearchParams()
  if (params?.connection_id != null) {
    queryParams.append('connection_id', String(params.connection_id))
  }
  if (params?.database != null && params.database !== '') {
    queryParams.append('database', params.database)
  }
  if (params?.sort) {
    queryParams.append('sort', params.sort)
  }
  const qs = queryParams.toString()
  const url = qs ? `/favorites?${qs}` : '/favorites'
  const response = await request.get<ApiResponse<Favorite[]>>(url)
  return response.data
}

/**
 * 创建收藏
 */
export async function createFavorite(body: FavoriteCreateRequest): Promise<Favorite> {
  const response = await request.post<ApiResponse<Favorite>>('/favorites', body)
  return response.data
}

/**
 * 更新收藏
 */
export async function updateFavorite(id: number, body: FavoriteUpdateRequest): Promise<Favorite> {
  const response = await request.put<ApiResponse<Favorite>>(`/favorites/${id}`, body)
  return response.data
}

/**
 * 删除收藏
 */
export async function deleteFavorite(id: number): Promise<void> {
  await request.delete<ApiResponse<void>>(`/favorites/${id}`)
}

/**
 * 记录收藏使用（更新最近使用时间）
 */
export async function recordFavoriteUse(id: number): Promise<void> {
  await request.post<ApiResponse<void>>(`/favorites/${id}/use`)
}
