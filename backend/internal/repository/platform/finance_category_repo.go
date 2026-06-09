package platform

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type FinanceCategoryRepository interface {
	List(db *gorm.DB, req *dto.FinanceCategoryListReq) ([]entity.FinanceCategory, error)
	GetByID(db *gorm.DB, id uint64) (*entity.FinanceCategory, error)
	GetByNameAndType(db *gorm.DB, name string, categoryType int16, parentID uint64) (*entity.FinanceCategory, error)
	Create(db *gorm.DB, cat *entity.FinanceCategory) error
	Update(db *gorm.DB, cat *entity.FinanceCategory) error
	Delete(db *gorm.DB, id uint64) error
	HasChildren(db *gorm.DB, id uint64) (bool, error)
	HasShopSync(db *gorm.DB, id uint64) (bool, error)
}

type financeCategoryRepository struct{}

func NewFinanceCategoryRepository() FinanceCategoryRepository {
	return &financeCategoryRepository{}
}

func (r *financeCategoryRepository) List(db *gorm.DB, req *dto.FinanceCategoryListReq) ([]entity.FinanceCategory, error) {
	var cats []entity.FinanceCategory

	q := db.Model(&entity.FinanceCategory{})
	if req.CategoryType != nil {
		q = q.Where("category_type = ?", *req.CategoryType)
	}
	if req.CategoryName != "" {
		q = q.Where("category_name LIKE ?", "%"+req.CategoryName+"%")
	}

	if err := q.Order("sort ASC, id ASC").Find(&cats).Error; err != nil {
		return nil, err
	}
	return cats, nil
}

func (r *financeCategoryRepository) GetByID(db *gorm.DB, id uint64) (*entity.FinanceCategory, error) {
	var cat entity.FinanceCategory
	if err := db.First(&cat, id).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *financeCategoryRepository) GetByNameAndType(db *gorm.DB, name string, categoryType int16, parentID uint64) (*entity.FinanceCategory, error) {
	var cat entity.FinanceCategory
	if err := db.Where("category_name = ? AND category_type = ? AND parent_id = ?", name, categoryType, parentID).First(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *financeCategoryRepository) Create(db *gorm.DB, cat *entity.FinanceCategory) error {
	return db.Create(cat).Error
}

func (r *financeCategoryRepository) Update(db *gorm.DB, cat *entity.FinanceCategory) error {
	return db.Save(cat).Error
}

func (r *financeCategoryRepository) Delete(db *gorm.DB, id uint64) error {
	return db.Delete(&entity.FinanceCategory{}, id).Error
}

func (r *financeCategoryRepository) HasChildren(db *gorm.DB, id uint64) (bool, error) {
	var count int64
	if err := db.Model(&entity.FinanceCategory{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *financeCategoryRepository) HasShopSync(db *gorm.DB, id uint64) (bool, error) {
	var count int64
	if err := db.Table("shop_finance_category").Where("platform_category_id = ?", id).Count(&count).Error; err != nil {
		return false, nil
	}
	return count > 0, nil
}
