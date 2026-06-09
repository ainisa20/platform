package platform

import (
	"fmt"

	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type DeptRepository interface {
	Create(db *gorm.DB, dept *entity.SysDept) error
	Update(db *gorm.DB, dept *entity.SysDept) error
	Delete(db *gorm.DB, id, tenantID uint64) error
	GetByID(db *gorm.DB, id uint64) (*entity.SysDept, error)
	ListByTenantID(db *gorm.DB, tenantID uint64) ([]entity.SysDept, error)
	GetDescendantIDs(db *gorm.DB, deptID, tenantID uint64) ([]uint64, error)
	CountUsersByDeptID(db *gorm.DB, deptID uint64) (int64, error)
	RebuildClosureForSubtree(db *gorm.DB, subtreeIDs []uint64, tenantID uint64) error
}

type deptRepository struct{}

func NewDeptRepository() DeptRepository {
	return &deptRepository{}
}

func (r *deptRepository) Create(db *gorm.DB, dept *entity.SysDept) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if dept.ParentID > 0 {
			var parent entity.SysDept
			if err := tx.First(&parent, dept.ParentID).Error; err != nil {
				return fmt.Errorf("parent dept not found: %w", err)
			}
			dept.Ancestors = parent.Ancestors + "," + fmt.Sprintf("%d", parent.ID)
		} else {
			dept.Ancestors = "0"
		}

		if err := tx.Create(dept).Error; err != nil {
			return err
		}

		if err := tx.Create(&entity.SysDeptClosure{
			TenantID:     dept.TenantID,
			AncestorID:   dept.ID,
			DescendantID: dept.ID,
			Depth:        0,
		}).Error; err != nil {
			return err
		}

		if dept.ParentID > 0 {
			var parentClosures []entity.SysDeptClosure
			if err := tx.Where("descendant_id = ? AND tenant_id = ?", dept.ParentID, dept.TenantID).
				Find(&parentClosures).Error; err != nil {
				return err
			}
			for _, pc := range parentClosures {
				if err := tx.Create(&entity.SysDeptClosure{
					TenantID:     dept.TenantID,
					AncestorID:   pc.AncestorID,
					DescendantID: dept.ID,
					Depth:        pc.Depth + 1,
				}).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *deptRepository) Update(db *gorm.DB, dept *entity.SysDept) error {
	return db.Save(dept).Error
}

func (r *deptRepository) Delete(db *gorm.DB, id, tenantID uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&entity.SysDept{}, id).Error; err != nil {
			return err
		}
		if err := tx.Where("tenant_id = ? AND (ancestor_id = ? OR descendant_id = ?)",
			tenantID, id, id).Delete(&entity.SysDeptClosure{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *deptRepository) GetByID(db *gorm.DB, id uint64) (*entity.SysDept, error) {
	var dept entity.SysDept
	if err := db.First(&dept, id).Error; err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *deptRepository) ListByTenantID(db *gorm.DB, tenantID uint64) ([]entity.SysDept, error) {
	var depts []entity.SysDept
	if err := db.Where("tenant_id = ?", tenantID).
		Order("sort ASC, id ASC").Find(&depts).Error; err != nil {
		return nil, err
	}
	return depts, nil
}

func (r *deptRepository) GetDescendantIDs(db *gorm.DB, deptID, tenantID uint64) ([]uint64, error) {
	var closures []entity.SysDeptClosure
	if err := db.Where("ancestor_id = ? AND tenant_id = ? AND depth > 0",
		deptID, tenantID).Find(&closures).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(closures))
	for _, c := range closures {
		ids = append(ids, c.DescendantID)
	}
	return ids, nil
}

func (r *deptRepository) CountUsersByDeptID(db *gorm.DB, deptID uint64) (int64, error) {
	var count int64
	if err := db.Model(&entity.SysUser{}).Where("dept_id = ?", deptID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *deptRepository) RebuildClosureForSubtree(db *gorm.DB, subtreeIDs []uint64, tenantID uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ? AND (ancestor_id IN ? OR descendant_id IN ?)",
			tenantID, subtreeIDs, subtreeIDs).Delete(&entity.SysDeptClosure{}).Error; err != nil {
			return err
		}

		for _, id := range subtreeIDs {
			if err := tx.Create(&entity.SysDeptClosure{
				TenantID:     tenantID,
				AncestorID:   id,
				DescendantID: id,
				Depth:        0,
			}).Error; err != nil {
				return err
			}

			var dept entity.SysDept
			if err := tx.First(&dept, id).Error; err != nil {
				return err
			}

			if dept.ParentID > 0 {
				var parentClosures []entity.SysDeptClosure
				if err := tx.Where("descendant_id = ? AND tenant_id = ?",
					dept.ParentID, tenantID).Find(&parentClosures).Error; err != nil {
					return err
				}
				for _, pc := range parentClosures {
					if err := tx.Create(&entity.SysDeptClosure{
						TenantID:     tenantID,
						AncestorID:   pc.AncestorID,
						DescendantID: id,
						Depth:        pc.Depth + 1,
					}).Error; err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}
