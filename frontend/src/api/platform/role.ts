import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

export interface PermissionResp {
  id: number
  parent_id: number
  system_type: string
  name: string
  type: number
  path: string
  component: string
  perms_code: string
  icon: string
  sort: number
  visible: boolean
  status: number
  children: PermissionResp[]
}

export interface RoleResp {
  id: number
  tenant_id: number
  role_name: string
  role_code: string
  data_scope: number
  sort: number
  status: number
  remark: string
  created_at: string
  created_by: number
  updated_at: string
  updated_by: number
  permissions: PermissionResp[]
}

export interface RoleCreateReq {
  role_name: string
  role_code: string
  remark?: string
  data_scope?: number
  sort?: number
  status?: number
}

export interface RoleUpdateReq {
  role_name: string
  role_code: string
  remark?: string
  data_scope?: number
  sort?: number
  status?: number
}

export interface RoleListReq {
  page?: number
  page_size?: number
  role_name?: string
  status?: number | null
}

export interface PaginatedData<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export function getRoleById(id: number) {
  return request.get<ApiResponse<RoleResp>>(`/v1/platform/sys/roles/${id}`)
}

export function getRoleList(params: RoleListReq) {
  return request.get<ApiResponse<PaginatedData<RoleResp>>>('/v1/platform/sys/roles', { params })
}

export function createRole(data: RoleCreateReq) {
  return request.post<ApiResponse<RoleResp>>('/v1/platform/sys/roles', data)
}

export function updateRole(id: number, data: RoleUpdateReq) {
  return request.put<ApiResponse<null>>(`/v1/platform/sys/roles/${id}`, data)
}

export function deleteRole(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/platform/sys/roles/${id}`)
}

export function assignRolePermissions(id: number, permission_ids: number[]) {
  return request.put<ApiResponse<null>>(`/v1/platform/sys/roles/${id}/permissions`, { permission_ids })
}

export function getPermissionTree() {
  return request.get<ApiResponse<PermissionResp[]>>('/v1/platform/sys/permissions')
}
