import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

/** 部门响应 */
export interface DeptResp {
  id: number
  tenant_id: number
  parent_id: number
  ancestors: string
  dept_name: string
  sort: number
  leader: string
  phone: string
  status: number
  created_at: string
  created_by: number
  updated_at: string
  updated_by: number
  children: DeptResp[]
}

/** 创建部门请求 */
export interface DeptCreateReq {
  parent_id: number
  dept_name: string
  leader?: string
  phone?: string
  sort?: number
  status?: number
}

/** 更新部门请求 */
export interface DeptUpdateReq {
  parent_id?: number
  dept_name?: string
  leader?: string
  phone?: string
  sort?: number
  status?: number
}

/** 获取部门树 */
export function getDeptTree() {
  return request.get<ApiResponse<DeptResp[]>>('/v1/platform/sys/depts')
}

/** 创建部门 */
export function createDept(data: DeptCreateReq) {
  return request.post<ApiResponse<DeptResp>>('/v1/platform/sys/depts', data)
}

/** 更新部门 */
export function updateDept(id: number, data: DeptUpdateReq) {
  return request.put<ApiResponse<null>>(`/v1/platform/sys/depts/${id}`, data)
}

/** 删除部门 */
export function deleteDept(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/platform/sys/depts/${id}`)
}
