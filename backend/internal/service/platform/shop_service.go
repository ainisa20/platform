package platform

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"platform/internal/model/dto"
	"platform/internal/model/entity"
	"platform/internal/repository/platform"
	"platform/internal/service/shared"
)

type ShopService struct {
	shopRepo platform.ShopRepository
	userRepo platform.UserRepository
	roleRepo platform.RoleRepository
	deptRepo platform.DeptRepository
}

func NewShopService(
	shopRepo platform.ShopRepository,
	userRepo platform.UserRepository,
	roleRepo platform.RoleRepository,
	deptRepo platform.DeptRepository,
) *ShopService {
	return &ShopService{
		shopRepo: shopRepo,
		userRepo: userRepo,
		roleRepo: roleRepo,
		deptRepo: deptRepo,
	}
}

func (s *ShopService) List(db *gorm.DB, req *dto.ShopListReq) ([]dto.ShopResp, int64, error) {
	shops, total, err := s.shopRepo.List(db, req)
	if err != nil {
		return nil, 0, err
	}
	resps := make([]dto.ShopResp, 0, len(shops))
	for i := range shops {
		resp := s.shopToResp(&shops[i])
		if shops[i].AdminUserID != nil {
			if admin, err := s.userRepo.GetByID(db, *shops[i].AdminUserID); err == nil {
				resp.AdminUsername = admin.Username
			}
		}
		resps = append(resps, resp)
	}
	return resps, total, nil
}

func (s *ShopService) GetByID(db *gorm.DB, id uint64) (*dto.ShopResp, error) {
	shop, err := s.shopRepo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrShopNotFound
		}
		return nil, err
	}
	resp := s.shopToResp(shop)
	if shop.AdminUserID != nil {
		if admin, err := s.userRepo.GetByID(db, *shop.AdminUserID); err == nil {
			resp.AdminUsername = admin.Username
		}
	}
	return &resp, nil
}

func (s *ShopService) Create(db *gorm.DB, createdBy uint64, req *dto.ShopCreateReq) (*dto.ShopResp, error) {
	if existing, err := s.shopRepo.GetByCode(db, req.ShopCode); err == nil && existing != nil {
		return nil, shared.ErrShopCodeExists
	}
	if existing, err := s.userRepo.GetByUsername(db, 0, req.AdminUsername); err == nil && existing != nil {
		return nil, shared.ErrUsernameExists
	}

	realName := req.AdminRealName
	if realName == "" {
		realName = "店长"
	}

	shop := &entity.SysShop{
		ShopCode: req.ShopCode,
		ShopName: req.ShopName,
		Contact:  req.Contact,
		Phone:    req.Phone,
		Email:    req.Email,
		Address:  req.Address,
		Remark:   req.Remark,
		Status:   1,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(shop).Error; err != nil {
			return fmt.Errorf("create shop: %w", err)
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		dept := &entity.SysDept{
			TenantID:  shop.ID,
			ParentID:  0,
			Ancestors: "0",
			DeptName:  "总部",
			Sort:      0,
			Status:    1,
			CreatedBy: createdBy,
			UpdatedBy: createdBy,
		}
		if err := s.deptRepo.Create(tx, dept); err != nil {
			return fmt.Errorf("create root dept: %w", err)
		}

		admin := &entity.SysUser{
			TenantID:  shop.ID,
			DeptID:    &dept.ID,
			Username:  req.AdminUsername,
			Password:  string(hashed),
			RealName:  realName,
			Phone:     req.Phone,
			Email:     req.Email,
			Status:    1,
			CreatedBy: createdBy,
			UpdatedBy: createdBy,
		}
		if err := s.userRepo.Create(tx, admin); err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}

		role := &entity.SysRole{
			TenantID:  shop.ID,
			RoleName:  "店铺管理员",
			RoleCode:  "shop_admin",
			DataScope: 1,
			Sort:      0,
			Status:    1,
			CreatedBy: createdBy,
			UpdatedBy: createdBy,
		}
		if err := tx.Create(role).Error; err != nil {
			return fmt.Errorf("create admin role: %w", err)
		}

		perms, err := s.roleRepo.GetPermissionsBySystemType(tx, "shop")
		if err != nil {
			return fmt.Errorf("fetch perms: %w", err)
		}
		permIDs := make([]uint64, 0, len(perms))
		for _, p := range perms {
			if p.PermsCode != "" {
				permIDs = append(permIDs, p.ID)
			}
		}
		if err := s.roleRepo.AssignPermissions(tx, role.ID, shop.ID, permIDs); err != nil {
			return fmt.Errorf("assign perms: %w", err)
		}

		if err := s.userRepo.AssignRoles(tx, admin.ID, []uint64{role.ID}); err != nil {
			return fmt.Errorf("assign role to user: %w", err)
		}

		if err := s.shopRepo.UpdateAdminUserID(tx, shop.ID, admin.ID); err != nil {
			return fmt.Errorf("link admin: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	refreshed, err := s.shopRepo.GetByID(db, shop.ID)
	if err != nil {
		return nil, err
	}
	resp := s.shopToResp(refreshed)
	resp.AdminUsername = req.AdminUsername
	return &resp, nil
}

func (s *ShopService) Update(db *gorm.DB, id, updatedBy uint64, req *dto.ShopUpdateReq) error {
	shop, err := s.shopRepo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopNotFound
		}
		return err
	}
	shop.ShopName = req.ShopName
	shop.Contact = req.Contact
	shop.Phone = req.Phone
	shop.Email = req.Email
	shop.Address = req.Address
	shop.Remark = req.Remark
	shop.UpdatedBy = updatedBy
	return s.shopRepo.Update(db, shop)
}

func (s *ShopService) UpdateStatus(db *gorm.DB, id uint64, status int16) error {
	if _, err := s.shopRepo.GetByID(db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopNotFound
		}
		return err
	}
	return s.shopRepo.UpdateStatus(db, id, status)
}

func (s *ShopService) Delete(db *gorm.DB, id uint64) error {
	if _, err := s.shopRepo.GetByID(db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopNotFound
		}
		return err
	}
	var count int64
	if err := db.Model(&entity.SysUser{}).Where("tenant_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 1 {
		return shared.ErrShopHasUsers
	}
	return s.shopRepo.Delete(db, id)
}

func (s *ShopService) ResetAdminPassword(db *gorm.DB, id uint64, newPassword string) error {
	shop, err := s.shopRepo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopNotFound
		}
		return err
	}
	if shop.AdminUserID == nil {
		return errors.New("店铺未关联管理员账户")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(db, *shop.AdminUserID, string(hashed))
}

func (s *ShopService) shopToResp(s2 *entity.SysShop) dto.ShopResp {
	return dto.ShopResp{
		ID:           s2.ID,
		ShopCode:     s2.ShopCode,
		ShopName:     s2.ShopName,
		Contact:      s2.Contact,
		Phone:        s2.Phone,
		Email:        s2.Email,
		Address:      s2.Address,
		Remark:       s2.Remark,
		Status:       s2.Status,
		AdminUserID:  s2.AdminUserID,
		ExpiresAt:    s2.ExpiresAt,
		CreatedAt:    s2.CreatedAt,
		CreatedBy:    s2.CreatedBy,
		UpdatedAt:    s2.UpdatedAt,
		UpdatedBy:    s2.UpdatedBy,
	}
}
