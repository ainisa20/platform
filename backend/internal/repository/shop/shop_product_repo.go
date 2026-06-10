package shop

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type ShopProductRepository interface {
	List(db *gorm.DB, tenantID uint64, req *dto.ShopProductListReq) ([]entity.ShopProduct, int64, error)
	GetByID(db *gorm.DB, id uint64) (*entity.ShopProduct, error)
	GetByIDInTenant(db *gorm.DB, id, tenantID uint64) (*entity.ShopProduct, error)
	FindByPlatformID(db *gorm.DB, tenantID uint64, platformID uint64) (*entity.ShopProduct, error)
	FindByPlatformIDs(db *gorm.DB, tenantID uint64, platformIDs []uint64) ([]entity.ShopProduct, error)
	Create(db *gorm.DB, sp *entity.ShopProduct) error
	Update(db *gorm.DB, sp *entity.ShopProduct) error
	Delete(db *gorm.DB, id, tenantID uint64) error
}

type shopProductRepository struct{}

func NewShopProductRepository() ShopProductRepository {
	return &shopProductRepository{}
}

func (r *shopProductRepository) List(db *gorm.DB, tenantID uint64, req *dto.ShopProductListReq) ([]entity.ShopProduct, int64, error) {
	var products []entity.ShopProduct
	var total int64

	q := db.Model(&entity.ShopProduct{}).Where("tenant_id = ?", tenantID)
	if req.ProductName != "" {
		q = q.Where("product_name LIKE ?", "%"+req.ProductName+"%")
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

	if err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *shopProductRepository) GetByID(db *gorm.DB, id uint64) (*entity.ShopProduct, error) {
	var sp entity.ShopProduct
	if err := db.First(&sp, id).Error; err != nil {
		return nil, err
	}
	return &sp, nil
}

func (r *shopProductRepository) GetByIDInTenant(db *gorm.DB, id, tenantID uint64) (*entity.ShopProduct, error) {
	var sp entity.ShopProduct
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&sp).Error; err != nil {
		return nil, err
	}
	return &sp, nil
}

func (r *shopProductRepository) FindByPlatformID(db *gorm.DB, tenantID uint64, platformID uint64) (*entity.ShopProduct, error) {
	var sp entity.ShopProduct
	if err := db.Where("tenant_id = ? AND platform_product_id = ?", tenantID, platformID).First(&sp).Error; err != nil {
		return nil, err
	}
	return &sp, nil
}

func (r *shopProductRepository) FindByPlatformIDs(db *gorm.DB, tenantID uint64, platformIDs []uint64) ([]entity.ShopProduct, error) {
	if len(platformIDs) == 0 {
		return nil, nil
	}
	var products []entity.ShopProduct
	if err := db.Where("tenant_id = ? AND platform_product_id IN ?", tenantID, platformIDs).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *shopProductRepository) Create(db *gorm.DB, sp *entity.ShopProduct) error {
	return db.Create(sp).Error
}

func (r *shopProductRepository) Update(db *gorm.DB, sp *entity.ShopProduct) error {
	return db.Save(sp).Error
}

func (r *shopProductRepository) Delete(db *gorm.DB, id, tenantID uint64) error {
	return db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.ShopProduct{}).Error
}
