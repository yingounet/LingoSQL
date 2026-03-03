import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { User } from '@/types/api'

export interface InstallStatus {
  installed: boolean
  allow_registration?: boolean
}

export interface InstallSetupRequest {
  admin: {
    username: string
    email: string
    password: string
  }
  settings: {
    site_name: string
    allow_registration: boolean
    rate_limit_enabled: boolean
    rate_limit_default_rpm: number
    rate_limit_polling_rpm: number
    cors_allowed_origins: string[]
  }
}

export interface InstallSetupResponse {
  user: User
  token: string
  refresh_token: string
}

export function getInstallStatus() {
  return request.get<ApiResponse<InstallStatus>>('/install/status')
}

export function setupInstall(data: InstallSetupRequest) {
  return request.post<ApiResponse<InstallSetupResponse>>('/install/setup', data)
}
