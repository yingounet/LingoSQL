/**
 * 权限管理相关类型定义
 */

export type PrivilegeType = 
  | 'SELECT'
  | 'INSERT'
  | 'UPDATE'
  | 'DELETE'
  | 'CREATE'
  | 'DROP'
  | 'ALTER'
  | 'INDEX'
  | 'REFERENCES'
  | 'TRIGGER'
  | 'EXECUTE'
  | 'USAGE'

export interface PermissionNode {
  type: 'database' | 'table' | 'column'
  name: string
  path: string
  privileges: PrivilegeType[]
  children?: PermissionNode[]
}

export interface GrantPermissionRequest {
  username: string
  host?: string // MySQL only
  target_type: 'database' | 'table' | 'column'
  target_name: string
  database?: string
  table?: string
  privileges: PrivilegeType[]
}

export interface RevokePermissionRequest {
  username: string
  host?: string // MySQL only
  target_type: 'database' | 'table' | 'column'
  target_name: string
  database?: string
  table?: string
  privileges: PrivilegeType[]
}
