package platform

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

func (s *FinanceReportService) GetSummary(db *gorm.DB, tenantIDs []uint64, startDate, endDate string) (*dto.FinanceSummaryResp, error) {
	q := db.Model(&entity.FinanceRecord{}).
		Where("review_status = 2")
	if len(tenantIDs) > 0 {
		q = q.Where("tenant_id IN ?", tenantIDs)
	}
	q = applyReportDateFilter(q, startDate, endDate)

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

func (s *FinanceReportService) GetTrend(db *gorm.DB, tenantIDs []uint64, months int) ([]dto.FinanceTrendItem, error) {
	if months <= 0 {
		months = 6
	}

	interval := fmt.Sprintf("%d months", months)
	var items []dto.FinanceTrendItem
	q := db.Model(&entity.FinanceRecord{}).
		Select(
			"TO_CHAR(record_date, 'YYYY-MM') as month",
			"COALESCE(SUM(CASE WHEN record_type = 1 THEN actual_amount ELSE 0 END), 0) as income",
			"COALESCE(SUM(CASE WHEN record_type = 2 THEN actual_amount ELSE 0 END), 0) as expense",
		).
		Where("review_status = 2 AND record_date >= NOW() - ?::interval", interval)
	if len(tenantIDs) > 0 {
		q = q.Where("tenant_id IN ?", tenantIDs)
	}
	err := q.Group("month").
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

func (s *FinanceReportService) GetProfitLoss(db *gorm.DB, tenantIDs []uint64, startDate, endDate string) (*dto.FinanceProfitLossResp, error) {
	q := db.Model(&entity.FinanceRecord{}).
		Where("review_status = 2")
	if len(tenantIDs) > 0 {
		q = q.Where("tenant_id IN ?", tenantIDs)
	}
	q = applyReportDateFilter(q, startDate, endDate)

	type row struct {
		CategoryL1 string
		CategoryL2 string
		CategoryL3 string
		RecordType int16
		Subtotal   float64
	}
	var rows []row
	err := q.Select(
		"category_l1",
		"category_l2",
		"category_l3",
		"record_type",
		"COALESCE(SUM(actual_amount), 0) as subtotal",
	).
		Group("category_l1, category_l2, category_l3, record_type").
		Order("record_type, category_l1, category_l2, category_l3").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	type l1Key struct {
		name  string
		rtype string
	}
	type l2Key struct {
		l1 l1Key
		l2 string
	}

	l1Map := make(map[l1Key]*dto.ProfitLossCategory)
	l2Map := make(map[l2Key]*dto.ProfitLossCategory)

	for _, r := range rows {
		rtype := "income"
		if r.RecordType == 2 {
			rtype = "expense"
		}
		k1 := l1Key{name: r.CategoryL1, rtype: rtype}
		k2 := l2Key{l1: k1, l2: r.CategoryL2}

		if _, ok := l1Map[k1]; !ok {
			l1Map[k1] = &dto.ProfitLossCategory{
				Name: r.CategoryL1,
				Type: rtype,
			}
		}
		if _, ok := l2Map[k2]; !ok {
			l2Node := &dto.ProfitLossCategory{
				Name: r.CategoryL2,
				Type: rtype,
			}
			l2Map[k2] = l2Node
			l1Map[k1].Children = append(l1Map[k1].Children, l2Node)
		}
		l3Node := &dto.ProfitLossCategory{
			Name:     r.CategoryL3,
			Type:     rtype,
			Subtotal: r.Subtotal,
		}
		l2Map[k2].Children = append(l2Map[k2].Children, l3Node)
		l2Map[k2].Subtotal += r.Subtotal
		l1Map[k1].Subtotal += r.Subtotal
	}

	var categories []*dto.ProfitLossCategory
	var income, expense []l1Key
	for k := range l1Map {
		if k.rtype == "income" {
			income = append(income, k)
		} else {
			expense = append(expense, k)
		}
	}
	for _, k := range append(income, expense...) {
		categories = append(categories, l1Map[k])
	}
	if categories == nil {
		categories = []*dto.ProfitLossCategory{}
	}
	return &dto.FinanceProfitLossResp{Categories: categories}, nil
}

func (s *FinanceReportService) GetPerShop(db *gorm.DB, startDate, endDate string) ([]dto.FinanceReportShopSummary, error) {
	q := db.Model(&entity.FinanceRecord{}).
		Where("review_status = 2")
	q = applyReportDateFilter(q, startDate, endDate)

	var summaries []dto.FinanceReportShopSummary
	err := q.Select(
		"tenant_id as shop_id",
		"COALESCE(SUM(CASE WHEN record_type = 1 THEN actual_amount ELSE 0 END), 0) as income",
		"COALESCE(SUM(CASE WHEN record_type = 2 THEN actual_amount ELSE 0 END), 0) as expense",
	).
		Group("tenant_id").
		Order("tenant_id").
		Scan(&summaries).Error
	if err != nil {
		return nil, err
	}

	if len(summaries) > 0 {
		shopIDs := make([]uint64, len(summaries))
		for i := range summaries {
			shopIDs[i] = summaries[i].ShopID
			summaries[i].NetProfit = summaries[i].Income - summaries[i].Expense
		}
		type shopRow struct {
			ID       uint64
			ShopName string
		}
		var rows []shopRow
		if err := db.Table("sys_shop").
			Select("id, shop_name").
			Where("id IN ?", shopIDs).
			Scan(&rows).Error; err == nil {
			nameMap := make(map[uint64]string, len(rows))
			for _, r := range rows {
				nameMap[r.ID] = r.ShopName
			}
			for i := range summaries {
				if name, ok := nameMap[summaries[i].ShopID]; ok {
					summaries[i].ShopName = name
				}
			}
		}
	}
	if summaries == nil {
		summaries = []dto.FinanceReportShopSummary{}
	}
	return summaries, nil
}

func applyReportDateFilter(q *gorm.DB, startDate, endDate string) *gorm.DB {
	if startDate != "" {
		q = q.Where("record_date >= ?", startDate)
	}
	if endDate != "" {
		q = q.Where("record_date <= ?", endDate)
	}
	return q
}
