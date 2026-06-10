package shop

import (
	"strconv"

	"platform/internal/model/dto"
	"platform/internal/pkg/response"
	shopsvc "platform/internal/service/shop"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FinanceReportCtrl struct {
	svc *shopsvc.FinanceReportService
}

func NewFinanceReportCtrl() *FinanceReportCtrl {
	return &FinanceReportCtrl{svc: shopsvc.NewFinanceReportService()}
}

func (ctrl *FinanceReportCtrl) Summary(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.FinanceReportReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := ctrl.svc.GetSummary(db, tenantID, req.StartDate, req.EndDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (ctrl *FinanceReportCtrl) Trend(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	months, _ := strconv.Atoi(c.DefaultQuery("months", "6"))

	result, err := ctrl.svc.GetTrend(db, tenantID, months)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (ctrl *FinanceReportCtrl) ProfitLoss(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.FinanceReportReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := ctrl.svc.GetProfitLoss(db, tenantID, req.StartDate, req.EndDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, result)
}
