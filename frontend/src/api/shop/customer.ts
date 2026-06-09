import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { PageResult } from '@/types/system'
import type {
  ShopCustomerResp,
  ShopCustomerListReq,
  ShopCustomerCreateReq,
  ShopCustomerUpdateReq,
  ShopCustomerOrderResp,
} from '@/types/system'

export function getCustomerList(params?: ShopCustomerListReq) {
  return request.get<ApiResponse<PageResult<ShopCustomerResp>>>('/v1/shop/customers', { params })
}

export function getCustomer(id: number) {
  return request.get<ApiResponse<ShopCustomerResp>>(`/v1/shop/customers/${id}`)
}

export function createCustomer(data: ShopCustomerCreateReq) {
  return request.post<ApiResponse<ShopCustomerResp>>('/v1/shop/customers', data)
}

export function updateCustomer(id: number, data: ShopCustomerUpdateReq) {
  return request.put<ApiResponse<null>>(`/v1/shop/customers/${id}`, data)
}

export function deleteCustomer(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/shop/customers/${id}`)
}

export function exportCustomers(params?: ShopCustomerListReq) {
  return request.get<ApiResponse<ShopCustomerResp[]>>('/v1/shop/customers/export', { params })
}

export function getCustomerOrders(id: number) {
  return request.get<ApiResponse<ShopCustomerOrderResp[]>>(`/v1/shop/customers/${id}/orders`)
}
