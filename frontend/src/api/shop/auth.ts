import request from '@/utils/request'
import type { ApiResponse, ShopLoginParams, LoginResult, UserInfo, RefreshParams } from '@/types/api'

export function login(data: ShopLoginParams) {
  return request.post<ApiResponse<LoginResult>>('/v1/shop/auth/login', data)
}

export function logout() {
  return request.post<ApiResponse<null>>('/v1/shop/auth/logout')
}

export function getUserInfo() {
  return request.get<ApiResponse<UserInfo>>('/v1/shop/auth/userinfo')
}

export function getPermissions() {
  return request.get<ApiResponse<string[]>>('/v1/shop/auth/permissions')
}

export function refreshToken(data: RefreshParams) {
  return request.post<ApiResponse<LoginResult>>('/v1/shop/auth/refresh', data)
}
