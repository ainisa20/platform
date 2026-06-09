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

type ProductCtrl struct {
	svc *platformsvc.ProductService
}

func NewProductCtrl(svc *platformsvc.ProductService) *ProductCtrl {
	return &ProductCtrl{svc: svc}
}

func (ctrl *ProductCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req dto.ProductListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	products, total, err := ctrl.svc.List(db, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Page(c, products, total, req.Page, req.PageSize)
}

func (ctrl *ProductCtrl) Get(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	product, err := ctrl.svc.GetByID(db, id)
	if err != nil {
		handleProductError(c, err)
		return
	}
	response.OK(c, product)
}

func (ctrl *ProductCtrl) Create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	createdBy := c.GetUint64("user_id")

	var req dto.ProductCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	product, err := ctrl.svc.Create(db, createdBy, &req)
	if err != nil {
		handleProductError(c, err)
		return
	}
	c.JSON(201, response.R{Code: 0, Message: "创建成功", Data: product})
}

func (ctrl *ProductCtrl) Update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	updatedBy := c.GetUint64("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.ProductUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Update(db, id, updatedBy, &req); err != nil {
		handleProductError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *ProductCtrl) Delete(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.Delete(db, id); err != nil {
		handleProductError(c, err)
		return
	}
	response.OKMsg(c, "删除成功")
}

func (ctrl *ProductCtrl) UpdateStatus(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.ProductStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.UpdateStatus(db, id, req.Status); err != nil {
		handleProductError(c, err)
		return
	}
	response.OKMsg(c, "状态更新成功")
}

func (ctrl *ProductCtrl) GetWorkflow(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	nodes, err := ctrl.svc.GetWorkflow(db, id)
	if err != nil {
		handleProductError(c, err)
		return
	}
	response.OK(c, nodes)
}

func (ctrl *ProductCtrl) SaveWorkflow(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	updatedBy := c.GetUint64("user_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.WorkflowSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.SaveWorkflow(db, id, updatedBy, &req); err != nil {
		handleProductError(c, err)
		return
	}
	response.OKMsg(c, "流程保存成功")
}

func handleProductError(c *gin.Context, err error) {
	switch err {
	case shared.ErrProductNotFound:
		response.NotFound(c, err.Error())
	case shared.ErrProductCodeExists:
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}
