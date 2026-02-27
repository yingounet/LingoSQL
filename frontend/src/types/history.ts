/**
 * 历史记录类型定义
 */

/**
 * 查询历史记录
 */
export interface QueryHistory {
  id: number
  connection_id: number
  connection_name?: string
  sql_query: string
  operation_type?: string // 系统执行记录才有此字段
  execution_time_ms: number
  rows_affected: number
  success: boolean
  error_message?: string
  created_at: string
}

/**
 * 查询历史列表响应
 */
export interface QueryHistoryListResponse {
  list: QueryHistory[]
  total: number
  page: number
  page_size: number
}
