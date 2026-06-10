package shop

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"platform/internal/middleware"
	"platform/internal/model/dto"
	"platform/internal/model/entity"
	"platform/internal/repository/shop"
	"platform/internal/service/shared"
)

type DeptService struct {
	deptRepo shop.DeptRepository
	userRepo shop.UserRepository
}

func NewDeptService(deptRepo shop.DeptRepository, userRepo shop.UserRepository) *DeptService {
	return &DeptService{deptRepo: deptRepo, userRepo: userRepo}
}

func (s *DeptService) Create(db *gorm.DB, tenantID, createdBy uint64, req *dto.DeptCreateReq) (*entity.SysDept, error) {
	if req.ParentID > 0 {
		if _, err := s.deptRepo.GetByIDInTenant(db, req.ParentID, tenantID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, shared.ErrDeptNotFound
			}
			return nil, err
		}
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	dept := &entity.SysDept{
		TenantID:  tenantID,
		ParentID:  req.ParentID,
		DeptName:  req.DeptName,
		Sort:      req.Sort,
		Leader:    req.Leader,
		Phone:     req.Phone,
		Status:    status,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}

	if err := s.deptRepo.Create(db, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *DeptService) Update(db *gorm.DB, tenantID, id, updatedBy uint64, req *dto.DeptUpdateReq) error {
	dept, err := s.deptRepo.GetByIDInTenant(db, id, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrDeptNotFound
		}
		return err
	}

	if req.ParentID == id {
		return shared.ErrSameDeptParent
	}

	if req.ParentID > 0 {
		if _, err := s.deptRepo.GetByIDInTenant(db, req.ParentID, tenantID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return shared.ErrDeptNotFound
			}
			return err
		}
	}

	descendantIDs, _ := s.deptRepo.GetDescendantIDs(db, id, dept.TenantID)
	for _, did := range descendantIDs {
		if did == req.ParentID {
			return shared.ErrSameDeptParent
		}
	}

	parentChanged := req.ParentID != dept.ParentID

	dept.ParentID = req.ParentID
	dept.UpdatedBy = updatedBy
	if req.DeptName != "" {
		dept.DeptName = req.DeptName
	}
	dept.Leader = req.Leader
	dept.Phone = req.Phone
	dept.Sort = req.Sort
	dept.Status = req.Status

	if parentChanged {
		if req.ParentID > 0 {
			parent, err := s.deptRepo.GetByIDInTenant(db, req.ParentID, tenantID)
			if err != nil {
				return shared.ErrDeptNotFound
			}
			dept.Ancestors = parent.Ancestors + "," + fmt.Sprintf("%d", parent.ID)
		} else {
			dept.Ancestors = "0"
		}
	}

	if err := s.deptRepo.Update(db, dept); err != nil {
		return err
	}

	if parentChanged {
		allIDs := append([]uint64{id}, descendantIDs...)
		if err := s.rebuildAncestors(db, descendantIDs, dept.TenantID); err != nil {
			return err
		}
		if err := s.deptRepo.RebuildClosureForSubtree(db, allIDs, dept.TenantID); err != nil {
			return err
		}
	}

	return nil
}

func (s *DeptService) Delete(db *gorm.DB, id, tenantID uint64) error {
	dept, err := s.deptRepo.GetByIDInTenant(db, id, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrDeptNotFound
		}
		return err
	}

	count, err := s.deptRepo.CountUsersByDeptIDInTenant(db, id, tenantID)
	if err != nil {
		return err
	}
	if count > 0 {
		return shared.ErrDeptHasUsers
	}

	descendantIDs, _ := s.deptRepo.GetDescendantIDs(db, id, dept.TenantID)
	if len(descendantIDs) > 0 {
		return fmt.Errorf("部门下存在子部门，无法删除")
	}

	return s.deptRepo.Delete(db, id, tenantID)
}

func (s *DeptService) GetTree(c *gin.Context, db *gorm.DB, tenantID uint64) ([]dto.DeptResp, error) {
	depts, err := s.deptRepo.ListByTenantID(middleware.ApplyDeptScope(c, db), tenantID)
	if err != nil {
		return nil, err
	}

	flat := make([]dto.DeptResp, 0, len(depts))
	for _, d := range depts {
		flat = append(flat, shared.DeptToResp(&d))
	}

	return shared.BuildDeptTree(flat, 0), nil
}

func (s *DeptService) rebuildAncestors(db *gorm.DB, descendantIDs []uint64, tenantID uint64) error {
	for _, did := range descendantIDs {
		dept, err := s.deptRepo.GetByID(db, did)
		if err != nil {
			continue
		}
		if dept.ParentID > 0 {
			parent, err := s.deptRepo.GetByID(db, dept.ParentID)
			if err != nil {
				continue
			}
			dept.Ancestors = parent.Ancestors + "," + fmt.Sprintf("%d", parent.ID)
		} else {
			dept.Ancestors = "0"
		}
		_ = s.deptRepo.Update(db, dept)
	}
	return nil
}
