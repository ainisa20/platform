package shop

import (
	"net/http"

	"platform/internal/model/dto"
	"platform/internal/pkg/response"
	shopsvc "platform/internal/service/shop"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SysDeptCtrl struct {
	svc *shopsvc.DeptService
}

func NewSysDeptCtrl(svc *shopsvc.DeptService) *SysDeptCtrl {
	return &SysDeptCtrl{svc: svc}
}

func (ctrl *SysDeptCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	tree, err := ctrl.svc.GetTree(c, db, tenantID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, tree)
}

func (ctrl *SysDeptCtrl) Create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")

	var req dto.DeptCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dept, err := ctrl.svc.Create(db, tenantID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.R{Code: 0, Message: "success", Data: dept})
}

func (ctrl *SysDeptCtrl) Update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.DeptUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Update(db, id, userID, &req); err != nil {
		handleError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *SysDeptCtrl) Delete(c *gin.Context) {
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
