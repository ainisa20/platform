package shop

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

const orderStatusCancelled int16 = 4

type OrderRepository interface {
	ListGroups(db *gorm.DB, tenantID uint64, req *dto.OrderListReq) ([]entity.OrderGroup, int64, error)
	GetGroup(db *gorm.DB, id uint64) (*entity.OrderGroup, error)
	CreateGroup(db *gorm.DB, g *entity.OrderGroup) error
	UpdateGroup(db *gorm.DB, g *entity.OrderGroup) error
	CountGroupsByOrderNoPrefix(db *gorm.DB, tenantID uint64, prefix string) (int64, error)

	ListItemsByGroup(db *gorm.DB, groupID uint64) ([]entity.OrderItem, error)
	GetItem(db *gorm.DB, id uint64) (*entity.OrderItem, error)
	CreateItem(db *gorm.DB, item *entity.OrderItem) error
	UpdateItem(db *gorm.DB, item *entity.OrderItem) error

	CreateItemNodes(db *gorm.DB, nodes []entity.OrderItemNode) error
	ListItemNodes(db *gorm.DB, itemID uint64) ([]entity.OrderItemNode, error)
	ListItemsNodes(db *gorm.DB, itemIDs []uint64) ([]entity.OrderItemNode, error)

	ListWorkflowLogs(db *gorm.DB, itemID uint64) ([]entity.OrderWorkflowLog, error)
	CreateWorkflowLog(db *gorm.DB, log *entity.OrderWorkflowLog) error

	ListAttachments(db *gorm.DB, itemID uint64) ([]entity.OrderAttachment, error)
	GetAttachment(db *gorm.DB, id uint64) (*entity.OrderAttachment, error)
	CreateAttachment(db *gorm.DB, a *entity.OrderAttachment) error
}

type orderRepository struct{}

func NewOrderRepository() OrderRepository {
	return &orderRepository{}
}

func (r *orderRepository) ListGroups(db *gorm.DB, tenantID uint64, req *dto.OrderListReq) ([]entity.OrderGroup, int64, error) {
	var groups []entity.OrderGroup
	var total int64

	q := db.Model(&entity.OrderGroup{}).Where("tenant_id = ?", tenantID)
	if req.OrderNo != "" {
		q = q.Where("order_no LIKE ?", "%"+req.OrderNo+"%")
	}
	if req.CustomerID != nil {
		q = q.Where("customer_id = ?", *req.CustomerID)
	}
	if req.OrderStatus != nil {
		q = q.Where("order_status = ?", *req.OrderStatus)
	}
	if req.ExcludeCancelled {
		q = q.Where("order_status <> ?", orderStatusCancelled)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	if err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&groups).Error; err != nil {
		return nil, 0, err
	}
	return groups, total, nil
}

func (r *orderRepository) GetGroup(db *gorm.DB, id uint64) (*entity.OrderGroup, error) {
	var g entity.OrderGroup
	if err := db.First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *orderRepository) CreateGroup(db *gorm.DB, g *entity.OrderGroup) error {
	return db.Create(g).Error
}

func (r *orderRepository) UpdateGroup(db *gorm.DB, g *entity.OrderGroup) error {
	return db.Save(g).Error
}

func (r *orderRepository) CountGroupsByOrderNoPrefix(db *gorm.DB, tenantID uint64, prefix string) (int64, error) {
	var count int64
	if err := db.Model(&entity.OrderGroup{}).
		Where("tenant_id = ? AND order_no LIKE ?", tenantID, prefix+"%").
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *orderRepository) ListItemsByGroup(db *gorm.DB, groupID uint64) ([]entity.OrderItem, error) {
	var items []entity.OrderItem
	if err := db.Where("order_group_id = ? AND deleted_at IS NULL", groupID).
		Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *orderRepository) GetItem(db *gorm.DB, id uint64) (*entity.OrderItem, error) {
	var item entity.OrderItem
	if err := db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *orderRepository) CreateItem(db *gorm.DB, item *entity.OrderItem) error {
	return db.Create(item).Error
}

func (r *orderRepository) UpdateItem(db *gorm.DB, item *entity.OrderItem) error {
	return db.Save(item).Error
}

func (r *orderRepository) ListWorkflowLogs(db *gorm.DB, itemID uint64) ([]entity.OrderWorkflowLog, error) {
	var logs []entity.OrderWorkflowLog
	if err := db.Where("order_item_id = ?", itemID).
		Order("id ASC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *orderRepository) CreateWorkflowLog(db *gorm.DB, log *entity.OrderWorkflowLog) error {
	return db.Create(log).Error
}

func (r *orderRepository) ListAttachments(db *gorm.DB, itemID uint64) ([]entity.OrderAttachment, error) {
	var atts []entity.OrderAttachment
	if err := db.Where("order_item_id = ? AND deleted_at IS NULL", itemID).
		Order("id DESC").Find(&atts).Error; err != nil {
		return nil, err
	}
	return atts, nil
}

func (r *orderRepository) GetAttachment(db *gorm.DB, id uint64) (*entity.OrderAttachment, error) {
	var a entity.OrderAttachment
	if err := db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *orderRepository) CreateAttachment(db *gorm.DB, a *entity.OrderAttachment) error {
	return db.Create(a).Error
}

func (r *orderRepository) CreateItemNodes(db *gorm.DB, nodes []entity.OrderItemNode) error {
	return db.Create(&nodes).Error
}

func (r *orderRepository) ListItemNodes(db *gorm.DB, itemID uint64) ([]entity.OrderItemNode, error) {
	var nodes []entity.OrderItemNode
	err := db.Where("order_item_id = ?", itemID).Order("node_index ASC").Find(&nodes).Error
	return nodes, err
}

func (r *orderRepository) ListItemsNodes(db *gorm.DB, itemIDs []uint64) ([]entity.OrderItemNode, error) {
	var nodes []entity.OrderItemNode
	err := db.Where("order_item_id IN ?", itemIDs).Order("node_index ASC").Find(&nodes).Error
	return nodes, err
}