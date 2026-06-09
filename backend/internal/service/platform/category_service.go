package platform

import (
	"errors"

	"gorm.io/gorm"

	"platform/internal/model/dto"
	"platform/internal/model/entity"
	"platform/internal/repository/platform"
	"platform/internal/service/shared"
)

type CategoryService struct {
	catRepo platform.CategoryRepository
}

func NewCategoryService(catRepo platform.CategoryRepository) *CategoryService {
	return &CategoryService{catRepo: catRepo}
}

func (s *CategoryService) List(db *gorm.DB, req *dto.CategoryListReq) ([]dto.CategoryResp, int64, error) {
	cats, total, err := s.catRepo.List(db, req)
	if err != nil {
		return nil, 0, err
	}
	resps := make([]dto.CategoryResp, 0, len(cats))
	for i := range cats {
		resps = append(resps, s.catToResp(&cats[i]))
	}
	return resps, total, nil
}

func (s *CategoryService) Create(db *gorm.DB, createdBy uint64, req *dto.CategoryCreateReq) (*dto.CategoryResp, error) {
	if existing, err := s.catRepo.GetByName(db, req.CategoryName); err == nil && existing != nil {
		return nil, shared.ErrCategoryNameExists
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	cat := &entity.ProductCategory{
		CategoryCode: req.CategoryCode,
		CategoryName: req.CategoryName,
		Sort:         req.Sort,
		Status:       status,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
	}

	if err := s.catRepo.Create(db, cat); err != nil {
		return nil, err
	}

	refreshed, err := s.catRepo.GetByID(db, cat.ID)
	if err != nil {
		return nil, err
	}
	resp := s.catToResp(refreshed)
	return &resp, nil
}

func (s *CategoryService) Update(db *gorm.DB, id, updatedBy uint64, req *dto.CategoryUpdateReq) error {
	cat, err := s.catRepo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrCategoryNotFound
		}
		return err
	}

	if cat.CategoryName != req.CategoryName {
		if existing, err := s.catRepo.GetByName(db, req.CategoryName); err == nil && existing != nil && existing.ID != id {
			return shared.ErrCategoryNameExists
		}
	}

	cat.CategoryCode = req.CategoryCode
	cat.CategoryName = req.CategoryName
	cat.Sort = req.Sort
	cat.Status = req.Status
	cat.UpdatedBy = updatedBy
	return s.catRepo.Update(db, cat)
}

func (s *CategoryService) Delete(db *gorm.DB, id uint64) error {
	if _, err := s.catRepo.GetByID(db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrCategoryNotFound
		}
		return err
	}
	return s.catRepo.Delete(db, id)
}

func (s *CategoryService) catToResp(c *entity.ProductCategory) dto.CategoryResp {
	return dto.CategoryResp{
		ID:           c.ID,
		CategoryCode: c.CategoryCode,
		CategoryName: c.CategoryName,
		Sort:         c.Sort,
		Status:       c.Status,
		CreatedAt:    c.CreatedAt,
		CreatedBy:    c.CreatedBy,
		UpdatedAt:    c.UpdatedAt,
		UpdatedBy:    c.UpdatedBy,
	}
}
