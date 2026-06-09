package platform

import (
	"errors"

	"gorm.io/gorm"

	"platform/internal/model/dto"
	"platform/internal/model/entity"
	"platform/internal/repository/platform"
	"platform/internal/service/shared"
)

type FinanceCategoryService struct {
	repo platform.FinanceCategoryRepository
}

func NewFinanceCategoryService(repo platform.FinanceCategoryRepository) *FinanceCategoryService {
	return &FinanceCategoryService{repo: repo}
}

func (s *FinanceCategoryService) List(db *gorm.DB, req *dto.FinanceCategoryListReq) ([]dto.FinanceCategoryResp, error) {
	cats, err := s.repo.List(db, req)
	if err != nil {
		return nil, err
	}
	flat := make([]dto.FinanceCategoryResp, 0, len(cats))
	for i := range cats {
		flat = append(flat, s.catToResp(&cats[i]))
	}
	return s.buildTree(flat, 0), nil
}

func (s *FinanceCategoryService) Create(db *gorm.DB, createdBy uint64, req *dto.FinanceCategoryCreateReq) (*dto.FinanceCategoryResp, error) {
	level := int16(1)
	if req.ParentID != 0 {
		parent, err := s.repo.GetByID(db, req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, shared.ErrFinanceCategoryNotFound
			}
			return nil, err
		}
		if parent.CategoryType != req.CategoryType {
			return nil, shared.ErrFinanceCategoryTypeMismatch
		}
		level = parent.Level + 1
	}
	if level > 3 {
		return nil, shared.ErrFinanceCategoryMaxLevel
	}

	if existing, err := s.repo.GetByNameAndType(db, req.CategoryName, req.CategoryType, req.ParentID); err == nil && existing != nil {
		return nil, shared.ErrFinanceCategoryNameExists
	}

	cat := &entity.FinanceCategory{
		ParentID:     req.ParentID,
		Level:        level,
		CategoryType: req.CategoryType,
		CategoryCode: req.CategoryCode,
		CategoryName: req.CategoryName,
		FinanceCode:  req.FinanceCode,
		Sort:         req.Sort,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
	}

	if err := s.repo.Create(db, cat); err != nil {
		return nil, err
	}

	refreshed, err := s.repo.GetByID(db, cat.ID)
	if err != nil {
		return nil, err
	}
	resp := s.catToResp(refreshed)
	return &resp, nil
}

func (s *FinanceCategoryService) Update(db *gorm.DB, id, updatedBy uint64, req *dto.FinanceCategoryUpdateReq) error {
	cat, err := s.repo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrFinanceCategoryNotFound
		}
		return err
	}

	if req.ParentID != cat.ParentID {
		if req.ParentID != 0 {
			parent, err := s.repo.GetByID(db, req.ParentID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return shared.ErrFinanceCategoryNotFound
				}
				return err
			}
			categoryType := req.CategoryType
			if categoryType == 0 {
				categoryType = cat.CategoryType
			}
			if parent.CategoryType != categoryType {
				return shared.ErrFinanceCategoryTypeMismatch
			}
			if parent.Level+1 > 3 {
				return shared.ErrFinanceCategoryMaxLevel
			}
			cat.Level = parent.Level + 1
		} else {
			cat.Level = 1
		}
		cat.ParentID = req.ParentID
	}

	if req.CategoryType != 0 {
		cat.CategoryType = req.CategoryType
	}

	if req.CategoryName != cat.CategoryName {
		parentID := cat.ParentID
		categoryType := cat.CategoryType
		if existing, err := s.repo.GetByNameAndType(db, req.CategoryName, categoryType, parentID); err == nil && existing != nil && existing.ID != id {
			return shared.ErrFinanceCategoryNameExists
		}
		cat.CategoryName = req.CategoryName
	}

	cat.CategoryCode = req.CategoryCode
	cat.FinanceCode = req.FinanceCode
	cat.Sort = req.Sort
	cat.UpdatedBy = updatedBy
	return s.repo.Update(db, cat)
}

func (s *FinanceCategoryService) Delete(db *gorm.DB, id uint64) error {
	if _, err := s.repo.GetByID(db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrFinanceCategoryNotFound
		}
		return err
	}

	hasChildren, err := s.repo.HasChildren(db, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return shared.ErrFinanceCategoryHasChildren
	}

	hasSync, err := s.repo.HasShopSync(db, id)
	if err != nil {
		return err
	}
	if hasSync {
		return shared.ErrFinanceCategorySynced
	}

	return s.repo.Delete(db, id)
}

func (s *FinanceCategoryService) buildTree(flat []dto.FinanceCategoryResp, parentID uint64) []dto.FinanceCategoryResp {
	var tree []dto.FinanceCategoryResp
	for _, item := range flat {
		if item.ParentID == parentID {
			item.Children = s.buildTree(flat, item.ID)
			if item.Children == nil {
				item.Children = []dto.FinanceCategoryResp{}
			}
			tree = append(tree, item)
		}
	}
	return tree
}

func (s *FinanceCategoryService) catToResp(c *entity.FinanceCategory) dto.FinanceCategoryResp {
	return dto.FinanceCategoryResp{
		ID:           c.ID,
		ParentID:     c.ParentID,
		Level:        c.Level,
		CategoryType: c.CategoryType,
		CategoryCode: c.CategoryCode,
		CategoryName: c.CategoryName,
		FinanceCode:  c.FinanceCode,
		Sort:         c.Sort,
		CreatedAt:    c.CreatedAt,
		CreatedBy:    c.CreatedBy,
		UpdatedAt:    c.UpdatedAt,
		UpdatedBy:    c.UpdatedBy,
	}
}
