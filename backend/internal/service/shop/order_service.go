package shop

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"platform/internal/middleware"
	"platform/internal/model/dto"
	"platform/internal/model/entity"
	"platform/internal/pkg/storage"
	"platform/internal/repository/platform"
	"platform/internal/repository/shop"
	"platform/internal/service/shared"
)

const (
	orderItemStatusPending   int16 = 1
	orderItemStatusInProcess int16 = 2
	orderItemStatusCompleted int16 = 3
	orderItemStatusCancelled int16 = 4

	orderStatusPending   int16 = 1
	orderStatusInProcess int16 = 2
	orderStatusCompleted int16 = 3
	orderStatusCancelled int16 = 4
)

type OrderService struct {
	repo         shop.OrderRepository
	customerRepo shop.ShopCustomerRepository
	productRepo  shop.ShopProductRepository
	platProdRepo platform.ProductRepository
	wfRepo       platform.WorkflowRepository
	userRepo     shop.UserRepository
	storage      *storage.MinIOStorage
}

func NewOrderService(
	repo shop.OrderRepository,
	customerRepo shop.ShopCustomerRepository,
	productRepo shop.ShopProductRepository,
	platProdRepo platform.ProductRepository,
	wfRepo platform.WorkflowRepository,
	userRepo shop.UserRepository,
	minioStorage *storage.MinIOStorage,
) *OrderService {
	return &OrderService{
		repo:         repo,
		customerRepo: customerRepo,
		productRepo:  productRepo,
		platProdRepo: platProdRepo,
		wfRepo:       wfRepo,
		userRepo:     userRepo,
		storage:      minioStorage,
	}
}

func (s *OrderService) List(c *gin.Context, db *gorm.DB, tenantID uint64, req *dto.OrderListReq) ([]dto.OrderResp, int64, error) {
	groups, total, err := s.repo.ListGroups(db, tenantID, req)
	if err != nil {
		return nil, 0, err
	}
	if len(groups) == 0 {
		return []dto.OrderResp{}, total, nil
	}
	nameMap := s.fetchUserNames(db, collectOrderCreatedBy(groups))
	list := make([]dto.OrderResp, 0, len(groups))
	for i := range groups {
		g := &groups[i]
		var itemCount int
		if items, err := s.repo.ListItemsByGroup(db, g.ID); err == nil {
			itemCount = len(items)
		}
		resp := dto.OrderResp{
			ID:            g.ID,
			OrderNo:       g.OrderNo,
			CustomerID:    g.CustomerID,
			CustomerName:  g.CustomerName,
			TotalAmount:   g.TotalAmount,
			OrderStatus:   g.OrderStatus,
			Remark:        g.Remark,
			ItemCount:     itemCount,
			CreatedAt:     g.CreatedAt,
			CreatedBy:     g.CreatedBy,
			CreatedByName: nameMap[g.CreatedBy],
			UpdatedAt:     g.UpdatedAt,
		}
		list = append(list, resp)
	}
	return list, total, nil
}

