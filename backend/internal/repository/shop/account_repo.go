package shop

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type ShopFinAccountRepository interface {
	List(db *gorm.DB, tenantID uint64, req *dto.ShopFinAccountListReq) ([]entity.ShopFinanceAccount, int64, error)
	GetByID(db *gorm.DB, id uint64) (*entity.ShopFinanceAccount, error)
	CountByName(db *gorm.DB, tenantID uint64, name string, excludeID uint64) (int64, error)
	Create(db *gorm.DB, a *entity.ShopFinanceAccount) error
	Update(db *gorm.DB, a *entity.ShopFinanceAccount) error
	Delete(db *gorm.DB, id uint64) error
}

type shopFinAccountRepository struct{}

func NewShopFinAccountRepository() ShopFinAccountRepository {
	return &shopFinAccountRepository{}
}

func (r *shopFinAccountRepository) List(db *gorm.DB, tenantID uint64, req *dto.ShopFinAccountListReq) ([]entity.ShopFinanceAccount, int64, error) {
	var accounts []entity.ShopFinanceAccount
	var total int64

	q := db.Model(&entity.ShopFinanceAccount{}).Where("tenant_id = ?", tenantID)
	if req.AccountName != "" {
		q = q.Where("account_name LIKE ?", "%"+req.AccountName+"%")
	}
	if req.AccountType != nil {
		q = q.Where("account_type = ?", *req.AccountType)
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

	if err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&accounts).Error; err != nil {
		return nil, 0, err
	}
	return accounts, total, nil
}

func (r *shopFinAccountRepository) GetByID(db *gorm.DB, id uint64) (*entity.ShopFinanceAccount, error) {
	var a entity.ShopFinanceAccount
	if err := db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *shopFinAccountRepository) CountByName(db *gorm.DB, tenantID uint64, name string, excludeID uint64) (int64, error) {
	var count int64
	q := db.Model(&entity.ShopFinanceAccount{}).
		Where("tenant_id = ? AND account_name = ? AND id != ? AND deleted_at IS NULL",
			tenantID, name, excludeID)
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *shopFinAccountRepository) Create(db *gorm.DB, a *entity.ShopFinanceAccount) error {
	return db.Create(a).Error
}

func (r *shopFinAccountRepository) Update(db *gorm.DB, a *entity.ShopFinanceAccount) error {
	return db.Save(a).Error
}

func (r *shopFinAccountRepository) Delete(db *gorm.DB, id uint64) error {
	return db.Delete(&entity.ShopFinanceAccount{}, id).Error
}
