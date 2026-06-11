package platform

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"platform/internal/model/dto"
	"platform/internal/model/entity"
	"platform/internal/repository/platform"
	"platform/internal/service/shared"
)

type ProductService struct {
	productRepo  platform.ProductRepository
	workflowRepo platform.WorkflowRepository
}

func NewProductService(
	productRepo platform.ProductRepository,
	workflowRepo platform.WorkflowRepository,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		workflowRepo: workflowRepo,
	}
}

func (s *ProductService) List(db *gorm.DB, req *dto.ProductListReq) ([]dto.ProductResp, int64, error) {
	products, total, err := s.productRepo.List(db, req)
	if err != nil {
		return nil, 0, err
	}
	resps := make([]dto.ProductResp, 0, len(products))
	for i := range products {
		resps = append(resps, s.productToResp(&products[i]))
	}
	return resps, total, nil
}

func (s *ProductService) GetByID(db *gorm.DB, id uint64) (*dto.ProductResp, error) {
	product, err := s.productRepo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrProductNotFound
		}
		return nil, err
	}
	resp := s.productToResp(product)

	nodes, err := s.workflowRepo.ListByProductID(db, id)
	if err != nil {
		return nil, err
	}
	resp.WorkflowNodes = s.nodesToResps(nodes)
	return &resp, nil
}

func (s *ProductService) Create(db *gorm.DB, createdBy uint64, req *dto.ProductCreateReq) (*dto.ProductResp, error) {
	if existing, err := s.productRepo.GetByCode(db, req.ProductCode); err == nil && existing != nil {
		return nil, shared.ErrProductCodeExists
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	product := &entity.Product{
		ProductCode:     req.ProductCode,
		ProductName:     req.ProductName,
		CategoryID:      req.CategoryID,
		Price:           req.Price,
		Sort:            req.Sort,
		Status:          status,
		MallProductCode: req.MallProductCode,
		Description:     req.Description,
		CreatedBy:       createdBy,
		UpdatedBy:       createdBy,
	}

	if err := s.productRepo.Create(db, product); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	nodes := make([]entity.ProductWorkflowNode, len(req.WorkflowNodes))
	for i, n := range req.WorkflowNodes {
		nodes[i] = entity.ProductWorkflowNode{
			ProductID: product.ID,
			NodeIndex: n.NodeIndex,
			NodeCode:  n.NodeCode,
			NodeName:  n.NodeName,
			CreatedBy: createdBy,
			UpdatedBy: createdBy,
		}
	}
	if err := s.workflowRepo.BatchCreate(db, nodes); err != nil {
		return nil, fmt.Errorf("create workflow nodes: %w", err)
	}

	return s.GetByID(db, product.ID)
}

func (s *ProductService) Update(db *gorm.DB, id, updatedBy uint64, req *dto.ProductUpdateReq) error {
	product, err := s.productRepo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrProductNotFound
		}
		return err
	}

	if req.ProductCode != "" {
		if existing, err := s.productRepo.GetByCode(db, req.ProductCode); err == nil && existing != nil && existing.ID != id {
			return shared.ErrProductCodeExists
		}
		product.ProductCode = req.ProductCode
	}

	product.ProductName = req.ProductName
	product.CategoryID = req.CategoryID
	product.Price = req.Price
	product.Sort = req.Sort
	product.Status = req.Status
	product.MallProductCode = req.MallProductCode
	product.Description = req.Description
	product.UpdatedBy = updatedBy
	return s.productRepo.Update(db, product)
}

func (s *ProductService) UpdateStatus(db *gorm.DB, id uint64, status int16) error {
	if _, err := s.productRepo.GetByID(db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrProductNotFound
		}
		return err
	}
	return s.productRepo.UpdateStatus(db, id, status)
}

func (s *ProductService) Delete(db *gorm.DB, id uint64) error {
	if _, err := s.productRepo.GetByID(db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrProductNotFound
		}
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := s.workflowRepo.DeleteByProductID(tx, id); err != nil {
			return fmt.Errorf("delete workflow nodes: %w", err)
		}
		if err := s.productRepo.Delete(tx, id); err != nil {
			return fmt.Errorf("delete product: %w", err)
		}
		return nil
	})
}

func (s *ProductService) GetWorkflow(db *gorm.DB, productID uint64) ([]dto.WorkflowNodeResp, error) {
	if _, err := s.productRepo.GetByID(db, productID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrProductNotFound
		}
		return nil, err
	}
	nodes, err := s.workflowRepo.ListByProductID(db, productID)
	if err != nil {
		return nil, err
	}
	return s.nodesToResps(nodes), nil
}

func (s *ProductService) SaveWorkflow(db *gorm.DB, productID, updatedBy uint64, req *dto.WorkflowSaveReq) error {
	if _, err := s.productRepo.GetByID(db, productID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrProductNotFound
		}
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := s.workflowRepo.DeleteByProductID(tx, productID); err != nil {
			return fmt.Errorf("delete old workflow: %w", err)
		}

		nodes := make([]entity.ProductWorkflowNode, len(req.Nodes))
		for i, n := range req.Nodes {
			nodes[i] = entity.ProductWorkflowNode{
				ProductID: productID,
				NodeIndex: n.NodeIndex,
				NodeCode:  n.NodeCode,
				NodeName:  n.NodeName,
				CreatedBy: updatedBy,
				UpdatedBy: updatedBy,
			}
		}
		if len(nodes) > 0 {
			if err := s.workflowRepo.BatchCreate(tx, nodes); err != nil {
				return fmt.Errorf("create workflow nodes: %w", err)
			}
		}
		return nil
	})
}

func (s *ProductService) productToResp(p *entity.Product) dto.ProductResp {
	return dto.ProductResp{
		ID:              p.ID,
		ProductCode:     p.ProductCode,
		ProductName:     p.ProductName,
		CategoryID:      p.CategoryID,
		Price:           p.Price,
		Sort:            p.Sort,
		Status:          p.Status,
		MallProductCode: p.MallProductCode,
		Description:     p.Description,
		CreatedAt:       p.CreatedAt,
		CreatedBy:       p.CreatedBy,
		UpdatedAt:       p.UpdatedAt,
		UpdatedBy:       p.UpdatedBy,
	}
}

func (s *ProductService) nodesToResps(nodes []entity.ProductWorkflowNode) []dto.WorkflowNodeResp {
	resps := make([]dto.WorkflowNodeResp, 0, len(nodes))
	for i := range nodes {
		resps = append(resps, dto.WorkflowNodeResp{
			ID:        nodes[i].ID,
			ProductID: nodes[i].ProductID,
			NodeIndex: nodes[i].NodeIndex,
			NodeCode:  nodes[i].NodeCode,
			NodeName:  nodes[i].NodeName,
			CreatedAt: nodes[i].CreatedAt,
			CreatedBy: nodes[i].CreatedBy,
			UpdatedAt: nodes[i].UpdatedAt,
			UpdatedBy: nodes[i].UpdatedBy,
		})
	}
	return resps
}