func (s *OrderService) Get(c *gin.Context, db *gorm.DB, tenantID, id uint64) (*dto.OrderResp, error) {
	q := db.Model(&entity.OrderGroup{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var group entity.OrderGroup
	if err := q.First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrOrderNotFound
		}
		return nil, err
	}
	items, err := s.repo.ListItemsByGroup(db, group.ID)
	if err != nil {
		return nil, err
	}
	itemIDs := make([]uint64, 0, len(items))
	for _, it := range items {
		itemIDs = append(itemIDs, it.ID)
	}
	allNodes, _ := s.repo.ListItemsNodes(db, itemIDs)
	nodeMap := make(map[uint64][]entity.OrderItemNode)
	for _, n := range allNodes {
		nodeMap[n.OrderItemID] = append(nodeMap[n.OrderItemID], n)
	}
	var fallbackIDs []uint64
	for _, it := range items {
		if len(nodeMap[it.ID]) == 0 {
			fallbackIDs = append(fallbackIDs, it.ShopProductID)
		}
	}
	var fallbackMap map[uint64][]entity.ProductWorkflowNode
	if len(fallbackIDs) > 0 {
		fallbackMap = s.loadWorkflowByShopProductIDs(db, fallbackIDs)
	}
	itemResps := make([]dto.OrderItemResp, 0, len(items))
	for i := range items {
		it := &items[i]
		nodes := nodeMap[it.ID]
		currentName := ""
		nextName := ""
		if len(nodes) > 0 {
			for _, n := range nodes {
				if n.NodeIndex == it.CurrentNodeIndex {
					currentName = n.NodeName
				}
				if n.NodeIndex == it.CurrentNodeIndex+1 {
					nextName = n.NodeName
				}
			}
		} else if fallbackMap != nil {
			for _, n := range fallbackMap[it.ShopProductID] {
				if n.NodeIndex == it.CurrentNodeIndex {
					currentName = n.NodeName
				}
				if n.NodeIndex == it.CurrentNodeIndex+1 {
					nextName = n.NodeName
				}
			}
		}
		itemResps = append(itemResps, dto.OrderItemResp{
			ID:               it.ID,
			OrderGroupID:     it.OrderGroupID,
			ShopProductID:    it.ShopProductID,
			ProductName:      it.ProductName,
			Quantity:         it.Quantity,
			UnitPrice:        it.UnitPrice,
			TotalPrice:       it.TotalPrice,
			CurrentNodeIndex: it.CurrentNodeIndex,
			CurrentNodeName:  currentName,
			NextNodeName:     nextName,
			ItemStatus:       it.ItemStatus,
			CreatedAt:        it.CreatedAt,
		})
	}
	resp := &dto.OrderResp{
		ID:           group.ID,
		OrderNo:      group.OrderNo,
		CustomerID:   group.CustomerID,
		CustomerName: group.CustomerName,
		TotalAmount:  group.TotalAmount,
		OrderStatus:  group.OrderStatus,
		Remark:       group.Remark,
		ItemCount:    len(items),
		Items:        itemResps,
		CreatedAt:    group.CreatedAt,
		CreatedBy:    group.CreatedBy,
		UpdatedAt:    group.UpdatedAt,
	}
	if group.CreatedBy != 0 {
		nameMap := s.fetchUserNames(db, []uint64{group.CreatedBy})
		resp.CreatedByName = nameMap[group.CreatedBy]
	}
	return resp, nil
}

