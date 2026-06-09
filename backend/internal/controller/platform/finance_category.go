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

type FinanceCategoryCtrl struct {
	svc *platformsvc.FinanceCategoryService
}

func NewFinanceCategoryCtrl(svc *platformsvc.FinanceCategoryService) *FinanceCategoryCtrl {
	return &FinanceCategoryCtrl{svc: svc}
}

func (ctrl *FinanceCategoryCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req dto.FinanceCategoryListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tree, err := ctrl.svc.List(db, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, tree)
}

func (ctrl *FinanceCategoryCtrl) Create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	createdBy := c.GetUint64("user_id")

	var req dto.FinanceCategoryCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cat, err := ctrl.svc.Create(db, createdBy, &req)
	if err != nil {
		handleFinanceCategoryError(c, err)
		return
	}
	c.JSON(201, response.R{Code: 0, Message: "创建成功", Data: cat})
}

func (ctrl *FinanceCategoryCtrl) Update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	updatedBy := c.GetUint64("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.FinanceCategoryUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Update(db, id, updatedBy, &req); err != nil {
		handleFinanceCategoryError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *FinanceCategoryCtrl) Delete(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.Delete(db, id); err != nil {
		handleFinanceCategoryError(c, err)
		return
	}
	response.OKMsg(c, "删除成功")
}

func handleFinanceCategoryError(c *gin.Context, err error) {
	switch err {
	case shared.ErrFinanceCategoryNotFound:
		response.NotFound(c, err.Error())
	case shared.ErrFinanceCategoryNameExists:
		response.BadRequest(c, err.Error())
	case shared.ErrFinanceCategoryHasChildren:
		response.BadRequest(c, err.Error())
	case shared.ErrFinanceCategorySynced:
		response.BadRequest(c, err.Error())
	case shared.ErrFinanceCategoryMaxLevel:
		response.BadRequest(c, err.Error())
	case shared.ErrFinanceCategoryTypeMismatch:
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}
