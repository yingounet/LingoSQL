export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface User {
  id: number
  username: string
  email: string
  created_at: string
}

export interface Connection {
  id: number
  name: string
  db_type: 'mysql' | 'postgresql'
  host: string
  port: number
  database: string
  username: string
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface QueryResult {
  columns: string[]
  rows: any[][]
  execution_time_ms: number
  rows_affected: number
  statements_executed: number
  query_id?: number
}
