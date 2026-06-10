import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { PageResult } from '@/types/system'
import type {
  OrderResp,
  OrderListReq,
  OrderCreateReq,
  OrderWorkflowAdvanceReq,
  OrderWorkflowLogResp,
  OrderWorkflowNodeResp,
  OrderAttachmentResp,
} from '@/types/system'

export function getOrderList(params?: OrderListReq) {
  return request.get<ApiResponse<PageResult<OrderResp>>>('/v1/shop/orders', { params })
}

export function getOrder(id: number) {
  return request.get<ApiResponse<OrderResp>>(`/v1/shop/orders/${id}`)
}

export function createOrder(data: OrderCreateReq) {
  return request.post<ApiResponse<OrderResp>>('/v1/shop/orders', data)
}

export function cancelOrder(id: number) {
  return request.put<ApiResponse<null>>(`/v1/shop/orders/${id}/cancel`)
}

export function cancelOrderItem(id: number, itemId: number) {
  return request.put<ApiResponse<null>>(`/v1/shop/orders/${id}/items/${itemId}/cancel`)
}

export function exportOrders(params?: OrderListReq) {
  return request.get<ApiResponse<OrderResp[]>>('/v1/shop/orders/export', { params })
}

export function getItemWorkflow(orderId: number, itemId: number) {
  return request.get<ApiResponse<OrderWorkflowNodeResp[]>>(`/v1/shop/orders/${orderId}/items/${itemId}/workflow`)
}

export function getItemWorkflowLogs(orderId: number, itemId: number) {
  return request.get<ApiResponse<OrderWorkflowLogResp[]>>(`/v1/shop/orders/${orderId}/items/${itemId}/workflow/logs`)
}

export function advanceItemWorkflow(orderId: number, itemId: number, data: OrderWorkflowAdvanceReq) {
  return request.post<ApiResponse<{ workflow_log_id: number }>>(`/v1/shop/orders/${orderId}/items/${itemId}/workflow/advance`, data)
}

export function getItemAttachments(orderId: number, itemId: number) {
  return request.get<ApiResponse<OrderAttachmentResp[]>>(`/v1/shop/orders/${orderId}/items/${itemId}/attachments`)
}

export function createItemAttachment(orderId: number, itemId: number, formData: FormData) {
  return request.post<ApiResponse<OrderAttachmentResp>>(`/v1/shop/orders/${orderId}/items/${itemId}/attachments`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function getItemAttachmentDownloadURL(orderId: number, itemId: number, attId: number) {
  return request.get<ApiResponse<{ id: number; url: string }>>(`/v1/shop/orders/${orderId}/items/${itemId}/attachments/${attId}`)
}
