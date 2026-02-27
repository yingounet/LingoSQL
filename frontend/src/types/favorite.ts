/**
 * 收藏类型定义
 */

export interface Favorite {
  id: number
  connection_id: number
  connection_name?: string
  database?: string
  name: string
  sql_query: string
  description?: string
  created_at: string
  last_used_at?: string | null
}

export interface FavoriteListParams {
  connection_id?: number | null
  database?: string | null
  sort?: 'created_at' | 'last_used_at'
}

export interface FavoriteCreateRequest {
  connection_id: number
  database?: string
  name: string
  sql_query: string
  description?: string
}

export interface FavoriteUpdateRequest {
  name?: string
  sql_query?: string
  description?: string
}
