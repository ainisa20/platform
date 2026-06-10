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

type OrderCtrl struct {
	svc *shopsvc.OrderService
}

func NewOrderCtrl(svc *shopsvc.OrderService) *OrderCtrl {
	return &OrderCtrl{svc: svc}
}

func (ctrl *OrderCtrl) List(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.OrderListReq
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

func (ctrl *OrderCtrl) Export(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")

	var req dto.OrderListReq
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

func (ctrl *OrderCtrl) Get(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	order, err := ctrl.svc.Get(c, db, tenantID, id)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.OK(c, order)
}

func (ctrl *OrderCtrl) Create(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")

	var req dto.OrderCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	order, err := ctrl.svc.Create(c, db, tenantID, userID, &req)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.R{Code: 0, Message: "success", Data: order})
}

func (ctrl *OrderCtrl) CancelGroup(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := ctrl.svc.CancelGroup(c, db, tenantID, id, userID); err != nil {
		handleOrderError(c, err)
		return
	}
	response.OKMsg(c, "取消成功")
}

func (ctrl *OrderCtrl) CancelItem(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	itemID, err := parseItemID(c)
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}

	if err := ctrl.svc.CancelItem(c, db, tenantID, id, itemID, userID); err != nil {
		handleOrderError(c, err)
		return
	}
	response.OKMsg(c, "明细已取消")
}

func (ctrl *OrderCtrl) GetItemWorkflow(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	itemID, err := parseItemID(c)
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}

	nodes, currentIndex, err := ctrl.svc.GetItemWorkflow(c, db, tenantID, id, itemID)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.OK(c, gin.H{
		"nodes":              nodes,
		"current_node_index": currentIndex,
	})
}

func (ctrl *OrderCtrl) GetItemWorkflowLogs(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	itemID, err := parseItemID(c)
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}

	logs, err := ctrl.svc.GetItemWorkflowLogs(c, db, tenantID, id, itemID)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.OK(c, logs)
}

func (ctrl *OrderCtrl) AdvanceItemWorkflow(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	itemID, err := parseItemID(c)
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}

	var req dto.OrderWorkflowAdvanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userName := ctrl.lookupUserName(c, db, tenantID, userID)

	logID, err := ctrl.svc.AdvanceItemWorkflow(c, db, tenantID, id, itemID, userID, userName, &req)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.R{Code: 0, Message: "success", Data: gin.H{"workflow_log_id": logID}})
}

func (ctrl *OrderCtrl) ListItemAttachments(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	itemID, err := parseItemID(c)
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}

	atts, err := ctrl.svc.ListItemAttachments(c, db, tenantID, id, itemID)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.OK(c, atts)
}

func (ctrl *OrderCtrl) GetItemAttachment(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	itemID, err := parseItemID(c)
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}
	attID, err := parseAttID(c)
	if err != nil {
		response.BadRequest(c, "invalid attachment id")
		return
	}

	url, err := ctrl.svc.GetItemAttachmentDownloadURL(c, db, tenantID, id, itemID, attID)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	response.OK(c, gin.H{
		"id":  attID,
		"url": url,
	})
}

func (ctrl *OrderCtrl) CreateItemAttachment(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint64("tenant_id")
	userID := c.GetUint64("user_id")
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	itemID, err := parseItemID(c)
	if err != nil {
		response.BadRequest(c, "invalid item id")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	if file.Size > 20<<20 {
		response.BadRequest(c, "file too large (max 20MB)")
		return
	}

	var workflowLogID *uint64
	if raw := c.PostForm("workflow_log_id"); raw != "" {
		parsed, parseErr := strconv.ParseUint(raw, 10, 64)
		if parseErr != nil {
			response.BadRequest(c, "invalid workflow_log_id")
			return
		}
		workflowLogID = &parsed
	}

	att, err := ctrl.svc.CreateItemAttachment(c, db, tenantID, id, itemID, userID, file, workflowLogID)
	if err != nil {
		handleOrderError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.R{Code: 0, Message: "success", Data: att})
}

func parseItemID(c *gin.Context) (uint64, error) {
	s := c.Param("itemId")
	if s == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseUint(s, 10, 64)
}

func parseAttID(c *gin.Context) (uint64, error) {
	s := c.Param("attId")
	if s == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseUint(s, 10, 64)
}

func (ctrl *OrderCtrl) lookupUserName(c *gin.Context, db *gorm.DB, tenantID, userID uint64) string {
	type row struct {
		RealName string
	}
	var r row
	if err := db.Table("sys_user").Select("real_name").Where("id = ? AND tenant_id = ?", userID, tenantID).Scan(&r).Error; err != nil {
		return ""
	}
	return r.RealName
}

func handleOrderError(c *gin.Context, err error) {
	switch err {
	case shared.ErrOrderNotFound, shared.ErrOrderItemNotFound:
		response.NotFound(c, err.Error())
	case shared.ErrOrderInProgress, shared.ErrOrderCompleted,
		shared.ErrOrderItemCannotCancel, shared.ErrOrderNoItems,
		shared.ErrWorkflowEmpty, shared.ErrShopProductNotFound,
		shared.ErrShopProductNotOnShelf, shared.ErrShopCustomerNotFound:
		response.BadRequest(c, err.Error())
	default:
		if err.Error() == "已到达最后节点" {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
	}
}