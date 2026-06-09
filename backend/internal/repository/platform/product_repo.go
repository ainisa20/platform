package platform

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type ProductRepository interface {
	List(db *gorm.DB, req *dto.ProductListReq) ([]entity.Product, int64, error)
	GetByID(db *gorm.DB, id uint64) (*entity.Product, error)
	GetByCode(db *gorm.DB, code string) (*entity.Product, error)
	Create(db *gorm.DB, product *entity.Product) error
	Update(db *gorm.DB, product *entity.Product) error
	UpdateStatus(db *gorm.DB, id uint64, status int16) error
	Delete(db *gorm.DB, id uint64) error
}

type productRepository struct{}

func NewProductRepository() ProductRepository {
	return &productRepository{}
}

func (r *productRepository) List(db *gorm.DB, req *dto.ProductListReq) ([]entity.Product, int64, error) {
	var products []entity.Product
	var total int64

	q := db.Model(&entity.Product{})
	if req.ProductName != "" {
		q = q.Where("product_name LIKE ?", "%"+req.ProductName+"%")
	}
	if req.CategoryID != nil {
		q = q.Where("category_id = ?", *req.CategoryID)
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

	if err := q.Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *productRepository) GetByID(db *gorm.DB, id uint64) (*entity.Product, error) {
	var product entity.Product
	if err := db.First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetByCode(db *gorm.DB, code string) (*entity.Product, error) {
	var product entity.Product
	if err := db.Where("product_code = ?", code).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Create(db *gorm.DB, product *entity.Product) error {
	return db.Create(product).Error
}

func (r *productRepository) Update(db *gorm.DB, product *entity.Product) error {
	return db.Save(product).Error
}

func (r *productRepository) UpdateStatus(db *gorm.DB, id uint64, status int16) error {
	return db.Model(&entity.Product{}).Where("id = ?", id).Update("status", status).Error
}

func (r *productRepository) Delete(db *gorm.DB, id uint64) error {
	return db.Delete(&entity.Product{}, id).Error
}
