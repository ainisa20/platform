package shop

import (
	"errors"

	"gorm.io/gorm"

	"platform/internal/model/dto"
	"platform/internal/model/entity"
	"platform/internal/repository/platform"
	"platform/internal/repository/shop"
	"platform/internal/service/shared"
)

type ShopProductService struct {
	repo         shop.ShopProductRepository
	platProdRepo platform.ProductRepository
	platCatRepo  platform.CategoryRepository
}

func NewShopProductService(repo shop.ShopProductRepository, platProdRepo platform.ProductRepository, platCatRepo platform.CategoryRepository) *ShopProductService {
	return &ShopProductService{repo: repo, platProdRepo: platProdRepo, platCatRepo: platCatRepo}
}

func (s *ShopProductService) List(db *gorm.DB, tenantID uint64, req *dto.ShopProductListReq) ([]dto.ShopProductResp, int64, error) {
	products, total, err := s.repo.List(db, tenantID, req)
	if err != nil {
		return nil, 0, err
	}
	list := make([]dto.ShopProductResp, 0, len(products))
	for i := range products {
		list = append(list, shopProductToResp(&products[i]))
	}
	return list, total, nil
}

func (s *ShopProductService) ListPlatformAvailable(db *gorm.DB, tenantID uint64) ([]dto.ShopPlatformProductResp, error) {
	statusOnShelf := int16(1)
	platProducts, _, err := s.platProdRepo.List(db, &dto.ProductListReq{
		Page:     1,
		PageSize: 10000,
		Status:   &statusOnShelf,
	})
	if err != nil {
		return nil, err
	}

	platIDs := make([]uint64, 0, len(platProducts))
	for _, p := range platProducts {
		platIDs = append(platIDs, p.ID)
	}

	selected, _ := s.repo.FindByPlatformIDs(db, tenantID, platIDs)
	selectedMap := make(map[uint64]bool, len(selected))
	for _, sp := range selected {
		selectedMap[sp.PlatformProductID] = true
	}

	cats, _, _ := s.platCatRepo.List(db, &dto.CategoryListReq{PageSize: 10000})
	catMap := make(map[uint64]string, len(cats))
	for _, c := range cats {
		catMap[c.ID] = c.CategoryName
	}

	list := make([]dto.ShopPlatformProductResp, 0, len(platProducts))
	for i := range platProducts {
		p := &platProducts[i]
		resp := dto.ShopPlatformProductResp{
			ID:           p.ID,
			ProductCode:  p.ProductCode,
			ProductName:  p.ProductName,
			Price:        p.Price,
			CategoryID:   p.CategoryID,
			Description:  p.Description,
		}
		if p.CategoryID != nil {
			if name, ok := catMap[*p.CategoryID]; ok {
				resp.CategoryName = name
			}
		}
		_ = selectedMap[p.ID]
		list = append(list, resp)
	}
	return list, nil
}

func (s *ShopProductService) Select(db *gorm.DB, tenantID, createdBy uint64, req *dto.ShopProductSelectReq) error {
	created := 0
	for _, platformID := range req.PlatformProductIDs {
		existing, err := s.repo.FindByPlatformID(db, tenantID, platformID)
		if err == nil && existing != nil {
			continue
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		platProd, err := s.platProdRepo.GetByID(db, platformID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}
		if platProd.Status != 1 {
			return shared.ErrShopProductNotOnShelf
		}

		sp := &entity.ShopProduct{
			TenantID:          tenantID,
			PlatformProductID: platformID,
			ProductCode:       platProd.ProductCode,
			ProductName:       platProd.ProductName,
			PlatformPrice:     platProd.Price,
			ShopPrice:         platProd.Price,
			Status:            1,
			CreatedBy:         createdBy,
			UpdatedBy:         createdBy,
		}
		if err := s.repo.Create(db, sp); err != nil {
			return err
		}
		created++
	}

	if created == 0 {
		return shared.ErrShopProductAlreadySelected
	}
	return nil
}

func (s *ShopProductService) UpdatePrice(db *gorm.DB, tenantID, id, updatedBy uint64, req *dto.ShopProductPriceReq) error {
	sp, err := s.repo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopProductNotFound
		}
		return err
	}
	_ = tenantID
	sp.ShopPrice = req.ShopPrice
	sp.UpdatedBy = updatedBy
	return s.repo.Update(db, sp)
}

func (s *ShopProductService) UpdateStatus(db *gorm.DB, tenantID, id, updatedBy uint64, req *dto.ShopProductStatusReq) error {
	sp, err := s.repo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopProductNotFound
		}
		return err
	}
	_ = tenantID
	sp.Status = req.Status
	sp.UpdatedBy = updatedBy
	return s.repo.Update(db, sp)
}

func (s *ShopProductService) DeleteSelection(db *gorm.DB, tenantID, id uint64) error {
	sp, err := s.repo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopProductNotFound
		}
		return err
	}
	_ = sp
	_ = tenantID
	return s.repo.Delete(db, id)
}

func shopProductToResp(sp *entity.ShopProduct) dto.ShopProductResp {
	return dto.ShopProductResp{
		ID:                sp.ID,
		PlatformProductID: sp.PlatformProductID,
		ProductCode:       sp.ProductCode,
		ProductName:       sp.ProductName,
		PlatformPrice:     sp.PlatformPrice,
		ShopPrice:         sp.ShopPrice,
		Status:            sp.Status,
		CreatedAt:         sp.CreatedAt,
		CreatedBy:         sp.CreatedBy,
		UpdatedAt:         sp.UpdatedAt,
		UpdatedBy:         sp.UpdatedBy,
	}
}
