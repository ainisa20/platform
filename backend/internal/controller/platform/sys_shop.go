package platform

import (
	"strconv"

	"platform/internal/model/dto"
	"platform/internal/pkg/response"
	platformsvc "platform/internal/service/platform"
	"platform/internal/service/shared"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SysShopCtrl struct {
	svc *platformsvc.ShopService
}

func NewSysShopCtrl(svc *platformsvc.ShopService) *SysShopCtrl {
	return &SysShopCtrl{svc: svc}
}

func (ctrl *SysShopCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req dto.ShopListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	shops, total, err := ctrl.svc.List(db, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, shops, total, req.Page, req.PageSize)
}

func (ctrl *SysShopCtrl) Get(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id, err := parseShopID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	shop, err := ctrl.svc.GetByID(db, id)
	if err != nil {
		handleShopError(c, err)
		return
	}
	response.OK(c, shop)
}

func (ctrl *SysShopCtrl) Create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	createdBy := c.GetUint64("user_id")

	var req dto.ShopCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	shop, err := ctrl.svc.Create(db, createdBy, &req)
	if err != nil {
		handleShopError(c, err)
		return
	}
	c.JSON(201, response.R{Code: 0, Message: "创建成功", Data: shop})
}

func (ctrl *SysShopCtrl) Update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	updatedBy := c.GetUint64("user_id")
	id, err := parseShopID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.ShopUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Update(db, id, updatedBy, &req); err != nil {
		handleShopError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *SysShopCtrl) Delete(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id, err := parseShopID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.Delete(db, id); err != nil {
		handleShopError(c, err)
		return
	}
	response.OKMsg(c, "删除成功")
}

func (ctrl *SysShopCtrl) UpdateStatus(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id, err := parseShopID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.ShopStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.UpdateStatus(db, id, req.Status); err != nil {
		handleShopError(c, err)
		return
	}
	response.OKMsg(c, "状态更新成功")
}

func (ctrl *SysShopCtrl) ResetAdminPassword(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id, err := parseShopID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.PasswordResetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.ResetAdminPassword(db, id, req.NewPassword); err != nil {
		handleShopError(c, err)
		return
	}
	response.OKMsg(c, "密码重置成功")
}

func parseShopID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}

func handleShopError(c *gin.Context, err error) {
	switch err {
	case shared.ErrShopNotFound:
		response.NotFound(c, err.Error())
	case shared.ErrShopCodeExists, shared.ErrUsernameExists, shared.ErrShopHasUsers:
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}
