package shop

import (
	"net/http"
	"strconv"

	"platform/internal/model/dto"
	"platform/internal/pkg/response"
	shopsvc "platform/internal/service/shop"
	"platform/internal/service/shared"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SysUserCtrl struct {
	svc *shopsvc.UserService
}

func NewSysUserCtrl(svc *shopsvc.UserService) *SysUserCtrl {
	return &SysUserCtrl{svc: svc}
}

func (ctrl *SysUserCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.UserListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	users, total, err := ctrl.svc.List(c, db, tenantID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, users, total, req.Page, req.PageSize)
}

func (ctrl *SysUserCtrl) Get(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	user, err := ctrl.svc.GetByID(db, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, user)
}

func (ctrl *SysUserCtrl) Create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	deptID := c.GetUint64("dept_id")
	var dataScope int16
	if v, ok := c.Get("data_scope"); ok {
		dataScope = v.(int16)
	}

	var req dto.UserCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := ctrl.svc.Create(db, tenantID, userID, &req, dataScope, deptID, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.R{Code: 0, Message: "success", Data: user})
}

func (ctrl *SysUserCtrl) Update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.GetUint64("user_id")
	tenantID := c.GetUint64("tenant_id")
	deptID := c.GetUint64("dept_id")
	var dataScope int16
	if v, ok := c.Get("data_scope"); ok {
		dataScope = v.(int16)
	}
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.UserUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Update(db, id, userID, &req, dataScope, deptID, userID, tenantID); err != nil {
		handleError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *SysUserCtrl) Delete(c *gin.Context) {
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

func (ctrl *SysUserCtrl) AssignRoles(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req struct {
		RoleIDs []uint64 `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.AssignRoles(db, id, tenantID, req.RoleIDs); err != nil {
		handleError(c, err)
		return
	}
	response.OKMsg(c, "分配成功")
}

func (ctrl *SysUserCtrl) ResetPassword(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.PasswordResetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.ResetPassword(db, id, tenantID, &req); err != nil {
		handleError(c, err)
		return
	}
	response.OKMsg(c, "密码重置成功")
}

func parseID(c *gin.Context) (uint64, error) {
	s := c.Param("id")
	if s == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseUint(s, 10, 64)
}

func handleError(c *gin.Context, err error) {
	switch err {
	case shared.ErrUserNotFound, shared.ErrRoleNotFound, shared.ErrDeptNotFound:
		response.NotFound(c, err.Error())
	case shared.ErrUsernameExists, shared.ErrRoleCodeExists,
		shared.ErrDeptHasUsers, shared.ErrInvalidPermission,
		shared.ErrSameDeptParent, shared.ErrPermissionExceeded,
		shared.ErrDeptOutOfScope, shared.ErrRoleNotAssignable:
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}
