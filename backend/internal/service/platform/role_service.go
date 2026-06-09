package platform

import (
	"errors"

	"platform/internal/model/dto"
	"platform/internal/model/enum"
	"platform/internal/model/entity"
	"platform/internal/repository/platform"
	"platform/internal/service/shared"

	"gorm.io/gorm"
)

type RoleService struct {
	roleRepo platform.RoleRepository
}

func NewRoleService(roleRepo platform.RoleRepository) *RoleService {
	return &RoleService{roleRepo: roleRepo}
}

func (s *RoleService) Create(db *gorm.DB, tenantID, createdBy uint64, req *dto.RoleCreateReq) (*entity.SysRole, error) {
	existing, err := s.roleRepo.GetByCode(db, tenantID, req.RoleCode)
	if err == nil && existing != nil {
		return nil, shared.ErrRoleCodeExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	status := req.Status
	if status == 0 {
		status = 1
	}
	dataScope := req.DataScope
	if dataScope == 0 {
		dataScope = 1
	}

	role := &entity.SysRole{
		TenantID:  tenantID,
		RoleName:  req.RoleName,
		RoleCode:  req.RoleCode,
		DataScope: dataScope,
		Sort:      req.Sort,
		Status:    status,
		Remark:    req.Remark,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}

	if err := s.roleRepo.Create(db, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *RoleService) Update(db *gorm.DB, id, updatedBy uint64, req *dto.RoleUpdateReq) error {
	role, err := s.roleRepo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrRoleNotFound
		}
		return err
	}

	if req.RoleCode != role.RoleCode {
		existing, err := s.roleRepo.GetByCode(db, role.TenantID, req.RoleCode)
		if err == nil && existing != nil && existing.ID != id {
			return shared.ErrRoleCodeExists
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	role.RoleName = req.RoleName
	role.RoleCode = req.RoleCode
	role.Remark = req.Remark
	role.DataScope = req.DataScope
	role.Sort = req.Sort
	role.Status = req.Status
	role.UpdatedBy = updatedBy

	return s.roleRepo.Update(db, role)
}

func (s *RoleService) Delete(db *gorm.DB, id uint64) error {
	if _, err := s.roleRepo.GetByID(db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrRoleNotFound
		}
		return err
	}
	return s.roleRepo.Delete(db, id)
}

func (s *RoleService) GetByID(db *gorm.DB, id, tenantID uint64) (*dto.RoleResp, error) {
	role, err := s.roleRepo.GetByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrRoleNotFound
		}
		return nil, err
	}

	permIDs, _ := s.roleRepo.GetPermissionIDs(db, id, tenantID)
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

	resp := shared.RoleToResp(role, perms)
	return &resp, nil
}

func (s *RoleService) List(db *gorm.DB, tenantID uint64, req *dto.RoleListReq) ([]dto.RoleResp, int64, error) {
	roles, total, err := s.roleRepo.List(db, tenantID, req)
	if err != nil {
		return nil, 0, err
	}

	resps := make([]dto.RoleResp, 0, len(roles))
	for _, r := range roles {
		resps = append(resps, shared.RoleToResp(&r, []dto.PermissionResp{}))
	}
	return resps, total, nil
}

func (s *RoleService) AssignPermissions(db *gorm.DB, roleID, tenantID uint64, req *dto.RoleAssignPermsReq, currentUserID uint64, dataScope int16) error {
	if _, err := s.roleRepo.GetByID(db, roleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrRoleNotFound
		}
		return err
	}

	if len(req.PermissionIDs) > 0 {
		perms, err := s.roleRepo.GetPermissionsByIDs(db, req.PermissionIDs)
		if err != nil {
			return err
		}
		if len(perms) != len(req.PermissionIDs) {
			return shared.ErrInvalidPermission
		}
	}

	if dataScope != enum.DataScopeAll {
		userPermIDs, err := s.roleRepo.GetUserPermissionIDs(db, currentUserID, tenantID)
		if err != nil {
			return err
		}
		userPermSet := make(map[uint64]struct{}, len(userPermIDs))
		for _, id := range userPermIDs {
			userPermSet[id] = struct{}{}
		}
		for _, permID := range req.PermissionIDs {
			if _, ok := userPermSet[permID]; !ok {
				return shared.ErrPermissionExceeded
			}
		}
	}

	return s.roleRepo.AssignPermissions(db, roleID, tenantID, req.PermissionIDs)
}

func (s *RoleService) GetPermissionTree(db *gorm.DB, systemType string, currentUserID uint64, dataScope int16, tenantID uint64) ([]dto.PermissionResp, error) {
	perms, err := s.roleRepo.GetPermissionsBySystemType(db, systemType)
	if err != nil {
		return nil, err
	}

	flat := make([]dto.PermissionResp, 0, len(perms))
	for _, p := range perms {
		flat = append(flat, shared.PermissionToResp(&p))
	}

	if dataScope != enum.DataScopeAll {
		userPermIDs, err := s.roleRepo.GetUserPermissionIDs(db, currentUserID, tenantID)
		if err != nil {
			return nil, err
		}
		allowedSet := make(map[uint64]struct{}, len(userPermIDs))
		for _, id := range userPermIDs {
			allowedSet[id] = struct{}{}
		}
		permMap := make(map[uint64]dto.PermissionResp, len(flat))
		for _, p := range flat {
			permMap[p.ID] = p
		}
		for _, id := range userPermIDs {
			p, ok := permMap[id]
			if !ok {
				continue
			}
			parentID := p.ParentID
			for parentID != 0 {
				if _, exists := allowedSet[parentID]; exists {
					break
				}
				allowedSet[parentID] = struct{}{}
				if pp, ok := permMap[parentID]; ok {
					parentID = pp.ParentID
				} else {
					break
				}
			}
		}
		filtered := make([]dto.PermissionResp, 0, len(allowedSet))
		for _, p := range flat {
			if _, ok := allowedSet[p.ID]; ok {
				filtered = append(filtered, p)
			}
		}
		return shared.BuildPermissionTree(filtered, 0), nil
	}

	return shared.BuildPermissionTree(flat, 0), nil
}

func (s *RoleService) GetAssignableRoles(db *gorm.DB, tenantID, currentUserID uint64, dataScope int16) ([]dto.RoleResp, error) {
	var userPermIDs []uint64
	if dataScope != enum.DataScopeAll {
		var err error
		userPermIDs, err = s.roleRepo.GetUserPermissionIDs(db, currentUserID, tenantID)
		if err != nil {
			return nil, err
		}
	}

	roles, err := s.roleRepo.ListAssignable(db, tenantID, userPermIDs)
	if err != nil {
		return nil, err
	}

	resps := make([]dto.RoleResp, 0, len(roles))
	for _, r := range roles {
		resps = append(resps, shared.RoleToResp(&r, nil))
	}
	return resps, nil
}
