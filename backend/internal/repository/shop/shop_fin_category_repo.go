package shop

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type ShopFinCategoryRepository interface {
	ListSynced(db *gorm.DB, tenantID uint64, req *dto.ShopFinCategoryListReq) ([]entity.ShopFinanceCategory, error)
	FindByPlatformID(db *gorm.DB, tenantID uint64, platformID uint64) (*entity.ShopFinanceCategory, error)
	FindByPlatformIDs(db *gorm.DB, tenantID uint64, platformIDs []uint64) ([]entity.ShopFinanceCategory, error)
	Create(db *gorm.DB, cat *entity.ShopFinanceCategory) error
	Delete(db *gorm.DB, id, tenantID uint64) error
	HasReference(db *gorm.DB, id uint64) (bool, error)
	HasChildren(db *gorm.DB, id uint64) (bool, error)
	GetByID(db *gorm.DB, id uint64) (*entity.ShopFinanceCategory, error)
	GetByIDInTenant(db *gorm.DB, id, tenantID uint64) (*entity.ShopFinanceCategory, error)
}

type shopFinCategoryRepository struct{}

func NewShopFinCategoryRepository() ShopFinCategoryRepository {
	return &shopFinCategoryRepository{}
}

func (r *shopFinCategoryRepository) ListSynced(db *gorm.DB, tenantID uint64, req *dto.ShopFinCategoryListReq) ([]entity.ShopFinanceCategory, error) {
	var cats []entity.ShopFinanceCategory
	q := db.Model(&entity.ShopFinanceCategory{}).Where("tenant_id = ?", tenantID)
	if req.CategoryType != nil {
		q = q.Where("category_type = ?", *req.CategoryType)
	}
	if err := q.Order("id ASC").Find(&cats).Error; err != nil {
		return nil, err
	}
	return cats, nil
}

func (r *shopFinCategoryRepository) FindByPlatformID(db *gorm.DB, tenantID uint64, platformID uint64) (*entity.ShopFinanceCategory, error) {
	var cat entity.ShopFinanceCategory
	if err := db.Where("tenant_id = ? AND platform_category_id = ?", tenantID, platformID).First(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *shopFinCategoryRepository) FindByPlatformIDs(db *gorm.DB, tenantID uint64, platformIDs []uint64) ([]entity.ShopFinanceCategory, error) {
	if len(platformIDs) == 0 {
		return nil, nil
	}
	var cats []entity.ShopFinanceCategory
	if err := db.Where("tenant_id = ? AND platform_category_id IN ?", tenantID, platformIDs).Find(&cats).Error; err != nil {
		return nil, err
	}
	return cats, nil
}

func (r *shopFinCategoryRepository) Create(db *gorm.DB, cat *entity.ShopFinanceCategory) error {
	return db.Create(cat).Error
}

func (r *shopFinCategoryRepository) Delete(db *gorm.DB, id, tenantID uint64) error {
	return db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.ShopFinanceCategory{}).Error
}

func (r *shopFinCategoryRepository) HasReference(db *gorm.DB, id uint64) (bool, error) {
	if !db.Migrator().HasTable("finance_record") {
		return false, nil
	}
	var count int64
	if err := db.Table("finance_record").Where("category_id = ?", id).Count(&count).Error; err != nil {
		return false, nil
	}
	return count > 0, nil
}

func (r *shopFinCategoryRepository) HasChildren(db *gorm.DB, id uint64) (bool, error) {
	var count int64
	if err := db.Model(&entity.ShopFinanceCategory{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *shopFinCategoryRepository) GetByID(db *gorm.DB, id uint64) (*entity.ShopFinanceCategory, error) {
	var cat entity.ShopFinanceCategory
	if err := db.First(&cat, id).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *shopFinCategoryRepository) GetByIDInTenant(db *gorm.DB, id, tenantID uint64) (*entity.ShopFinanceCategory, error) {
	var cat entity.ShopFinanceCategory
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}
