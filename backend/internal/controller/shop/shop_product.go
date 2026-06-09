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

type ShopProductCtrl struct {
	svc *shopsvc.ShopProductService
}

func NewShopProductCtrl(svc *shopsvc.ShopProductService) *ShopProductCtrl {
	return &ShopProductCtrl{svc: svc}
}

func (ctrl *ShopProductCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.ShopProductListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	list, total, err := ctrl.svc.List(db, tenantID, &req)
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

func (ctrl *ShopProductCtrl) ListPlatform(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	list, err := ctrl.svc.ListPlatformAvailable(db, tenantID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (ctrl *ShopProductCtrl) Select(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")

	var req dto.ShopProductSelectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Select(db, tenantID, userID, &req); err != nil {
		handleShopProductError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.R{Code: 0, Message: "选品成功"})
}

func (ctrl *ShopProductCtrl) UpdatePrice(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.ShopProductPriceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.UpdatePrice(db, tenantID, id, userID, &req); err != nil {
		handleShopProductError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *ShopProductCtrl) UpdateStatus(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.ShopProductStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.UpdateStatus(db, tenantID, id, userID, &req); err != nil {
		handleShopProductError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *ShopProductCtrl) DeleteSelection(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.DeleteSelection(db, tenantID, id); err != nil {
		handleShopProductError(c, err)
		return
	}
	response.OKMsg(c, "删除成功")
}

func handleShopProductError(c *gin.Context, err error) {
	switch err {
	case shared.ErrShopProductNotFound:
		response.NotFound(c, err.Error())
	case shared.ErrShopProductAlreadySelected, shared.ErrShopProductNotOnShelf:
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}


