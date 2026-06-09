package shop

import (
	"platform/internal/model/enum"
	"platform/internal/pkg/response"
	shopsvc "platform/internal/service/shop"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SysPermissionCtrl struct {
	svc *shopsvc.RoleService
}

func NewSysPermissionCtrl(svc *shopsvc.RoleService) *SysPermissionCtrl {
	return &SysPermissionCtrl{svc: svc}
}

func (ctrl *SysPermissionCtrl) GetTree(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.GetUint64("user_id")
	tenantID := c.GetUint64("tenant_id")
	var dataScope int16
	if v, ok := c.Get("data_scope"); ok {
		dataScope = v.(int16)
	}

	tree, err := ctrl.svc.GetPermissionTree(db, enum.SystemTypeShop, userID, dataScope, tenantID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, tree)
}
