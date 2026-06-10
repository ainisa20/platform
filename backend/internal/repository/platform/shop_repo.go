package platform

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type ShopRepository interface {
	List(db *gorm.DB, req *dto.ShopListReq) ([]entity.SysShop, int64, error)
	GetByID(db *gorm.DB, id uint64) (*entity.SysShop, error)
	GetByCode(db *gorm.DB, code string) (*entity.SysShop, error)
	Create(db *gorm.DB, shop *entity.SysShop) error
	Update(db *gorm.DB, shop *entity.SysShop) error
	UpdateStatus(db *gorm.DB, id uint64, status int16) error
	UpdateAdminUserID(db *gorm.DB, id, adminUserID uint64) error
	Delete(db *gorm.DB, id uint64) error
}

type shopRepository struct{}

func NewShopRepository() ShopRepository {
	return &shopRepository{}
}

func (r *shopRepository) List(db *gorm.DB, req *dto.ShopListReq) ([]entity.SysShop, int64, error) {
	var shops []entity.SysShop
	var total int64

	q := db.Model(&entity.SysShop{})
	if req.ShopCode != "" {
		q = q.Where("shop_code LIKE ?", "%"+req.ShopCode+"%")
	}
	if req.ShopName != "" {
		q = q.Where("shop_name LIKE ?", "%"+req.ShopName+"%")
	}
	if req.Province != "" {
		q = q.Where("province = ?", req.Province)
	}
	if req.City != "" {
		q = q.Where("city = ?", req.City)
	}
	if req.District != "" {
		q = q.Where("district = ?", req.District)
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

	if err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&shops).Error; err != nil {
		return nil, 0, err
	}
	return shops, total, nil
}

func (r *shopRepository) GetByID(db *gorm.DB, id uint64) (*entity.SysShop, error) {
	var shop entity.SysShop
	if err := db.First(&shop, id).Error; err != nil {
		return nil, err
	}
	return &shop, nil
}

func (r *shopRepository) GetByCode(db *gorm.DB, code string) (*entity.SysShop, error) {
	var shop entity.SysShop
	if err := db.Where("shop_code = ?", code).First(&shop).Error; err != nil {
		return nil, err
	}
	return &shop, nil
}

func (r *shopRepository) Create(db *gorm.DB, shop *entity.SysShop) error {
	return db.Create(shop).Error
}

func (r *shopRepository) Update(db *gorm.DB, shop *entity.SysShop) error {
	return db.Save(shop).Error
}

func (r *shopRepository) UpdateStatus(db *gorm.DB, id uint64, status int16) error {
	return db.Model(&entity.SysShop{}).Where("id = ?", id).Update("status", status).Error
}

func (r *shopRepository) UpdateAdminUserID(db *gorm.DB, id, adminUserID uint64) error {
	return db.Model(&entity.SysShop{}).Where("id = ?", id).Update("admin_user_id", adminUserID).Error
}

func (r *shopRepository) Delete(db *gorm.DB, id uint64) error {
	return db.Delete(&entity.SysShop{}, id).Error
}
