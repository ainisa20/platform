package shop

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(db *gorm.DB, role *entity.SysRole) error
	Update(db *gorm.DB, role *entity.SysRole) error
	Delete(db *gorm.DB, id uint64) error
	GetByID(db *gorm.DB, id uint64) (*entity.SysRole, error)
	GetByIDs(db *gorm.DB, ids []uint64) ([]entity.SysRole, error)
	GetByCode(db *gorm.DB, tenantID uint64, roleCode string) (*entity.SysRole, error)
	List(db *gorm.DB, tenantID uint64, req *dto.RoleListReq) ([]entity.SysRole, int64, error)
	AssignPermissions(db *gorm.DB, roleID, tenantID uint64, permIDs []uint64) error
	GetPermissionIDs(db *gorm.DB, roleID, tenantID uint64) ([]uint64, error)
	GetPermissionsBySystemType(db *gorm.DB, systemType string) ([]entity.SysPermission, error)
	GetPermissionsByIDs(db *gorm.DB, ids []uint64) ([]entity.SysPermission, error)
	GetUserPermissionIDs(db *gorm.DB, userID, tenantID uint64) ([]uint64, error)
	ListAssignable(db *gorm.DB, tenantID uint64, userPermIDs []uint64) ([]entity.SysRole, error)
}

type roleRepository struct{}

func NewRoleRepository() RoleRepository {
	return &roleRepository{}
}

func (r *roleRepository) Create(db *gorm.DB, role *entity.SysRole) error {
	return db.Create(role).Error
}

func (r *roleRepository) Update(db *gorm.DB, role *entity.SysRole) error {
	return db.Save(role).Error
}

func (r *roleRepository) Delete(db *gorm.DB, id uint64) error {
	return db.Delete(&entity.SysRole{}, id).Error
}

func (r *roleRepository) GetByID(db *gorm.DB, id uint64) (*entity.SysRole, error) {
	var role entity.SysRole
	if err := db.First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByIDs(db *gorm.DB, ids []uint64) ([]entity.SysRole, error) {
	var roles []entity.SysRole
	if len(ids) == 0 {
		return roles, nil
	}
	if err := db.Where("id IN ?", ids).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) GetByCode(db *gorm.DB, tenantID uint64, roleCode string) (*entity.SysRole, error) {
	var role entity.SysRole
	if err := db.Where("tenant_id = ? AND role_code = ?", tenantID, roleCode).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) List(db *gorm.DB, tenantID uint64, req *dto.RoleListReq) ([]entity.SysRole, int64, error) {
	var roles []entity.SysRole
	var total int64

	query := db.Model(&entity.SysRole{}).Where("tenant_id = ?", tenantID)

	if req.RoleName != "" {
		query = query.Where("role_name LIKE ?", "%"+req.RoleName+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
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

	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id ASC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *roleRepository) AssignPermissions(db *gorm.DB, roleID, tenantID uint64, permIDs []uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ? AND tenant_id = ?", roleID, tenantID).
			Delete(&entity.SysRolePermission{}).Error; err != nil {
			return err
		}
		for _, permID := range permIDs {
			if err := tx.Create(&entity.SysRolePermission{
				TenantID:     tenantID,
				RoleID:       roleID,
				PermissionID: permID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *roleRepository) GetPermissionIDs(db *gorm.DB, roleID, tenantID uint64) ([]uint64, error) {
	var rolePerms []entity.SysRolePermission
	if err := db.Where("role_id = ? AND tenant_id = ?", roleID, tenantID).Find(&rolePerms).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(rolePerms))
	for _, rp := range rolePerms {
		ids = append(ids, rp.PermissionID)
	}
	return ids, nil
}

func (r *roleRepository) GetPermissionsBySystemType(db *gorm.DB, systemType string) ([]entity.SysPermission, error) {
	var perms []entity.SysPermission
	if err := db.Where("system_type = ? AND status = 1", systemType).
		Order("sort ASC, id ASC").Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *roleRepository) GetPermissionsByIDs(db *gorm.DB, ids []uint64) ([]entity.SysPermission, error) {
	var perms []entity.SysPermission
	if len(ids) == 0 {
		return perms, nil
	}
	if err := db.Where("id IN ?", ids).Order("sort ASC, id ASC").Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *roleRepository) GetUserPermissionIDs(db *gorm.DB, userID, tenantID uint64) ([]uint64, error) {
	var ids []uint64
	err := db.Table("sys_user_role").
		Select("DISTINCT sys_role_permission.permission_id").
		Joins("JOIN sys_role_permission ON sys_role_permission.role_id = sys_user_role.role_id AND sys_role_permission.tenant_id = ?", tenantID).
		Where("sys_user_role.user_id = ?", userID).
		Pluck("sys_role_permission.permission_id", &ids).Error
	return ids, err
}

func (r *roleRepository) ListAssignable(db *gorm.DB, tenantID uint64, userPermIDs []uint64) ([]entity.SysRole, error) {
	var roles []entity.SysRole
	query := db.Where("tenant_id = ?", tenantID)
	if userPermIDs == nil {
	} else if len(userPermIDs) == 0 {
		subQuery := db.Table("sys_role_permission").
			Select("1").
			Where("role_id = sys_role.id AND tenant_id = ?", tenantID)
		query = query.Where("NOT EXISTS (?)", subQuery)
	} else {
		subQuery := db.Table("sys_role_permission").
			Select("1").
			Where("role_id = sys_role.id AND tenant_id = ? AND permission_id NOT IN ?", tenantID, userPermIDs)
		query = query.Where("NOT EXISTS (?)", subQuery)
	}
	if err := query.Order("sort ASC, id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
