package platform

import (
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type WorkflowRepository interface {
	ListByProductID(db *gorm.DB, productID uint64) ([]entity.ProductWorkflowNode, error)
	DeleteByProductID(db *gorm.DB, productID uint64) error
	BatchCreate(db *gorm.DB, nodes []entity.ProductWorkflowNode) error
}

type workflowRepository struct{}

func NewWorkflowRepository() WorkflowRepository {
	return &workflowRepository{}
}

func (r *workflowRepository) ListByProductID(db *gorm.DB, productID uint64) ([]entity.ProductWorkflowNode, error) {
	var nodes []entity.ProductWorkflowNode
	if err := db.Where("product_id = ?", productID).Order("node_index ASC").Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *workflowRepository) DeleteByProductID(db *gorm.DB, productID uint64) error {
	return db.Where("product_id = ?", productID).Delete(&entity.ProductWorkflowNode{}).Error
}

func (r *workflowRepository) BatchCreate(db *gorm.DB, nodes []entity.ProductWorkflowNode) error {
	return db.Create(&nodes).Error
}