func (s *OrderService) Create(c *gin.Context, db *gorm.DB, tenantID, createdBy uint64, req *dto.OrderCreateReq) (*dto.OrderResp, error) {
	_ = c
	if len(req.Items) == 0 {
		return nil, shared.ErrOrderNoItems
	}
	customer, err := s.customerRepo.GetByID(db, req.CustomerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrShopCustomerNotFound
		}
		return nil, err
	}
	if customer.TenantID != tenantID {
		return nil, shared.ErrShopCustomerNotFound
	}
	type productMeta struct {
		ShopProductID uint64
		ShopPrice     float64
		ProductName   string
		WFNodes       []entity.ProductWorkflowNode
	}
	metas := make([]productMeta, 0, len(req.Items))
	for _, it := range req.Items {
		sp, err := s.productRepo.GetByID(db, it.ShopProductID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, shared.ErrShopProductNotFound
			}
			return nil, err
		}
		if sp.TenantID != tenantID {
			return nil, shared.ErrShopProductNotFound
		}
		if sp.Status != 1 {
			return nil, shared.ErrShopProductNotOnShelf
		}
		nodes, err := s.wfRepo.ListByProductID(db, sp.PlatformProductID)
		if err != nil {
			return nil, err
		}
		if len(nodes) == 0 {
			return nil, shared.ErrWorkflowEmpty
		}
		metas = append(metas, productMeta{
			ShopProductID: sp.ID,
			ShopPrice:     sp.ShopPrice,
			ProductName:   sp.ProductName,
			WFNodes:       nodes,
		})
	}
	var createdGroup *entity.OrderGroup
	err = db.Transaction(func(tx *gorm.DB) error {
		orderNo := s.generateOrderNo(tx, tenantID)
		var totalAmount float64
		for i, it := range req.Items {
			total := metas[i].ShopPrice * float64(it.Quantity)
			totalAmount += total
		}
		group := &entity.OrderGroup{
			TenantID:     tenantID,
			OrderNo:      orderNo,
			CustomerID:   customer.ID,
			CustomerName: customer.CustomerName,
			TotalAmount:  round2(totalAmount),
			OrderStatus:  orderStatusPending,
			Remark:       req.Remark,
			CreatedBy:    createdBy,
			UpdatedBy:    createdBy,
		}
		if err := s.repo.CreateGroup(tx, group); err != nil {
			return fmt.Errorf("create order group: %w", err)
		}
		var allSnapshotNodes []entity.OrderItemNode
		for i, it := range req.Items {
			total := metas[i].ShopPrice * float64(it.Quantity)
			item := &entity.OrderItem{
				TenantID:         tenantID,
				OrderGroupID:     group.ID,
				ShopProductID:    metas[i].ShopProductID,
				ProductName:      metas[i].ProductName,
				Quantity:         it.Quantity,
				UnitPrice:        metas[i].ShopPrice,
				TotalPrice:       round2(total),
				CurrentNodeIndex: 0,
				ItemStatus:       orderItemStatusPending,
				CreatedBy:        createdBy,
				UpdatedBy:        createdBy,
			}
			if err := s.repo.CreateItem(tx, item); err != nil {
				return fmt.Errorf("create order item: %w", err)
			}
			for _, n := range metas[i].WFNodes {
				allSnapshotNodes = append(allSnapshotNodes, entity.OrderItemNode{
					OrderItemID: item.ID,
					NodeIndex:   n.NodeIndex,
					NodeCode:    n.NodeCode,
					NodeName:    n.NodeName,
				})
			}
		}
		if len(allSnapshotNodes) > 0 {
			if err := s.repo.CreateItemNodes(tx, allSnapshotNodes); err != nil {
				return fmt.Errorf("create item nodes snapshot: %w", err)
			}
		}
		createdGroup = group
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Get(c, db, tenantID, createdGroup.ID)
}

func (s *OrderService) CancelGroup(c *gin.Context, db *gorm.DB, tenantID, id, userID uint64) error {
	q := db.Model(&entity.OrderGroup{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var group entity.OrderGroup
	if err := q.First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrOrderNotFound
		}
		return err
	}
	switch group.OrderStatus {
	case orderStatusCancelled:
		return shared.ErrOrderNotFound
	case orderStatusCompleted:
		return shared.ErrOrderCompleted
	case orderStatusInProcess:
		return shared.ErrOrderInProgress
	case orderStatusPending:
		items, err := s.repo.ListItemsByGroup(db, group.ID)
		if err != nil {
			return err
		}
		for i := range items {
			if items[i].ItemStatus != orderItemStatusPending {
				return shared.ErrOrderInProgress
			}
		}
	default:
		return shared.ErrOrderNotFound
	}
	return db.Transaction(func(tx *gorm.DB) error {
		items, err := s.repo.ListItemsByGroup(tx, group.ID)
		if err != nil {
			return err
		}
		for i := range items {
			items[i].ItemStatus = orderItemStatusCancelled
			items[i].UpdatedBy = userID
			if err := s.repo.UpdateItem(tx, &items[i]); err != nil {
				return err
			}
		}
		group.OrderStatus = orderStatusCancelled
		group.UpdatedBy = userID
		return s.repo.UpdateGroup(tx, &group)
	})
}

func (s *OrderService) CancelItem(c *gin.Context, db *gorm.DB, tenantID, id, itemID, userID uint64) error {
	group, err := s.fetchGroupInScope(c, db, tenantID, id)
	if err != nil {
		return err
	}
	item, err := s.repo.GetItem(db, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrOrderItemNotFound
		}
		return err
	}
	if item.OrderGroupID != group.ID {
		return shared.ErrOrderItemNotFound
	}
	switch item.ItemStatus {
	case orderItemStatusCompleted, orderItemStatusCancelled:
		return shared.ErrOrderItemCannotCancel
	}
	return db.Transaction(func(tx *gorm.DB) error {
		item.ItemStatus = orderItemStatusCancelled
		item.UpdatedBy = userID
		if err := s.repo.UpdateItem(tx, item); err != nil {
			return err
		}
		return s.recomputeAndSaveOrderStatus(tx, group.ID, userID)
	})
}

