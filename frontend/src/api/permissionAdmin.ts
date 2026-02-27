/**
 * 权限管理 API
 * 注意：这些API需要后端实现，当前为占位实现
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { PermissionNode, GrantPermissionRequest, RevokePermissionRequest } from '@/types/permissionAdmin'

/**
 * 获取用户完整权限树
 */
export async function getPermissionTree(connectionId: number, username: string, host?: string): Promise<PermissionNode[]> {
  const response = await request.get<ApiResponse<PermissionNode[]>>('/admin/permissions/tree', {
    params: { 
      connection_id: connectionId,
      username,
      host
    }
  })
  return response.data
}

/**
 * 授予权限
 */
export async function grantPermission(connectionId: number, data: GrantPermissionRequest): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/permissions/grant', data, {
    params: { connection_id: connectionId }
  })
}

/**
 * 撤销权限
 */
export async function revokePermission(connectionId: number, data: RevokePermissionRequest): Promise<void> {
  await request.post<ApiResponse<void>>('/admin/permissions/revoke', data, {
    params: { connection_id: connectionId }
  })
}
