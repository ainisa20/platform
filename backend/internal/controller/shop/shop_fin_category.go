package shop

import (
	"platform/internal/model/dto"
	"platform/internal/pkg/response"
	shopsvc "platform/internal/service/shop"
	"platform/internal/service/shared"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ShopFinCategoryCtrl struct {
	svc *shopsvc.ShopFinCategoryService
}

func NewShopFinCategoryCtrl(svc *shopsvc.ShopFinCategoryService) *ShopFinCategoryCtrl {
	return &ShopFinCategoryCtrl{svc: svc}
}

func (ctrl *ShopFinCategoryCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.ShopFinCategoryListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tree, err := ctrl.svc.ListSynced(db, tenantID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, tree)
}

func (ctrl *ShopFinCategoryCtrl) Available(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	tree, err := ctrl.svc.ListAvailable(db)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, tree)
}

func (ctrl *ShopFinCategoryCtrl) Sync(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")

	var req dto.ShopFinCategorySyncReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Sync(db, tenantID, userID, &req); err != nil {
		handleShopFinCategoryError(c, err)
		return
	}
	response.OKMsg(c, "同步成功")
}

func (ctrl *ShopFinCategoryCtrl) CancelSync(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.CancelSync(db, tenantID, id, userID); err != nil {
		handleShopFinCategoryError(c, err)
		return
	}
	response.OKMsg(c, "取消同步成功")
}

func handleShopFinCategoryError(c *gin.Context, err error) {
	switch err {
	case shared.ErrShopFinCategoryNotFound:
		response.NotFound(c, err.Error())
	case shared.ErrShopFinCategoryAlreadySynced,
		shared.ErrShopFinCategoryReferenced,
		shared.ErrFinanceCategoryHasChildren:
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}