func (s *OrderService) GetItemWorkflow(c *gin.Context, db *gorm.DB, tenantID, id, itemID uint64) ([]dto.WorkflowNodeResp, int16, error) {
	group, err := s.fetchGroupInScope(c, db, tenantID, id)
	if err != nil {
		return nil, 0, err
	}
	item, err := s.fetchItemInGroup(db, itemID, group.ID)
	if err != nil {
		return nil, 0, err
	}
	snapshotNodes, err := s.repo.ListItemNodes(db, item.ID)
	if err != nil {
		return nil, 0, err
	}
	if len(snapshotNodes) > 0 {
		resps := make([]dto.WorkflowNodeResp, 0, len(snapshotNodes))
		for _, n := range snapshotNodes {
			resps = append(resps, dto.WorkflowNodeResp{
				NodeIndex: n.NodeIndex,
				NodeCode:  n.NodeCode,
				NodeName:  n.NodeName,
			})
		}
		return resps, item.CurrentNodeIndex, nil
	}
	sp, err := s.productRepo.GetByID(db, item.ShopProductID)
	if err != nil {
		return nil, 0, err
	}
	nodes, err := s.wfRepo.ListByProductID(db, sp.PlatformProductID)
	if err != nil {
		return nil, 0, err
	}
	resps := make([]dto.WorkflowNodeResp, 0, len(nodes))
	for _, n := range nodes {
		resps = append(resps, dto.WorkflowNodeResp{
			ID:        n.ID,
			ProductID: n.ProductID,
			NodeIndex: n.NodeIndex,
			NodeCode:  n.NodeCode,
			NodeName:  n.NodeName,
			CreatedAt: n.CreatedAt,
			CreatedBy: n.CreatedBy,
			UpdatedAt: n.UpdatedAt,
			UpdatedBy: n.UpdatedBy,
		})
	}
	return resps, item.CurrentNodeIndex, nil
}

