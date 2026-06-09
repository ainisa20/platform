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

type ProductCategoryCtrl struct {
	svc *platformsvc.CategoryService
}

func NewProductCategoryCtrl(svc *platformsvc.CategoryService) *ProductCategoryCtrl {
	return &ProductCategoryCtrl{svc: svc}
}

func (ctrl *ProductCategoryCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req dto.CategoryListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cats, total, err := ctrl.svc.List(db, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, cats, total, req.Page, req.PageSize)
}

func (ctrl *ProductCategoryCtrl) Create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	createdBy := c.GetUint64("user_id")

	var req dto.CategoryCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cat, err := ctrl.svc.Create(db, createdBy, &req)
	if err != nil {
		handleCategoryError(c, err)
		return
	}
	c.JSON(201, response.R{Code: 0, Message: "创建成功", Data: cat})
}

func (ctrl *ProductCategoryCtrl) Update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	updatedBy := c.GetUint64("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.CategoryUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Update(db, id, updatedBy, &req); err != nil {
		handleCategoryError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *ProductCategoryCtrl) Delete(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.Delete(db, id); err != nil {
		handleCategoryError(c, err)
		return
	}
	response.OKMsg(c, "删除成功")
}

func handleCategoryError(c *gin.Context, err error) {
	switch err {
	case shared.ErrCategoryNotFound:
		response.NotFound(c, err.Error())
	case shared.ErrCategoryNameExists:
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}
