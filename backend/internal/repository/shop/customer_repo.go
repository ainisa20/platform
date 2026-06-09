package shop

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type ShopCustomerRepository interface {
	List(db *gorm.DB, tenantID uint64, req *dto.ShopCustomerListReq) ([]entity.ShopCustomer, int64, error)
	GetByID(db *gorm.DB, id uint64) (*entity.ShopCustomer, error)
	CountByNameAndType(db *gorm.DB, tenantID uint64, name string, customerType int16, excludeID uint64) (int64, error)
	Create(db *gorm.DB, c *entity.ShopCustomer) error
	Update(db *gorm.DB, c *entity.ShopCustomer) error
	Delete(db *gorm.DB, id uint64) error
	HasOrders(db *gorm.DB, id uint64) (bool, error)
	ListOrders(db *gorm.DB, tenantID uint64, customerID uint64) ([]dto.ShopCustomerOrderResp, error)
}

type shopCustomerRepository struct{}

func NewShopCustomerRepository() ShopCustomerRepository {
	return &shopCustomerRepository{}
}

func (r *shopCustomerRepository) List(db *gorm.DB, tenantID uint64, req *dto.ShopCustomerListReq) ([]entity.ShopCustomer, int64, error) {
	var customers []entity.ShopCustomer
	var total int64

	q := db.Model(&entity.ShopCustomer{}).Where("tenant_id = ?", tenantID)
	if req.CustomerName != "" {
		q = q.Where("customer_name LIKE ?", "%"+req.CustomerName+"%")
	}
	if req.ContactPerson != "" {
		q = q.Where("contact_person LIKE ?", "%"+req.ContactPerson+"%")
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
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

	if err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&customers).Error; err != nil {
		return nil, 0, err
	}
	return customers, total, nil
}

func (r *shopCustomerRepository) GetByID(db *gorm.DB, id uint64) (*entity.ShopCustomer, error) {
	var c entity.ShopCustomer
	if err := db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *shopCustomerRepository) CountByNameAndType(db *gorm.DB, tenantID uint64, name string, customerType int16, excludeID uint64) (int64, error) {
	var count int64
	q := db.Model(&entity.ShopCustomer{}).
		Where("tenant_id = ? AND customer_name = ? AND customer_type = ? AND id != ? AND deleted_at IS NULL",
			tenantID, name, customerType, excludeID)
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *shopCustomerRepository) Create(db *gorm.DB, c *entity.ShopCustomer) error {
	return db.Create(c).Error
}

func (r *shopCustomerRepository) Update(db *gorm.DB, c *entity.ShopCustomer) error {
	return db.Save(c).Error
}

func (r *shopCustomerRepository) Delete(db *gorm.DB, id uint64) error {
	return db.Delete(&entity.ShopCustomer{}, id).Error
}

func (r *shopCustomerRepository) HasOrders(db *gorm.DB, id uint64) (bool, error) {
	if !db.Migrator().HasTable("order_group") {
		return false, nil
	}
	var count int64
	if err := db.Table("order_group").
		Where("customer_id = ? AND deleted_at IS NULL", id).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *shopCustomerRepository) ListOrders(db *gorm.DB, tenantID uint64, customerID uint64) ([]dto.ShopCustomerOrderResp, error) {
	orders := make([]dto.ShopCustomerOrderResp, 0)
	if !db.Migrator().HasTable("order_group") {
		return orders, nil
	}
	if err := db.Table("order_group").
		Select("id, order_no, total_amount, status, created_at, created_by").
		Where("tenant_id = ? AND customer_id = ? AND deleted_at IS NULL", tenantID, customerID).
		Order("id DESC").
		Scan(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}