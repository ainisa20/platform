import request from '@/utils/request'
import type {
  ShopResp,
  ShopListReq,
  ShopCreateReq,
  ShopUpdateReq,
  ShopStatusReq,
} from '@/types/system'
import type { ApiResponse } from '@/types/api'
import type { PageResult } from '@/types/system'

export function getShopList(params: ShopListReq) {
  return request.get<ApiResponse<PageResult<ShopResp>>>('/v1/platform/shops', { params })
}

export function getShop(id: number) {
  return request.get<ApiResponse<ShopResp>>(`/v1/platform/shops/${id}`)
}

export function createShop(data: ShopCreateReq) {
  return request.post<ApiResponse<ShopResp>>('/v1/platform/shops', data)
}

export function updateShop(id: number, data: ShopUpdateReq) {
  return request.put<ApiResponse<null>>(`/v1/platform/shops/${id}`, data)
}

export function deleteShop(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/platform/shops/${id}`)
}

export function updateShopStatus(id: number, data: ShopStatusReq) {
  return request.put<ApiResponse<null>>(`/v1/platform/shops/${id}/status`, data)
}

export function resetShopAdminPassword(id: number, newPassword: string) {
  return request.put<ApiResponse<null>>(`/v1/platform/shops/${id}/admin-password`, {
    new_password: newPassword,
  })
}
