package shop

import (
	"net/http"

	"platform/internal/model/dto"
	"platform/internal/pkg/response"
	shopsvc "platform/internal/service/shop"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SysRoleCtrl struct {
	svc *shopsvc.RoleService
}

func NewSysRoleCtrl(svc *shopsvc.RoleService) *SysRoleCtrl {
	return &SysRoleCtrl{svc: svc}
}

func (ctrl *SysRoleCtrl) GetByID(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	role, err := ctrl.svc.GetByID(db, id, tenantID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.R{Code: 0, Message: "success", Data: role})
}

func (ctrl *SysRoleCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.RoleListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	roles, total, err := ctrl.svc.List(db, tenantID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, roles, total, req.Page, req.PageSize)
}

func (ctrl *SysRoleCtrl) Create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")

	var req dto.RoleCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	role, err := ctrl.svc.Create(db, tenantID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.R{Code: 0, Message: "success", Data: role})
}

func (ctrl *SysRoleCtrl) Update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.RoleUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Update(db, tenantID, id, userID, &req); err != nil {
		handleError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *SysRoleCtrl) Delete(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.Delete(db, id, tenantID); err != nil {
		handleError(c, err)
		return
	}
	response.OKMsg(c, "删除成功")
}

func (ctrl *SysRoleCtrl) AssignPermissions(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	var dataScope int16
	if v, ok := c.Get("data_scope"); ok {
		dataScope = v.(int16)
	}
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.RoleAssignPermsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.AssignPermissions(db, id, tenantID, &req, userID, dataScope); err != nil {
		handleError(c, err)
		return
	}
	response.OKMsg(c, "分配成功")
}

func (ctrl *SysRoleCtrl) AssignableRoles(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	var dataScope int16
	if v, ok := c.Get("data_scope"); ok {
		dataScope = v.(int16)
	}

	roles, err := ctrl.svc.GetAssignableRoles(db, tenantID, userID, dataScope)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, roles)
}
