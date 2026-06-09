import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type {
  FinanceCategoryResp,
  FinanceCategoryListReq,
  FinanceCategoryCreateReq,
  FinanceCategoryUpdateReq,
} from '@/types/system'

export function getFinanceCategoryList(params?: FinanceCategoryListReq) {
  return request.get<ApiResponse<FinanceCategoryResp[]>>('/v1/platform/finance-categories', { params })
}

export function createFinanceCategory(data: FinanceCategoryCreateReq) {
  return request.post<ApiResponse<FinanceCategoryResp>>('/v1/platform/finance-categories', data)
}

export function updateFinanceCategory(id: number, data: FinanceCategoryUpdateReq) {
  return request.put<ApiResponse<null>>(`/v1/platform/finance-categories/${id}`, data)
}

export function deleteFinanceCategory(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/platform/finance-categories/${id}`)
}
