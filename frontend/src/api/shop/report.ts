import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { FinanceSummaryResp, FinanceTrendItem, ProfitLossCategory } from '@/types/system'

export function getFinanceSummary(start?: string, end?: string) {
  return request.get<ApiResponse<FinanceSummaryResp>>('/v1/shop/finance/reports/summary', {
    params: { start_date: start, end_date: end },
  })
}

export function getFinanceTrend(months?: number) {
  return request.get<ApiResponse<FinanceTrendItem[]>>('/v1/shop/finance/reports/trend', {
    params: { months },
  })
}

export function getFinanceProfitLoss(start?: string, end?: string) {
  return request.get<ApiResponse<{ categories: ProfitLossCategory[] }>>('/v1/shop/finance/reports/profit-loss', {
    params: { start_date: start, end_date: end },
  })
}
