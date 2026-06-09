package shop

import (
	"net/http"

	"platform/internal/model/dto"
	"platform/internal/pkg/response"
	shopsvc "platform/internal/service/shop"
	"platform/internal/service/shared"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ShopCustomerCtrl struct {
	svc *shopsvc.ShopCustomerService
}

func NewShopCustomerCtrl(svc *shopsvc.ShopCustomerService) *ShopCustomerCtrl {
	return &ShopCustomerCtrl{svc: svc}
}

func (ctrl *ShopCustomerCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.ShopCustomerListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	list, total, err := ctrl.svc.List(c, db, tenantID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	response.Page(c, list, total, page, pageSize)
}

func (ctrl *ShopCustomerCtrl) Get(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	customer, err := ctrl.svc.Get(c, db, id, tenantID)
	if err != nil {
		handleShopCustomerError(c, err)
		return
	}
	response.OK(c, customer)
}

func (ctrl *ShopCustomerCtrl) Create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")

	var req dto.ShopCustomerCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	customer, err := ctrl.svc.Create(c, db, tenantID, userID, &req)
	if err != nil {
		handleShopCustomerError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.R{Code: 0, Message: "success", Data: customer})
}

func (ctrl *ShopCustomerCtrl) Update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.ShopCustomerUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Update(c, db, id, tenantID, userID, &req); err != nil {
		handleShopCustomerError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *ShopCustomerCtrl) Delete(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.Delete(c, db, id, tenantID); err != nil {
		handleShopCustomerError(c, err)
		return
	}
	response.OKMsg(c, "删除成功")
}

func (ctrl *ShopCustomerCtrl) Export(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.ShopCustomerListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	list, err := ctrl.svc.Export(c, db, tenantID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (ctrl *ShopCustomerCtrl) ListOrders(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	orders, err := ctrl.svc.ListOrders(c, db, tenantID, id)
	if err != nil {
		handleShopCustomerError(c, err)
		return
	}
	response.OK(c, orders)
}

func handleShopCustomerError(c *gin.Context, err error) {
	switch err {
	case shared.ErrShopCustomerNotFound:
		response.NotFound(c, err.Error())
	case shared.ErrShopCustomerNameExists, shared.ErrShopCustomerHasOrders:
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}