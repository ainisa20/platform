package shop

import (
	"fmt"

	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type FinanceReportService struct{}

func NewFinanceReportService() *FinanceReportService {
	return &FinanceReportService{}
}

func (s *FinanceReportService) GetSummary(db *gorm.DB, tenantID uint64, startDate, endDate string) (*dto.FinanceSummaryResp, error) {
	q := db.Model(&entity.FinanceRecord{}).
		Where("tenant_id = ? AND review_status = 2", tenantID)
	q = applyDateFilter(q, startDate, endDate)

	var result dto.FinanceSummaryResp
	err := q.Select(
		"COALESCE(SUM(CASE WHEN record_type = 1 THEN actual_amount ELSE 0 END), 0) as total_income",
		"COALESCE(SUM(CASE WHEN record_type = 2 THEN actual_amount ELSE 0 END), 0) as total_expense",
	).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	result.NetProfit = result.TotalIncome - result.TotalExpense
	return &result, nil
}

func (s *FinanceReportService) GetTrend(db *gorm.DB, tenantID uint64, months int) ([]dto.FinanceTrendItem, error) {
	if months <= 0 {
		months = 6
	}

	interval := fmt.Sprintf("%d months", months)
	var items []dto.FinanceTrendItem
	err := db.Model(&entity.FinanceRecord{}).
		Select(
			"TO_CHAR(record_date, 'YYYY-MM') as month",
			"COALESCE(SUM(CASE WHEN record_type = 1 THEN actual_amount ELSE 0 END), 0) as income",
			"COALESCE(SUM(CASE WHEN record_type = 2 THEN actual_amount ELSE 0 END), 0) as expense",
		).
		Where("tenant_id = ? AND review_status = 2 AND record_date >= NOW() - ?::interval", tenantID, interval).
		Group("month").
		Order("month ASC").
		Scan(&items).Error
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []dto.FinanceTrendItem{}
	}
	return items, nil
}

func (s *FinanceReportService) GetProfitLoss(db *gorm.DB, tenantID uint64, startDate, endDate string) (*dto.FinanceProfitLossResp, error) {
	q := db.Model(&entity.FinanceRecord{}).
		Where("tenant_id = ? AND review_status = 2", tenantID)
	q = applyDateFilter(q, startDate, endDate)

	var categories []dto.ProfitLossCategory
	err := q.Select(
		"category_l1 as name",
		"CASE WHEN record_type = 1 THEN 'income' ELSE 'expense' END as type",
		"COALESCE(SUM(actual_amount), 0) as subtotal",
	).
		Group("category_l1, record_type").
		Order("type, subtotal DESC").
		Scan(&categories).Error
	if err != nil {
		return nil, err
	}
	if categories == nil {
		categories = []dto.ProfitLossCategory{}
	}
	ptrs := make([]*dto.ProfitLossCategory, len(categories))
	for i := range categories {
		ptrs[i] = &categories[i]
	}
	return &dto.FinanceProfitLossResp{Categories: ptrs}, nil
}

func applyDateFilter(q *gorm.DB, startDate, endDate string) *gorm.DB {
	if startDate != "" {
		q = q.Where("record_date >= ?", startDate)
	}
	if endDate != "" {
		q = q.Where("record_date <= ?", endDate)
	}
	return q
}
