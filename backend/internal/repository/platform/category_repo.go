package platform

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	List(db *gorm.DB, req *dto.CategoryListReq) ([]entity.ProductCategory, int64, error)
	GetByID(db *gorm.DB, id uint64) (*entity.ProductCategory, error)
	GetByName(db *gorm.DB, name string) (*entity.ProductCategory, error)
	Create(db *gorm.DB, cat *entity.ProductCategory) error
	Update(db *gorm.DB, cat *entity.ProductCategory) error
	Delete(db *gorm.DB, id uint64) error
}

type categoryRepository struct{}

func NewCategoryRepository() CategoryRepository {
	return &categoryRepository{}
}

func (r *categoryRepository) List(db *gorm.DB, req *dto.CategoryListReq) ([]entity.ProductCategory, int64, error) {
	var cats []entity.ProductCategory
	var total int64

	q := db.Model(&entity.ProductCategory{})
	if req.CategoryName != "" {
		q = q.Where("category_name LIKE ?", "%"+req.CategoryName+"%")
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

	if err := q.Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&cats).Error; err != nil {
		return nil, 0, err
	}
	return cats, total, nil
}

func (r *categoryRepository) GetByID(db *gorm.DB, id uint64) (*entity.ProductCategory, error) {
	var cat entity.ProductCategory
	if err := db.First(&cat, id).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *categoryRepository) GetByName(db *gorm.DB, name string) (*entity.ProductCategory, error) {
	var cat entity.ProductCategory
	if err := db.Where("category_name = ?", name).First(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *categoryRepository) Create(db *gorm.DB, cat *entity.ProductCategory) error {
	return db.Create(cat).Error
}

func (r *categoryRepository) Update(db *gorm.DB, cat *entity.ProductCategory) error {
	return db.Save(cat).Error
}

func (r *categoryRepository) Delete(db *gorm.DB, id uint64) error {
	return db.Delete(&entity.ProductCategory{}, id).Error
}
