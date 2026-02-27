export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
}

export interface UpdateProfileRequest {
  username: string
  email: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}

export interface ConnectionCreateRequest {
  name: string
  db_type: 'mysql' | 'postgresql'
  host: string
  port: number
  database: string
  username: string
  password: string
}

export interface QueryExecuteRequest {
  connection_id: number
  database?: string
  sql: string
}
