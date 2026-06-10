import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { PageResult } from '@/types/system'
import type {
  FinanceRecordResp,
  FinanceRecordListReq,
  FinanceRecordCreateReq,
  FinanceRecordUpdateReq,
  FinanceReviewReq,
  FinanceAttachmentResp,
} from '@/types/system'

export function getRecordList(params?: FinanceRecordListReq) {
  return request.get<ApiResponse<PageResult<FinanceRecordResp>>>('/v1/shop/finance/records', { params })
}

export function getRecord(id: number) {
  return request.get<ApiResponse<FinanceRecordResp>>(`/v1/shop/finance/records/${id}`)
}

export function createRecord(data: FinanceRecordCreateReq) {
  return request.post<ApiResponse<FinanceRecordResp>>('/v1/shop/finance/records', data)
}

export function updateRecord(id: number, data: FinanceRecordUpdateReq) {
  return request.put<ApiResponse<null>>(`/v1/shop/finance/records/${id}`, data)
}

export function deleteRecord(id: number) {
  return request.delete<ApiResponse<null>>(`/v1/shop/finance/records/${id}`)
}

export function reviewRecord(id: number, data: FinanceReviewReq) {
  return request.post<ApiResponse<null>>(`/v1/shop/finance/records/${id}/review`, data)
}

export function exportRecords(params?: FinanceRecordListReq) {
  return request.get<ApiResponse<FinanceRecordResp[]>>('/v1/shop/finance/records/export', { params })
}

export function getRecordAttachments(id: number) {
  return request.get<ApiResponse<FinanceAttachmentResp[]>>(`/v1/shop/finance/records/${id}/attachments`)
}

export function createRecordAttachment(id: number, formData: FormData) {
  return request.post<ApiResponse<FinanceAttachmentResp>>(`/v1/shop/finance/records/${id}/attachments`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function downloadRecordAttachment(recordId: number, attId: number) {
  return request.get<ApiResponse<{ url: string }>>(`/v1/shop/finance/records/${recordId}/attachments/${attId}/download`)
}