package repository

import (
	"time"

	"gorm.io/gorm"
	"platform/internal/model/entity"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetByUsername(username string, tenantID uint64) (*entity.SysUser, error) {
	var user entity.SysUser
	err := r.db.Where("username = ? AND tenant_id = ?", username, tenantID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) GetUserByID(userID, tenantID uint64) (*entity.SysUser, error) {
	var user entity.SysUser
	err := r.db.Where("id = ? AND tenant_id = ?", userID, tenantID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) UpdateLoginInfo(userID, tenantID uint64, ip string) error {
	now := time.Now()
	return r.db.Model(&entity.SysUser{}).
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
		}).Error
}

func (r *AuthRepository) GetUserRoleCodes(userID, tenantID uint64) ([]string, error) {
	var codes []string
	err := r.db.Table("sys_user_role").
		Select("DISTINCT sys_role.role_code").
		Joins("JOIN sys_role ON sys_role.id = sys_user_role.role_id AND sys_role.deleted_at IS NULL").
		Where("sys_user_role.user_id = ? AND sys_role.tenant_id = ?", userID, tenantID).
		Pluck("sys_role.role_code", &codes).Error
	return codes, err
}

func (r *AuthRepository) GetUserPermissionCodes(userID, tenantID uint64) ([]string, error) {
	var codes []string
	err := r.db.Raw(`
		SELECT DISTINCT p.perms_code
		FROM sys_user_role ur
		JOIN sys_role_permission rp ON rp.role_id = ur.role_id AND rp.tenant_id = ?
		JOIN sys_permission p ON p.id = rp.permission_id
		WHERE ur.user_id = ?
		  AND p.perms_code IS NOT NULL AND p.perms_code <> ''
	`, tenantID, userID).Scan(&codes).Error
	return codes, err
}

func (r *AuthRepository) GetShopTenantIDByCode(shopCode string) (uint64, int16, error) {
	type shopRow struct {
		ID     uint64 `gorm:"primaryKey"`
		Status int16
	}
	var shop shopRow
	err := r.db.Table("sys_shop").
		Where("shop_code = ? AND deleted_at IS NULL", shopCode).
		First(&shop).Error
	if err != nil {
		return 0, 0, err
	}
	return shop.ID, shop.Status, nil
}

func (r *AuthRepository) GetUserMaxDataScope(userID, tenantID uint64) (int16, error) {
	var maxScope int16
	err := r.db.Table("sys_user_role").
		Select("COALESCE(MAX(sys_role.data_scope), 4)").
		Joins("JOIN sys_role ON sys_role.id = sys_user_role.role_id AND sys_role.deleted_at IS NULL").
		Where("sys_user_role.user_id = ? AND sys_role.tenant_id = ?", userID, tenantID).
		Scan(&maxScope).Error
	return maxScope, err
}

func (r *AuthRepository) GetPermissionsBySystemType(systemType string) ([]entity.SysPermission, error) {
	var perms []entity.SysPermission
	err := r.db.Where("system_type = ? AND status = ?", systemType, 1).
		Order("parent_id ASC, sort ASC").
		Find(&perms).Error
	return perms, err
}
