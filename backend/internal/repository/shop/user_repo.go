package shop

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(db *gorm.DB, user *entity.SysUser) error
	Update(db *gorm.DB, user *entity.SysUser) error
	Delete(db *gorm.DB, id, tenantID uint64) error
	GetByID(db *gorm.DB, id uint64) (*entity.SysUser, error)
	GetByIDInTenant(db *gorm.DB, id, tenantID uint64) (*entity.SysUser, error)
	GetByUsername(db *gorm.DB, tenantID uint64, username string) (*entity.SysUser, error)
	List(db *gorm.DB, tenantID uint64, req *dto.UserListReq) ([]entity.SysUser, int64, error)
	AssignRoles(db *gorm.DB, userID, tenantID uint64, roleIDs []uint64) error
	GetRoleIDs(db *gorm.DB, userID uint64) ([]uint64, error)
	UpdatePassword(db *gorm.DB, userID, tenantID uint64, hashedPassword string) error
}

type userRepository struct{}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) Create(db *gorm.DB, user *entity.SysUser) error {
	return db.Create(user).Error
}

func (r *userRepository) Update(db *gorm.DB, user *entity.SysUser) error {
	return db.Save(user).Error
}

func (r *userRepository) Delete(db *gorm.DB, id, tenantID uint64) error {
	return db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entity.SysUser{}).Error
}

func (r *userRepository) GetByID(db *gorm.DB, id uint64) (*entity.SysUser, error) {
	var user entity.SysUser
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByIDInTenant(db *gorm.DB, id, tenantID uint64) (*entity.SysUser, error) {
	var user entity.SysUser
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(db *gorm.DB, tenantID uint64, username string) (*entity.SysUser, error) {
	var user entity.SysUser
	if err := db.Where("tenant_id = ? AND username = ?", tenantID, username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(db *gorm.DB, tenantID uint64, req *dto.UserListReq) ([]entity.SysUser, int64, error) {
	var users []entity.SysUser
	var total int64

	query := db.Model(&entity.SysUser{}).Where("tenant_id = ?", tenantID)

	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.RealName != "" {
		query = query.Where("real_name LIKE ?", "%"+req.RealName+"%")
	}
	if req.Phone != "" {
		query = query.Where("phone LIKE ?", "%"+req.Phone+"%")
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

	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) AssignRoles(db *gorm.DB, userID, tenantID uint64, roleIDs []uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var matchedUser struct{ ID uint64 }
		if err := tx.Table("sys_user").
			Select("id").
			Where("id = ? AND tenant_id = ?", userID, tenantID).
			Take(&matchedUser).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&entity.SysUserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Create(&entity.SysUserRole{
				UserID: userID,
				RoleID: roleID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *userRepository) GetRoleIDs(db *gorm.DB, userID uint64) ([]uint64, error) {
	var userRoles []entity.SysUserRole
	if err := db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(userRoles))
	for _, ur := range userRoles {
		ids = append(ids, ur.RoleID)
	}
	return ids, nil
}

func (r *userRepository) UpdatePassword(db *gorm.DB, userID, tenantID uint64, hashedPassword string) error {
	return db.Model(&entity.SysUser{}).Where("id = ? AND tenant_id = ?", userID, tenantID).Update("password", hashedPassword).Error
}
