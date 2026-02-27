/**
 * 数据库管理相关类型定义
 */

export interface AdminPermission {
  has_database_admin: boolean
  has_user_admin: boolean
  has_permission_admin: boolean
}

export interface DatabaseInfo {
  name: string
  charset?: string
  collation?: string
  encoding?: string
  size?: number
  table_count?: number
  created_at?: string
}

export interface CreateDatabaseRequest {
  name: string
  charset?: string
  collation?: string
  encoding?: string
  lc_collate?: string
  lc_ctype?: string
}

export interface RenameDatabaseRequest {
  new_name: string
}
