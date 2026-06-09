package shop

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"platform/internal/middleware"
	"platform/internal/model/dto"
	"platform/internal/model/enum"
	"platform/internal/model/entity"
	"platform/internal/repository/shop"
	"platform/internal/service/shared"
)

type UserService struct {
	userRepo shop.UserRepository
	roleRepo shop.RoleRepository
}

func NewUserService(userRepo shop.UserRepository, roleRepo shop.RoleRepository) *UserService {
	return &UserService{userRepo: userRepo, roleRepo: roleRepo}
}

func (s *UserService) Create(db *gorm.DB, tenantID, createdBy uint64, req *dto.UserCreateReq, dataScope int16, currentUserDeptID uint64, currentUserID uint64) (*entity.SysUser, error) {
	if dataScope != enum.DataScopeAll {
		var allowedDeptIDs []uint64
		if dataScope == enum.DataScopeDeptAndSub {
			err := db.Table("sys_dept_closure").
				Where("ancestor_id = ? AND tenant_id = ?", currentUserDeptID, tenantID).
				Pluck("descendant_id", &allowedDeptIDs).Error
			if err != nil || len(allowedDeptIDs) == 0 {
				allowedDeptIDs = []uint64{currentUserDeptID}
			}
		} else {
			allowedDeptIDs = []uint64{currentUserDeptID}
		}
		deptOK := false
		for _, id := range allowedDeptIDs {
			if id == *req.DeptID {
				deptOK = true
				break
			}
		}
		if !deptOK {
			return nil, shared.ErrDeptOutOfScope
		}

		if len(req.RoleIDs) > 0 {
			userPermIDs, err := s.roleRepo.GetUserPermissionIDs(db, currentUserID, tenantID)
			if err != nil {
				return nil, err
			}
			assignableRoles, err := s.roleRepo.ListAssignable(db, tenantID, userPermIDs)
			if err != nil {
				return nil, err
			}
			assignableSet := make(map[uint64]struct{}, len(assignableRoles))
			for _, r := range assignableRoles {
				assignableSet[r.ID] = struct{}{}
			}
			for _, roleID := range req.RoleIDs {
				if _, ok := assignableSet[roleID]; !ok {
					return nil, shared.ErrRoleNotAssignable
				}
			}
		}
	}

	existing, err := s.userRepo.GetByUsername(db, tenantID, req.Username)
	if err == nil && existing != nil {
		return nil, shared.ErrUsernameExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt failed: %w", err)
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	user := &entity.SysUser{
		TenantID:  tenantID,
		DeptID:    req.DeptID,
		Username:  req.Username,
		Password:  string(hashedPwd),
		RealName:  req.RealName,
		Phone:     req.Phone,
		Email:     req.Email,
		Status:    status,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}

	if err := s.userRepo.Create(db, user); err != nil {
		return nil, err
	}

	if len(req.RoleIDs) > 0 {
		if err := s.userRepo.AssignRoles(db, user.ID, req.RoleIDs); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *UserService) Update(db *gorm.DB, id, updatedBy uint64, req *dto.UserUpdateReq, dataScope int16, currentUserDeptID uint64, currentUserID uint64, tenantID uint64) error {
	if dataScope != enum.DataScopeAll {
		if req.DeptID != nil {
			var allowedDeptIDs []uint64
			if dataScope == enum.DataScopeDeptAndSub {
				err := db.Table("sys_dept_closure").
					Where("ancestor_id = ? AND tenant_id = ?", currentUserDeptID, tenantID).
					Pluck("descendant_id", &allowedDeptIDs).Error
				if err != nil || len(allowedDeptIDs) == 0 {
					allowedDeptIDs = []uint64{currentUserDeptID}
				}
			} else {
				allowedDeptIDs = []uint64{currentUserDeptID}
			}
			deptOK := false
			for _, did := range allowedDeptIDs {
				if did == *req.DeptID {
					deptOK = true
					break
				}
			}
			if !deptOK {
				return shared.ErrDeptOutOfScope
			}
		}

		if req.RoleIDs != nil && len(req.RoleIDs) > 0 {
			userPermIDs, err := s.roleRepo.GetUserPermissionIDs(db, currentUserID, tenantID)
			if err != nil {
				return err
			}
			assignableRoles, err := s.roleRepo.ListAssignable(db, tenantID, userPermIDs)
			if err != nil {
				return err
			}
			assignableSet := make(map[uint64]struct{}, len(assignableRoles))
			for _, r := range assignableRoles {
				assignableSet[r.ID] = struct{}{}
			}
			for _, roleID := range req.RoleIDs {
				if _, ok := assignableSet[roleID]; !ok {
					return shared.ErrRoleNotAssignable
				}
			}
		}
	}

	user, err := s.userRepo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrUserNotFound
		}
		return err
	}

	updates := map[string]interface{}{"updated_by": updatedBy}
	if req.RealName != "" {
		updates["real_name"] = req.RealName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.DeptID != nil {
		updates["dept_id"] = *req.DeptID
	}
	if req.Status != 0 {
		updates["status"] = req.Status
	}

	if err := db.Model(user).Updates(updates).Error; err != nil {
		return err
	}

	if req.RoleIDs != nil {
		if err := s.userRepo.AssignRoles(db, id, req.RoleIDs); err != nil {
			return err
		}
	}

	return nil
}

func (s *UserService) Delete(db *gorm.DB, id uint64) error {
	if _, err := s.userRepo.GetByID(db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrUserNotFound
		}
		return err
	}
	return s.userRepo.Delete(db, id)
}

