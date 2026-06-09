package shop

import (
	"errors"
	"sort"

	"gorm.io/gorm"

	"platform/internal/model/dto"
	"platform/internal/model/entity"
	"platform/internal/repository/platform"
	"platform/internal/repository/shop"
	"platform/internal/service/shared"
)

type ShopFinCategoryService struct {
	repo        shop.ShopFinCategoryRepository
	platCatRepo platform.FinanceCategoryRepository
}

func NewShopFinCategoryService(repo shop.ShopFinCategoryRepository, platCatRepo platform.FinanceCategoryRepository) *ShopFinCategoryService {
	return &ShopFinCategoryService{repo: repo, platCatRepo: platCatRepo}
}

func (s *ShopFinCategoryService) ListSynced(db *gorm.DB, tenantID uint64, req *dto.ShopFinCategoryListReq) ([]dto.ShopFinCategoryResp, error) {
	cats, err := s.repo.ListSynced(db, tenantID, req)
	if err != nil {
		return nil, err
	}
	flat := make([]dto.ShopFinCategoryResp, 0, len(cats))
	for i := range cats {
		flat = append(flat, shopCatToResp(&cats[i]))
	}
	return buildShopFinCatTree(flat, 0), nil
}

func (s *ShopFinCategoryService) ListAvailable(db *gorm.DB) ([]dto.ShopFinCategoryAvailableResp, error) {
	cats, err := s.platCatRepo.List(db, &dto.FinanceCategoryListReq{})
	if err != nil {
		return nil, err
	}
	flat := make([]dto.ShopFinCategoryAvailableResp, 0, len(cats))
	for i := range cats {
		flat = append(flat, platCatToAvailableResp(&cats[i]))
	}
	return buildAvailableTree(flat, 0), nil
}

func (s *ShopFinCategoryService) Sync(db *gorm.DB, tenantID, createdBy uint64, req *dto.ShopFinCategorySyncReq) error {
	allSynced, _ := s.repo.ListSynced(db, tenantID, &dto.ShopFinCategoryListReq{})
	syncedMap := make(map[uint64]bool, len(allSynced))
	for _, sc := range allSynced {
		syncedMap[sc.PlatformCategoryID] = true
	}

	toSync := make(map[uint64]*entity.FinanceCategory)
	for _, pid := range req.PlatformCategoryIDs {
		if syncedMap[pid] {
			continue
		}
		cat, err := s.platCatRepo.GetByID(db, pid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}
		s.walkAncestors(db, cat, toSync, syncedMap)
	}

	if len(toSync) == 0 {
		return shared.ErrShopFinCategoryAlreadySynced
	}

	sorted := make([]*entity.FinanceCategory, 0, len(toSync))
	for _, cat := range toSync {
		sorted = append(sorted, cat)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Level < sorted[j].Level
	})

	for _, cat := range sorted {
		var shopParentID uint64
		if cat.ParentID != 0 {
			parentSynced, err := s.repo.FindByPlatformID(db, tenantID, cat.ParentID)
			if err == nil && parentSynced != nil {
				shopParentID = parentSynced.ID
			}
		}
		if err := s.repo.Create(db, &entity.ShopFinanceCategory{
			TenantID:           tenantID,
			PlatformCategoryID: cat.ID,
			ParentID:           shopParentID,
			Level:              cat.Level,
			CategoryType:       cat.CategoryType,
			CategoryCode:       cat.CategoryCode,
			CategoryName:       cat.CategoryName,
			CreatedBy:          createdBy,
			UpdatedBy:          createdBy,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *ShopFinCategoryService) CancelSync(db *gorm.DB, tenantID, id, userID uint64) error {
	cat, err := s.repo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopFinCategoryNotFound
		}
		return err
	}
	_ = cat

	hasRef, err := s.repo.HasReference(db, id)
	if err != nil {
		return err
	}
	if hasRef {
		return shared.ErrShopFinCategoryReferenced
	}

	hasChildren, err := s.repo.HasChildren(db, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return shared.ErrFinanceCategoryHasChildren
	}

	return s.repo.Delete(db, id)
}

func (s *ShopFinCategoryService) walkAncestors(db *gorm.DB, cat *entity.FinanceCategory, toSync map[uint64]*entity.FinanceCategory, syncedMap map[uint64]bool) {
	if syncedMap[cat.ID] {
		return
	}
	if _, exists := toSync[cat.ID]; exists {
		return
	}
	toSync[cat.ID] = cat

	if cat.ParentID != 0 {
		parent, err := s.platCatRepo.GetByID(db, cat.ParentID)
		if err != nil {
			return
		}
		s.walkAncestors(db, parent, toSync, syncedMap)
	}
}

func shopCatToResp(c *entity.ShopFinanceCategory) dto.ShopFinCategoryResp {
	return dto.ShopFinCategoryResp{
		ID:                 c.ID,
		PlatformCategoryID: c.PlatformCategoryID,
		ParentID:           c.ParentID,
		Level:              c.Level,
		CategoryType:       c.CategoryType,
		CategoryCode:       c.CategoryCode,
		CategoryName:       c.CategoryName,
		CreatedAt:          c.CreatedAt,
		CreatedBy:          c.CreatedBy,
	}
}

func platCatToAvailableResp(c *entity.FinanceCategory) dto.ShopFinCategoryAvailableResp {
	return dto.ShopFinCategoryAvailableResp{
		ID:           c.ID,
		ParentID:     c.ParentID,
		Level:        c.Level,
		CategoryType: c.CategoryType,
		CategoryCode: c.CategoryCode,
		CategoryName: c.CategoryName,
	}
}

func buildShopFinCatTree(flat []dto.ShopFinCategoryResp, parentID uint64) []dto.ShopFinCategoryResp {
	var tree []dto.ShopFinCategoryResp
	for _, item := range flat {
		if item.ParentID == parentID {
			item.Children = buildShopFinCatTree(flat, item.ID)
			if item.Children == nil {
				item.Children = []dto.ShopFinCategoryResp{}
			}
			tree = append(tree, item)
		}
	}
	return tree
}

func buildAvailableTree(flat []dto.ShopFinCategoryAvailableResp, parentID uint64) []dto.ShopFinCategoryAvailableResp {
	var tree []dto.ShopFinCategoryAvailableResp
	for _, item := range flat {
		if item.ParentID == parentID {
			item.Children = buildAvailableTree(flat, item.ID)
			if item.Children == nil {
				item.Children = []dto.ShopFinCategoryAvailableResp{}
			}
			tree = append(tree, item)
		}
	}
	return tree
}
