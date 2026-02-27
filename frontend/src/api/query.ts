/**
 * SQL 查询 API
 */

import request from '@/utils/request'
import type { ApiResponse, QueryResult } from '@/types/api'

/**
 * 查询执行请求参数
 */
export interface QueryExecuteRequest {
  connection_id: number
  database?: string
  sql: string
}

/**
 * 查询执行响应（复用 QueryResult 类型）
 */
export type QueryExecuteResponse = QueryResult

/**
 * 查询错误响应
 */
export interface QueryErrorResponse {
  error: string
  query_id?: number
}

/**
 * 执行计划请求参数
 */
export interface ExplainRequest {
  connection_id: number
  database?: string
  sql: string
}

/**
 * 执行计划响应
 */
export interface ExplainResponse {
  plan: Record<string, unknown>[]
  execution_time_ms: number
}

/**
 * 多语句执行请求参数
 */
export interface MultiQueryExecuteRequest {
  connection_id: number
  database?: string
  sql: string
  transaction?: boolean // 是否在事务中执行
}

/**
 * 多语句执行响应
 */
export interface MultiQueryExecuteResponse {
  results: Array<{
    sql: string
    result?: QueryResult
    error?: string
    execution_time_ms: number
  }>
  total_time_ms: number
}

/**
 * 执行 SQL 查询
 */
export async function executeQuery(params: QueryExecuteRequest): Promise<QueryExecuteResponse> {
  const response = await request.post<ApiResponse<QueryExecuteResponse>>('/query/execute', params)
  return response.data
}

/**
 * 获取 SQL 执行计划
 */
export async function explainQuery(params: ExplainRequest): Promise<ExplainResponse> {
  const response = await request.post<ApiResponse<ExplainResponse>>('/query/explain', params)
  return response.data
}

/**
 * 执行多个 SQL 语句
 */
export async function executeMultiQuery(params: MultiQueryExecuteRequest): Promise<MultiQueryExecuteResponse> {
  const response = await request.post<ApiResponse<MultiQueryExecuteResponse>>('/query/execute-multi', params)
  return response.data
}

/**
 * 开始事务
 */
export async function beginTransaction(connectionId: number, database?: string): Promise<void> {
  await request.post<ApiResponse<void>>('/query/transaction/begin', {
    connection_id: connectionId,
    database,
  })
}

/**
 * 提交事务
 */
export async function commitTransaction(connectionId: number, database?: string): Promise<void> {
  await request.post<ApiResponse<void>>('/query/transaction/commit', {
    connection_id: connectionId,
    database,
  })
}

/**
 * 回滚事务
 */
export async function rollbackTransaction(connectionId: number, database?: string): Promise<void> {
  await request.post<ApiResponse<void>>('/query/transaction/rollback', {
    connection_id: connectionId,
    database,
  })
}