func (s *UserService) GetByID(db *gorm.DB, id uint64) (*dto.UserResp, error) {
	user, err := s.userRepo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrUserNotFound
		}
		return nil, err
	}

	roles, err := s.loadRolesWithPerms(db, user.ID, user.TenantID)
	if err != nil {
		return nil, err
	}

	return shared.UserToResp(user, roles), nil
}

func (s *UserService) List(c *gin.Context, db *gorm.DB, tenantID uint64, req *dto.UserListReq) ([]dto.UserResp, int64, error) {
	users, total, err := s.userRepo.List(middleware.ApplyUserIDScope(c, db), tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	resps := make([]dto.UserResp, 0, len(users))
	for i := range users {
		roleIDs, _ := s.userRepo.GetRoleIDs(db, users[i].ID)
		roles, _ := s.loadRolesBrief(db, roleIDs)
		resps = append(resps, *shared.UserToResp(&users[i], roles))
	}

	return resps, total, nil
}

func (s *UserService) ResetPassword(db *gorm.DB, id uint64, req *dto.PasswordResetReq) error {
	if _, err := s.userRepo.GetByID(db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrUserNotFound
		}
		return err
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt failed: %w", err)
	}

	return s.userRepo.UpdatePassword(db, id, string(hashedPwd))
}

func (s *UserService) AssignRoles(db *gorm.DB, userID uint64, roleIDs []uint64) error {
	if _, err := s.userRepo.GetByID(db, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrUserNotFound
		}
		return err
	}
	return s.userRepo.AssignRoles(db, userID, roleIDs)
}

func (s *UserService) loadRolesBrief(db *gorm.DB, roleIDs []uint64) ([]dto.RoleResp, error) {
	if len(roleIDs) == 0 {
		return []dto.RoleResp{}, nil
	}
	entities, err := s.roleRepo.GetByIDs(db, roleIDs)
	if err != nil {
		return nil, err
	}
	resps := make([]dto.RoleResp, 0, len(entities))
	for _, r := range entities {
		resps = append(resps, shared.RoleToResp(&r, nil))
	}
	return resps, nil
}

func (s *UserService) loadRolesWithPerms(db *gorm.DB, userID, tenantID uint64) ([]dto.RoleResp, error) {
	roleIDs, err := s.userRepo.GetRoleIDs(db, userID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []dto.RoleResp{}, nil
	}

	roles, err := s.roleRepo.GetByIDs(db, roleIDs)
	if err != nil {
		return nil, err
	}

	resps := make([]dto.RoleResp, 0, len(roles))
	for _, r := range roles {
		permIDs, _ := s.roleRepo.GetPermissionIDs(db, r.ID, tenantID)
		var perms []dto.PermissionResp
		if len(permIDs) > 0 {
			permEntities, err := s.roleRepo.GetPermissionsByIDs(db, permIDs)
			if err == nil {
				perms = make([]dto.PermissionResp, 0, len(permEntities))
				for _, p := range permEntities {
					perms = append(perms, shared.PermissionToResp(&p))
				}
			}
		}
		resps = append(resps, shared.RoleToResp(&r, perms))
	}
	return resps, nil
}
