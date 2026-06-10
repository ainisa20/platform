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

type ShopFinAccountCtrl struct {
	svc *shopsvc.ShopFinAccountService
}

func NewShopFinAccountCtrl(svc *shopsvc.ShopFinAccountService) *ShopFinAccountCtrl {
	return &ShopFinAccountCtrl{svc: svc}
}

func (ctrl *ShopFinAccountCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.ShopFinAccountListReq
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

func (ctrl *ShopFinAccountCtrl) Create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")

	var req dto.ShopFinAccountCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	account, err := ctrl.svc.Create(db, tenantID, userID, &req)
	if err != nil {
		handleAccountError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.R{Code: 0, Message: "success", Data: account})
}

func (ctrl *ShopFinAccountCtrl) Update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.ShopFinAccountUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Update(db, tenantID, id, userID, &req); err != nil {
		handleAccountError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *ShopFinAccountCtrl) Delete(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.Delete(db, tenantID, id); err != nil {
		handleAccountError(c, err)
		return
	}
	response.OKMsg(c, "删除成功")
}

func handleAccountError(c *gin.Context, err error) {
	switch err {
	case shared.ErrShopFinAccountNotFound:
		response.NotFound(c, err.Error())
	case shared.ErrShopFinAccountNameExists:
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}
