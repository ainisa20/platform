import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type {
  FinanceSummaryResp,
  FinanceTrendItem,
  ProfitLossCategory,
  FinanceReportShopSummary,
  PlatformFinanceReportReq,
} from '@/types/system'

export function getPlatformFinanceSummary(params?: PlatformFinanceReportReq) {
  return request.get<ApiResponse<FinanceSummaryResp>>('/v1/platform/finance/reports/summary', { params })
}

export function getPlatformFinanceTrend(params?: PlatformFinanceReportReq & { months?: number }) {
  return request.get<ApiResponse<FinanceTrendItem[]>>('/v1/platform/finance/reports/trend', { params })
}

export function getPlatformFinanceProfitLoss(params?: PlatformFinanceReportReq) {
  return request.get<ApiResponse<{ categories: ProfitLossCategory[] }>>('/v1/platform/finance/reports/profit-loss', { params })
}

export function getPlatformFinancePerShop(params?: { start_date?: string; end_date?: string }) {
  return request.get<ApiResponse<FinanceReportShopSummary[]>>('/v1/platform/finance/reports/shops', { params })
}