func (s *OrderService) GetItemWorkflowLogs(c *gin.Context, db *gorm.DB, tenantID, id, itemID uint64) ([]dto.OrderWorkflowLogResp, error) {
	group, err := s.fetchGroupInScope(c, db, tenantID, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.fetchItemInGroup(db, itemID, group.ID); err != nil {
		return nil, err
	}
	logs, err := s.repo.ListWorkflowLogs(db, itemID)
	if err != nil {
		return nil, err
	}
	resps := make([]dto.OrderWorkflowLogResp, 0, len(logs))
	for _, l := range logs {
		resps = append(resps, dto.OrderWorkflowLogResp{
			ID:           l.ID,
			OrderItemID:  l.OrderItemID,
			NodeIndex:    l.NodeIndex,
			NodeCode:     l.NodeCode,
			NodeName:     l.NodeName,
			Notes:        l.Notes,
			OperatorID:   l.OperatorID,
			OperatorName: l.OperatorName,
			OperatedAt:   l.OperatedAt,
		})
	}
	return resps, nil
}

func (s *OrderService) AdvanceItemWorkflow(c *gin.Context, db *gorm.DB, tenantID, id, itemID, userID uint64, userName string, req *dto.OrderWorkflowAdvanceReq) (uint64, error) {
	group, err := s.fetchGroupInScope(c, db, tenantID, id)
	if err != nil {
		return 0, err
	}
	item, err := s.fetchItemInGroup(db, itemID, group.ID)
	if err != nil {
		return 0, err
	}
	if item.ItemStatus != orderItemStatusPending && item.ItemStatus != orderItemStatusInProcess {
		return 0, shared.ErrOrderItemCannotCancel
	}
	snapshotNodes, err := s.repo.ListItemNodes(db, item.ID)
	if err != nil {
		return 0, err
	}
	type nodeInfo struct {
		NodeIndex int16
		NodeCode  string
		NodeName  string
	}
	var nodes []nodeInfo
	if len(snapshotNodes) > 0 {
		for _, n := range snapshotNodes {
			nodes = append(nodes, nodeInfo{NodeIndex: n.NodeIndex, NodeCode: n.NodeCode, NodeName: n.NodeName})
		}
	} else {
		sp, err := s.productRepo.GetByID(db, item.ShopProductID)
		if err != nil {
			return 0, err
		}
		tplNodes, err := s.wfRepo.ListByProductID(db, sp.PlatformProductID)
		if err != nil {
			return 0, err
		}
		for _, n := range tplNodes {
			nodes = append(nodes, nodeInfo{NodeIndex: n.NodeIndex, NodeCode: n.NodeCode, NodeName: n.NodeName})
		}
	}
	if len(nodes) == 0 {
		return 0, shared.ErrWorkflowEmpty
	}
	currentIndex := item.CurrentNodeIndex
	if int(currentIndex) >= len(nodes)-1 {
		return 0, errors.New("已到达最后节点")
	}
	nextIndex := currentIndex + 1
	nextNode := nodes[nextIndex]
	var logID uint64
	err = db.Transaction(func(tx *gorm.DB) error {
		log := &entity.OrderWorkflowLog{
			TenantID:     tenantID,
			OrderItemID:  item.ID,
			NodeIndex:    nextNode.NodeIndex,
			NodeCode:     nextNode.NodeCode,
			NodeName:     nextNode.NodeName,
			Notes:        req.Notes,
			OperatorID:   userID,
			OperatorName: userName,
		}
		if err := s.repo.CreateWorkflowLog(tx, log); err != nil {
			return fmt.Errorf("create workflow log: %w", err)
		}
		logID = log.ID
		item.CurrentNodeIndex = nextIndex
		if nextIndex == int16(len(nodes)-1) {
			item.ItemStatus = orderItemStatusCompleted
		} else {
			item.ItemStatus = orderItemStatusInProcess
		}
		item.UpdatedBy = userID
		if err := s.repo.UpdateItem(tx, item); err != nil {
			return fmt.Errorf("update item: %w", err)
		}
		return s.recomputeAndSaveOrderStatus(tx, group.ID, userID)
	})
	if err != nil {
		return 0, err
	}
	return logID, nil
}

func (s *OrderService) ListItemAttachments(c *gin.Context, db *gorm.DB, tenantID, id, itemID uint64) ([]dto.OrderAttachmentResp, error) {
	group, err := s.fetchGroupInScope(c, db, tenantID, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.fetchItemInGroup(db, itemID, group.ID); err != nil {
		return nil, err
	}
	atts, err := s.repo.ListAttachments(db, itemID)
	if err != nil {
		return nil, err
	}
	resps := make([]dto.OrderAttachmentResp, 0, len(atts))
	for _, a := range atts {
		resps = append(resps, dto.OrderAttachmentResp{
			ID:            a.ID,
			FileName:      a.FileName,
			FileSize:      a.FileSize,
			FileType:      a.FileType,
			WorkflowLogID: a.WorkflowLogID,
			CreatedAt:     a.CreatedAt,
		})
	}
	return resps, nil
}

func (s *OrderService) GetItemAttachment(c *gin.Context, db *gorm.DB, tenantID, id, itemID, attID uint64) (*entity.OrderAttachment, error) {
	group, err := s.fetchGroupInScope(c, db, tenantID, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.fetchItemInGroup(db, itemID, group.ID); err != nil {
		return nil, err
	}
	a, err := s.repo.GetAttachment(db, attID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrOrderItemNotFound
		}
		return nil, err
	}
	if a.OrderItemID != itemID {
		return nil, shared.ErrOrderItemNotFound
	}
	return a, nil
}

func (s *OrderService) GetItemAttachmentDownloadURL(c *gin.Context, db *gorm.DB, tenantID, orderID, itemID, attID uint64) (string, error) {
	att, err := s.GetItemAttachment(c, db, tenantID, orderID, itemID, attID)
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(att.FilePath, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid file path format")
	}
	bucket, objectName := parts[0], parts[1]
	return s.storage.GetDownloadURL(c.Request.Context(), bucket, objectName, time.Hour)
}

func (s *OrderService) CreateItemAttachment(c *gin.Context, db *gorm.DB, tenantID, id, itemID, userID uint64, fileHeader *multipart.FileHeader, workflowLogID *uint64) (*dto.OrderAttachmentResp, error) {
	group, err := s.fetchGroupInScope(c, db, tenantID, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.fetchItemInGroup(db, itemID, group.ID); err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file for content detection: %w", err)
	}
	contentType := http.DetectContentType(buf[:n])
	if ext := filepath.Ext(fileHeader.Filename); ext != "" {
		if mt := mime.TypeByExtension(ext); mt != "" {
			contentType = mt
		}
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	token := newToken()
	ext := filepath.Ext(fileHeader.Filename)
	objectPath := fmt.Sprintf("order-attachments/%d/%d/%s%s", tenantID, itemID, token, ext)
	bucketName := fmt.Sprintf("tenant-%d", tenantID)

	if err := s.storage.Upload(c.Request.Context(), bucketName, objectPath, file, fileHeader.Size, contentType); err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	a := &entity.OrderAttachment{
		TenantID:      tenantID,
		OrderItemID:   itemID,
		WorkflowLogID: workflowLogID,
		FileName:      fileHeader.Filename,
		FilePath:      fmt.Sprintf("%s/%s", bucketName, objectPath),
		FileSize:      fileHeader.Size,
		FileType:      contentType,
		CreatedBy:     userID,
	}
	if err := s.repo.CreateAttachment(db, a); err != nil {
		return nil, err
	}
	return &dto.OrderAttachmentResp{
		ID:            a.ID,
		FileName:      a.FileName,
		FileSize:      a.FileSize,
		FileType:      a.FileType,
		WorkflowLogID: a.WorkflowLogID,
		CreatedAt:     a.CreatedAt,
	}, nil
}

func (s *OrderService) Export(c *gin.Context, db *gorm.DB, tenantID uint64, req *dto.OrderListReq) ([]dto.OrderResp, error) {
	list, _, err := s.List(c, db, tenantID, req)
	return list, err
}

func (s *OrderService) generateOrderNo(db *gorm.DB, tenantID uint64) string {
	timestamp := time.Now().Format("20060102150405")
	prefix := "ORD" + timestamp
	count, _ := s.repo.CountGroupsByOrderNoPrefix(db, tenantID, prefix)
	seq := count + 1
	return fmt.Sprintf("%s%03d", prefix, seq)
}

func (s *OrderService) fetchGroupInScope(c *gin.Context, db *gorm.DB, tenantID, id uint64) (*entity.OrderGroup, error) {
	q := db.Model(&entity.OrderGroup{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var group entity.OrderGroup
	if err := q.First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrOrderNotFound
		}
		return nil, err
	}
	return &group, nil
}

func (s *OrderService) fetchItemInGroup(db *gorm.DB, itemID uint64, groupID uint64) (*entity.OrderItem, error) {
	item, err := s.repo.GetItem(db, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrOrderItemNotFound
		}
		return nil, err
	}
	if item.OrderGroupID != groupID {
		return nil, shared.ErrOrderItemNotFound
	}
	return item, nil
}

func (s *OrderService) recomputeAndSaveOrderStatus(tx *gorm.DB, groupID uint64, userID uint64) error {
	group, err := s.repo.GetGroup(tx, groupID)
	if err != nil {
		return err
	}
	items, err := s.repo.ListItemsByGroup(tx, groupID)
	if err != nil {
		return err
	}
	group.OrderStatus = recomputeOrderStatus(items)
	group.UpdatedBy = userID
	return s.repo.UpdateGroup(tx, group)
}

func recomputeOrderStatus(items []entity.OrderItem) int16 {
	if len(items) == 0 {
		return orderStatusPending
	}
	allCancelled := true
	allCompleted := true
	allPending := true
	hasInProcess := false
	for _, it := range items {
		if it.ItemStatus != orderItemStatusCancelled {
			allCancelled = false
		}
		if it.ItemStatus != orderItemStatusCompleted {
			allCompleted = false
		}
		if it.ItemStatus != orderItemStatusPending {
			allPending = false
		}
		if it.ItemStatus == orderItemStatusInProcess {
			hasInProcess = true
		}
	}
	if allCancelled {
		return orderStatusCancelled
	}
	if allCompleted {
		return orderStatusCompleted
	}
	if allPending {
		return orderStatusPending
	}
	if hasInProcess {
		return orderStatusInProcess
	}
	return orderStatusInProcess
}

func (s *OrderService) loadWorkflowByShopProductIDs(db *gorm.DB, shopProductIDs []uint64) map[uint64][]entity.ProductWorkflowNode {
	result := make(map[uint64][]entity.ProductWorkflowNode)
	if len(shopProductIDs) == 0 {
		return result
	}
	var products []entity.ShopProduct
	if err := db.Where("id IN ?", shopProductIDs).Find(&products).Error; err != nil {
		return result
	}
	platformIDs := make([]uint64, 0, len(products))
	for _, p := range products {
		platformIDs = append(platformIDs, p.PlatformProductID)
	}
	if len(platformIDs) == 0 {
		return result
	}
	var nodes []entity.ProductWorkflowNode
	if err := db.Where("product_id IN ?", platformIDs).Order("node_index ASC").Find(&nodes).Error; err != nil {
		return result
	}
	byPlatform := make(map[uint64][]entity.ProductWorkflowNode)
	for _, n := range nodes {
		byPlatform[n.ProductID] = append(byPlatform[n.ProductID], n)
	}
	for _, p := range products {
		result[p.ID] = byPlatform[p.PlatformProductID]
	}
	return result
}

func (s *OrderService) fetchUserNames(db *gorm.DB, ids []uint64) map[uint64]string {
	result := make(map[uint64]string, len(ids))
	if len(ids) == 0 {
		return result
	}
	type row struct {
		ID       uint64
		RealName string
	}
	var rows []row
	if err := db.Table("sys_user").
		Select("id, real_name").
		Where("id IN ?", ids).
		Scan(&rows).Error; err != nil {
		return result
	}
	for _, r := range rows {
		result[r.ID] = r.RealName
	}
	return result
}



func collectOrderCreatedBy(groups []entity.OrderGroup) []uint64 {
	idSet := make(map[uint64]struct{}, len(groups))
	ids := make([]uint64, 0, len(groups))
	for _, g := range groups {
		if g.CreatedBy == 0 {
			continue
		}
		if _, ok := idSet[g.CreatedBy]; ok {
			continue
		}
		idSet[g.CreatedBy] = struct{}{}
		ids = append(ids, g.CreatedBy)
	}
	return ids
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func newToken() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}