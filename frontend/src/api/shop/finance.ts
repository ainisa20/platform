import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type {
  ShopFinCategoryResp,
  ShopFinCategoryAvailableResp,
  ShopFinCategorySyncReq,
} from '@/types/system'

export function getShopFinCategories(params?: { category_type?: number | null }) {
  return request.get<ApiResponse<ShopFinCategoryResp[]>>('/v1/shop/finance/categories', { params })
}

export function getAvailableFinCategories() {
  return request.get<ApiResponse<ShopFinCategoryAvailableResp[]>>('/v1/shop/finance/categories/available')
}

export function syncFinCategories(data: ShopFinCategorySyncReq) {
  return request.post<ApiResponse<null>>('/v1/shop/finance/categories/sync', data)
}

export function cancelSyncFinCategory(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/shop/finance/categories/${id}`)
}
