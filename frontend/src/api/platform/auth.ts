import request from '@/utils/request'
import type { ApiResponse, LoginParams, LoginResult, UserInfo, RefreshParams } from '@/types/api'

export function login(data: LoginParams) {
  return request.post<ApiResponse<LoginResult>>('/v1/platform/auth/login', data)
}

export function logout() {
  return request.post<ApiResponse<null>>('/v1/platform/auth/logout')
}

export function getUserInfo() {
  return request.get<ApiResponse<UserInfo>>('/v1/platform/auth/userinfo')
}

export function getPermissions() {
  return request.get<ApiResponse<string[]>>('/v1/platform/auth/permissions')
}

export function refreshToken(data: RefreshParams) {
  return request.post<ApiResponse<LoginResult>>('/v1/platform/auth/refresh', data)
}
