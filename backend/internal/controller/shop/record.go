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

type RecordCtrl struct {
	svc *shopsvc.RecordService
}

func NewRecordCtrl(svc *shopsvc.RecordService) *RecordCtrl {
	return &RecordCtrl{svc: svc}
}

func (ctrl *RecordCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.FinanceRecordListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	list, total, err := ctrl.svc.List(c, db, tenantID, &req)
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

func (ctrl *RecordCtrl) Get(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	rec, err := ctrl.svc.Get(c, db, id, tenantID)
	if err != nil {
		handleRecordError(c, err)
		return
	}
	response.OK(c, rec)
}

func (ctrl *RecordCtrl) Create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")

	var req dto.FinanceRecordCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	rec, err := ctrl.svc.Create(c, db, tenantID, userID, &req)
	if err != nil {
		handleRecordError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.R{Code: 0, Message: "success", Data: rec})
}

func (ctrl *RecordCtrl) Update(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.FinanceRecordUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Update(c, db, tenantID, id, userID, &req); err != nil {
		handleRecordError(c, err)
		return
	}
	response.OKMsg(c, "更新成功")
}

func (ctrl *RecordCtrl) Delete(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.Delete(c, db, tenantID, id); err != nil {
		handleRecordError(c, err)
		return
	}
	response.OKMsg(c, "删除成功")
}

func (ctrl *RecordCtrl) Review(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req dto.FinanceReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.svc.Review(c, db, tenantID, id, userID, &req); err != nil {
		handleRecordError(c, err)
		return
	}
	response.OKMsg(c, "审核完成")
}

func (ctrl *RecordCtrl) ListAttachments(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	atts, err := ctrl.svc.ListAttachments(c, db, tenantID, id)
	if err != nil {
		handleRecordError(c, err)
		return
	}
	response.OK(c, atts)
}

func (ctrl *RecordCtrl) CreateAttachment(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	if file.Size > 10<<20 {
		response.BadRequest(c, "file too large (max 10MB)")
		return
	}

	att, err := ctrl.svc.CreateAttachment(c, db, tenantID, id, userID, file)
	if err != nil {
		handleRecordError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.R{Code: 0, Message: "success", Data: att})
}

func (ctrl *RecordCtrl) DownloadAttachment(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	recordID, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid record id")
		return
	}
	attID, err := strconv.ParseUint(c.Param("attId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid attachment id")
		return
	}

	url, err := ctrl.svc.GetAttachmentDownloadURL(c, db, tenantID, recordID, attID)
	if err != nil {
		handleRecordError(c, err)
		return
	}
	response.OK(c, map[string]string{"url": url})
}

func (ctrl *RecordCtrl) Export(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.FinanceRecordListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	list, err := ctrl.svc.Export(c, db, tenantID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func handleRecordError(c *gin.Context, err error) {
	switch err {
	case shared.ErrFinRecordNotFound:
		response.NotFound(c, err.Error())
	case shared.ErrFinRecordAccountInvalid,
		shared.ErrFinRecordCategoryInvalid,
		shared.ErrFinRecordOrderInvalid,
		shared.ErrFinRecordApproved,
		shared.ErrFinRecordReviewerIsCreator,
		shared.ErrFinRecordActualAmountRequired,
		shared.ErrFinRecordInvalidAction,
		shared.ErrFinRecordInvalidStatus:
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}