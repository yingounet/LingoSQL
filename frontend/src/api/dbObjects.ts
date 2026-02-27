/**
 * 数据库对象管理 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

/**
 * 视图信息
 */
export interface ViewInfo {
  name: string
  definition: string
  created_at?: string
  updated_at?: string
}

/**
 * 存储过程信息
 */
export interface ProcedureInfo {
  name: string
  definition: string
  parameters?: string
  created_at?: string
  updated_at?: string
}

/**
 * 函数信息
 */
export interface FunctionInfo {
  name: string
  definition: string
  return_type: string
  parameters?: string
  created_at?: string
  updated_at?: string
}

/**
 * 触发器信息
 */
export interface TriggerInfo {
  name: string
  event: string
  table: string
  timing: string
  definition: string
  created_at?: string
}

/**
 * 事件信息（MySQL）
 */
export interface EventInfo {
  name: string
  definition: string
  status: string
  on_completion: string
  created_at?: string
}

/**
 * 获取视图列表
 */
export async function getViews(connectionId: number, database: string): Promise<ViewInfo[]> {
  const response = await request.get<ApiResponse<ViewInfo[]>>('/admin/views', {
    params: { connection_id: connectionId, database }
  })
  return response.data
}

/**
 * 创建视图
 */
export async function createView(
  connectionId: number,
  database: string,
  name: string,
  definition: string
): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/views', {
    database,
    name,
    definition
  }, {
    params: { connection_id: connectionId }
  })
}

/**
 * 删除视图
 */
export async function dropView(connectionId: number, database: string, name: string): Promise<void> {
  await request.delete<ApiResponse<void>>(`/admin/views/${name}`, {
    params: { connection_id: connectionId, database }
  })
}

/**
 * 获取存储过程列表
 */
export async function getProcedures(connectionId: number, database: string): Promise<ProcedureInfo[]> {
  const response = await request.get<ApiResponse<ProcedureInfo[]>>('/admin/procedures', {
    params: { connection_id: connectionId, database }
  })
  return response.data
}

/**
 * 创建存储过程
 */
export async function createProcedure(
  connectionId: number,
  database: string,
  name: string,
  definition: string
): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/procedures', {
    database,
    name,
    definition
  }, {
    params: { connection_id: connectionId }
  })
}

/**
 * 删除存储过程
 */
export async function dropProcedure(connectionId: number, database: string, name: string): Promise<void> {
  await request.delete<ApiResponse<void>>(`/admin/procedures/${name}`, {
    params: { connection_id: connectionId, database }
  })
}

/**
 * 执行存储过程
 */
export async function executeProcedure(
  connectionId: number,
  database: string,
  name: string,
  parameters: unknown[] = []
): Promise<unknown> {
  const response = await request.post<ApiResponse<unknown>>(`/admin/procedures/${name}/execute`, {
    database,
    parameters
  }, {
    params: { connection_id: connectionId }
  })
  return response.data
}

/**
 * 获取函数列表
 */
export async function getFunctions(connectionId: number, database: string): Promise<FunctionInfo[]> {
  const response = await request.get<ApiResponse<FunctionInfo[]>>('/admin/functions', {
    params: { connection_id: connectionId, database }
  })
  return response.data
}

/**
 * 创建函数
 */
export async function createFunction(
  connectionId: number,
  database: string,
  name: string,
  definition: string,
  returnType: string
): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/functions', {
    database,
    name,
    definition,
    return_type: returnType
  }, {
    params: { connection_id: connectionId }
  })
}

/**
 * 删除函数
 */
export async function dropFunction(connectionId: number, database: string, name: string): Promise<void> {
  await request.delete<ApiResponse<void>>(`/admin/functions/${name}`, {
    params: { connection_id: connectionId, database }
  })
}

/**
 * 获取触发器列表
 */
export async function getTriggers(connectionId: number, database: string): Promise<TriggerInfo[]> {
  const response = await request.get<ApiResponse<TriggerInfo[]>>('/admin/triggers', {
    params: { connection_id: connectionId, database }
  })
  return response.data
}

/**
 * 创建触发器
 */
export async function createTrigger(
  connectionId: number,
  database: string,
  name: string,
  table: string,
  event: string,
  timing: string,
  definition: string
): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/triggers', {
    database,
    name,
    table,
    event,
    timing,
    definition
  }, {
    params: { connection_id: connectionId }
  })
}

/**
 * 删除触发器
 */
export async function dropTrigger(connectionId: number, database: string, name: string): Promise<void> {
  await request.delete<ApiResponse<void>>(`/admin/triggers/${name}`, {
    params: { connection_id: connectionId, database }
  })
}

/**
 * 获取事件列表（MySQL）
 */
export async function getEvents(connectionId: number, database: string): Promise<EventInfo[]> {
  const response = await request.get<ApiResponse<EventInfo[]>>('/admin/events', {
    params: { connection_id: connectionId, database }
  })
  return response.data
}

/**
 * 创建事件（MySQL）
 */
export async function createEvent(
  connectionId: number,
  database: string,
  name: string,
  definition: string
): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/events', {
    database,
    name,
    definition
  }, {
    params: { connection_id: connectionId }
  })
}

/**
 * 删除事件（MySQL）
 */
export async function dropEvent(connectionId: number, database: string, name: string): Promise<void> {
  await request.delete<ApiResponse<void>>(`/admin/events/${name}`, {
    params: { connection_id: connectionId, database }
  })
}
