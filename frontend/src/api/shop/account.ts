import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { PageResult } from '@/types/system'
import type {
  ShopFinAccountResp,
  ShopFinAccountListReq,
  ShopFinAccountCreateReq,
  ShopFinAccountUpdateReq,
} from '@/types/system'

export function getFinAccountList(params?: ShopFinAccountListReq) {
  return request.get<ApiResponse<PageResult<ShopFinAccountResp>>>('/v1/shop/finance/accounts', { params })
}

export function createFinAccount(data: ShopFinAccountCreateReq) {
  return request.post<ApiResponse<ShopFinAccountResp>>('/v1/shop/finance/accounts', data)
}

export function updateFinAccount(id: number, data: ShopFinAccountUpdateReq) {
  return request.put<ApiResponse<null>>(`/v1/shop/finance/accounts/${id}`, data)
}

export function deleteFinAccount(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/shop/finance/accounts/${id}`)
}
