import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type {
  UserResp,
  UserCreateReq,
  UserUpdateReq,
  UserListReq,
  PasswordResetReq,
  AssignRolesReq,
  PageResult,
  RoleResp,
  DeptResp,
} from '@/types/system'

export function getUserList(params?: UserListReq) {
  return request.get<ApiResponse<PageResult<UserResp>>>('/v1/platform/sys/users', { params })
}

export function getUser(id: number) {
  return request.get<ApiResponse<UserResp>>(`/v1/platform/sys/users/${id}`)
}

export function createUser(data: UserCreateReq) {
  return request.post<ApiResponse<UserResp>>('/v1/platform/sys/users', data)
}

export function updateUser(id: number, data: UserUpdateReq) {
  return request.put<ApiResponse<null>>(`/v1/platform/sys/users/${id}`, data)
}

export function deleteUser(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/platform/sys/users/${id}`)
}

export function assignUserRoles(id: number, data: AssignRolesReq) {
  return request.put<ApiResponse<null>>(`/v1/platform/sys/users/${id}/roles`, data)
}

export function resetUserPassword(id: number, data: PasswordResetReq) {
  return request.put<ApiResponse<null>>(`/v1/platform/sys/users/${id}/password`, data)
}

export function getAssignableRoles() {
  return request.get<ApiResponse<RoleResp[]>>('/v1/platform/sys/roles/assignable')
}

export function getRoleList(params?: { page?: number; page_size?: number }) {
  return request.get<ApiResponse<PageResult<RoleResp>>>('/v1/platform/sys/roles', { params })
}

export function getDeptTree() {
  return request.get<ApiResponse<DeptResp[]>>('/v1/platform/sys/depts')
}
