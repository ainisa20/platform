package platform

import (
	"strconv"

	"platform/internal/model/dto"
	"platform/internal/pkg/response"
	platformsvc "platform/internal/service/platform"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FinanceReportCtrl struct {
	svc *platformsvc.FinanceReportService
}

func NewFinanceReportCtrl(svc *platformsvc.FinanceReportService) *FinanceReportCtrl {
	return &FinanceReportCtrl{svc: svc}
}

func (ctrl *FinanceReportCtrl) Summary(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req dto.PlatformFinanceReportReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tenantIDs := ctrl.resolveTenantIDs(req.ShopID)
	result, err := ctrl.svc.GetSummary(db, tenantIDs, req.StartDate, req.EndDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (ctrl *FinanceReportCtrl) Trend(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req dto.PlatformFinanceReportReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	months, _ := strconv.Atoi(c.DefaultQuery("months", "6"))

	tenantIDs := ctrl.resolveTenantIDs(req.ShopID)
	result, err := ctrl.svc.GetTrend(db, tenantIDs, months)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (ctrl *FinanceReportCtrl) ProfitLoss(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req dto.PlatformFinanceReportReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tenantIDs := ctrl.resolveTenantIDs(req.ShopID)
	result, err := ctrl.svc.GetProfitLoss(db, tenantIDs, req.StartDate, req.EndDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (ctrl *FinanceReportCtrl) PerShop(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	result, err := ctrl.svc.GetPerShop(db, startDate, endDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (ctrl *FinanceReportCtrl) resolveTenantIDs(shopID *uint64) []uint64 {
	if shopID == nil {
		return nil
	}
	return []uint64{*shopID}
}
