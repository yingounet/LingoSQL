/**
 * 连接管理 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type {
  Connection,
  ConnectionFormData,
  ConnectionListResponse,
  ConnectionListResponseRaw,
  ConnectionListItem,
  TestResult,
  DbType,
  ConnectionDetailResponse,
} from '@/types/connection'

/**
 * 获取连接列表
 */
export async function getConnections(params?: {
  page?: number
  page_size?: number
  db_type?: DbType | null
  search?: string
}): Promise<ConnectionListResponse> {
  const queryParams: Record<string, any> = {}
  
  if (params?.page) queryParams.page = params.page
  if (params?.page_size) queryParams.page_size = params.page_size
  if (params?.db_type) queryParams.db_type = params.db_type
  if (params?.search) queryParams.search = params.search

  const response = await request.get<ApiResponse<ConnectionListResponseRaw>>('/connections', {
    params: queryParams,
  })
  
  // 转换扁平化的列表项为嵌套格式
  const list = response.data.list.map(convertListItemToConnection)
  
  return {
    list,
    total: response.data.total,
    page: response.data.page,
    page_size: response.data.page_size,
  }
}

/**
 * 获取单个连接详情
 */
export async function getConnection(id: number): Promise<Connection | null> {
  try {
    const response = await request.get<ApiResponse<ConnectionDetailResponse>>(`/connections/${id}`)
    return convertDetailToConnection(response.data)
  } catch (error) {
    return null
  }
}

/**
 * 创建连接
 */
export async function createConnection(data: ConnectionFormData): Promise<Connection> {
  const requestData = buildConnectionRequest(data)
  const response = await request.post<ApiResponse<ConnectionDetailResponse>>('/connections', requestData)
  return convertDetailToConnection(response.data)
}

/**
 * 更新连接
 */
export async function updateConnection(id: number, data: ConnectionFormData): Promise<Connection> {
  const requestData = buildConnectionRequest(data)
  const response = await request.put<ApiResponse<ConnectionDetailResponse>>(`/connections/${id}`, requestData)
  return convertDetailToConnection(response.data)
}

/**
 * 删除连接
 */
export async function deleteConnection(id: number): Promise<void> {
  await request.delete(`/connections/${id}`)
}

/**
 * 测试已保存的连接
 */
export async function testConnection(id: number): Promise<TestResult> {
  try {
    const response = await request.post<ApiResponse<TestResult>>(`/connections/${id}/test`)
    return response.data
  } catch (error: any) {
    // 处理错误响应
    const errorMessage = error.response?.data?.message || error.message || '连接测试失败'
    return {
      connected: false,
      error: errorMessage,
    }
  }
}

/**
 * 测试连接配置（未保存的连接）
 */
export async function testConnectionConfig(data: {
  db_type: DbType
  connection_type: 'direct' | 'ssh_tunnel'
  db_config: ConnectionFormData['db_config']
  ssh_config?: ConnectionFormData['ssh_config']
}): Promise<TestResult> {
  try {
    const requestData = {
      db_type: data.db_type,
      connection_type: data.connection_type,
      db_config: {
        host: data.db_config.host,
        port: data.db_config.port,
        database: data.db_config.database || '',
        username: data.db_config.username,
        password: data.db_config.password,
        options: data.db_config.options,
      },
      ssh_config: data.connection_type === 'ssh_tunnel' && data.ssh_config ? {
        host: data.ssh_config.host,
        port: data.ssh_config.port,
        username: data.ssh_config.username,
        auth_type: data.ssh_config.auth_type,
        password: data.ssh_config.auth_type === 'password' ? data.ssh_config.password : undefined,
        private_key: data.ssh_config.auth_type === 'private_key' ? data.ssh_config.private_key : undefined,
        passphrase: data.ssh_config.auth_type === 'private_key' ? data.ssh_config.passphrase : undefined,
      } : undefined,
    }

    const response = await request.post<ApiResponse<TestResult>>('/connections/test', requestData)
    return response.data
  } catch (error: any) {
    const errorMessage = error.response?.data?.message || error.message || '连接测试失败'
    return {
      connected: false,
      error: errorMessage,
    }
  }
}

/**
 * 设置默认连接
 */
export async function setDefaultConnection(id: number): Promise<void> {
  await request.put(`/connections/${id}/default`)
}

/**
 * 更新连接的最后使用时间
 */
export async function updateLastUsed(id: number): Promise<void> {
  await request.put(`/connections/${id}/last-used`)
}

// ============ 辅助函数 ============

/**
 * 将列表项转换为 Connection 类型
 */
function convertListItemToConnection(item: ConnectionListItem): Connection {
  return {
    id: item.id,
    name: item.name,
    db_type: item.db_type,
    connection_type: item.connection_type,
    db_config: {
      host: item.host,
      port: item.port,
      database: item.database,
      username: '', // 列表响应不包含 username
    },
    is_default: item.is_default,
    last_used_at: item.last_used_at || null,
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

/**
 * 构建连接请求数据
 */
function buildConnectionRequest(data: ConnectionFormData) {
  return {
    name: data.name,
    db_type: data.db_type,
    connection_type: data.connection_type,
    db_config: {
      host: data.db_config.host,
      port: data.db_config.port,
      database: data.db_config.database || '',
      username: data.db_config.username,
      password: data.db_config.password,
      options: {
        ssl_mode: data.db_config.options.ssl_mode,
        charset: data.db_config.options.charset,
        timeout: data.db_config.options.timeout,
      },
    },
    ssh_config: data.connection_type === 'ssh_tunnel' ? {
      host: data.ssh_config.host,
      port: data.ssh_config.port,
      username: data.ssh_config.username,
      auth_type: data.ssh_config.auth_type,
      password: data.ssh_config.auth_type === 'password' ? data.ssh_config.password : undefined,
      private_key: data.ssh_config.auth_type === 'private_key' ? data.ssh_config.private_key : undefined,
      passphrase: data.ssh_config.auth_type === 'private_key' ? data.ssh_config.passphrase : undefined,
    } : undefined,
  }
}

/**
 * 将详情响应转换为 Connection 类型
 */
function convertDetailToConnection(detail: ConnectionDetailResponse): Connection {
  return {
    id: detail.id,
    name: detail.name,
    db_type: detail.db_type,
    connection_type: detail.connection_type,
    db_config: {
      host: detail.db_config?.host || '',
      port: detail.db_config?.port || 0,
      database: detail.db_config?.database,
      username: detail.db_config?.username || '',
      options: detail.db_config?.options ? {
        ssl_mode: detail.db_config.options.ssl_mode,
        charset: detail.db_config.options.charset,
        timeout: detail.db_config.options.timeout,
      } : undefined,
    },
    ssh_config: detail.ssh_config ? {
      host: detail.ssh_config.host,
      port: detail.ssh_config.port,
      username: detail.ssh_config.username,
      auth_type: detail.ssh_config.auth_type,
    } : undefined,
    is_default: detail.is_default,
    last_used_at: detail.last_used_at || null,
    created_at: detail.created_at,
    updated_at: detail.updated_at,
  }
}
