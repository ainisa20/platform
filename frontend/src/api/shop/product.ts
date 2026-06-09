import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { PageResult } from '@/types/system'
import type {
  ShopProductResp,
  ShopProductListReq,
  ShopProductSelectReq,
  ShopProductPriceReq,
  ShopProductStatusReq,
  ShopPlatformProductResp,
} from '@/types/system'

export function getShopProducts(params?: ShopProductListReq) {
  return request.get<ApiResponse<PageResult<ShopProductResp>>>('/v1/shop/products', { params })
}

export function getPlatformProducts() {
  return request.get<ApiResponse<ShopPlatformProductResp[]>>('/v1/shop/products/platform')
}

export function selectProducts(data: ShopProductSelectReq) {
  return request.post<ApiResponse<null>>('/v1/shop/products', data)
}

export function updateShopProductPrice(id: number, data: ShopProductPriceReq) {
  return request.put<ApiResponse<null>>(`/v1/shop/products/${id}/price`, data)
}

export function updateShopProductStatus(id: number, data: ShopProductStatusReq) {
  return request.put<ApiResponse<null>>(`/v1/shop/products/${id}/status`, data)
}

export function cancelShopProduct(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/shop/products/${id}`)
}
