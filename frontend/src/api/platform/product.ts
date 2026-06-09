import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { PageResult } from '@/types/system'
import type {
  ProductCategoryResp,
  ProductCategoryListReq,
  ProductCategoryCreateReq,
  ProductCategoryUpdateReq,
  ProductResp,
  ProductListReq,
  ProductCreateReq,
  ProductUpdateReq,
  WorkflowSaveReq,
} from '@/types/system'

export function getCategoryList(params?: ProductCategoryListReq) {
  return request.get<ApiResponse<PageResult<ProductCategoryResp>>>('/v1/platform/product-categories', { params })
}

export function createCategory(data: ProductCategoryCreateReq) {
  return request.post<ApiResponse<ProductCategoryResp>>('/v1/platform/product-categories', data)
}

export function updateCategory(id: number, data: ProductCategoryUpdateReq) {
  return request.put<ApiResponse<null>>(`/v1/platform/product-categories/${id}`, data)
}

export function deleteCategory(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/platform/product-categories/${id}`)
}

export function getProductList(params?: ProductListReq) {
  return request.get<ApiResponse<PageResult<ProductResp>>>('/v1/platform/products', { params })
}

export function getProduct(id: number) {
  return request.get<ApiResponse<ProductResp>>(`/v1/platform/products/${id}`)
}

export function createProduct(data: ProductCreateReq) {
  return request.post<ApiResponse<ProductResp>>('/v1/platform/products', data)
}

export function updateProduct(id: number, data: ProductUpdateReq) {
  return request.put<ApiResponse<null>>(`/v1/platform/products/${id}`, data)
}

export function deleteProduct(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/platform/products/${id}`)
}

export function updateProductStatus(id: number, data: { status: number }) {
  return request.put<ApiResponse<null>>(`/v1/platform/products/${id}/status`, data)
}

export function getProductWorkflow(productId: number) {
  return request.get<ApiResponse<import('@/types/system').WorkflowNodeResp[]>>(`/v1/platform/products/${productId}/workflow`)
}

export function saveProductWorkflow(productId: number, data: WorkflowSaveReq) {
  return request.put<ApiResponse<null>>(`/v1/platform/products/${productId}/workflow`, data)
}
