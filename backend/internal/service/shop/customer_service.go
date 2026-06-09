package shop

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"platform/internal/middleware"
	"platform/internal/model/dto"
	"platform/internal/model/entity"
	"platform/internal/repository/shop"
	"platform/internal/service/shared"
)

type ShopCustomerService struct {
	repo shop.ShopCustomerRepository
}

func NewShopCustomerService(repo shop.ShopCustomerRepository) *ShopCustomerService {
	return &ShopCustomerService{repo: repo}
}

func (s *ShopCustomerService) List(c *gin.Context, db *gorm.DB, tenantID uint64, req *dto.ShopCustomerListReq) ([]dto.ShopCustomerResp, int64, error) {
	q := db.Model(&entity.ShopCustomer{}).Where("tenant_id = ?", tenantID)
	q = middleware.ApplyUserScope(c, q)
	if req.CustomerName != "" {
		q = q.Where("customer_name LIKE ?", "%"+req.CustomerName+"%")
	}
	if req.ContactPerson != "" {
		q = q.Where("contact_person LIKE ?", "%"+req.ContactPerson+"%")
	}
	if req.CustomerType != nil {
		q = q.Where("customer_type = ?", *req.CustomerType)
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
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

	var customers []entity.ShopCustomer
	if err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&customers).Error; err != nil {
		return nil, 0, err
	}

	nameMap := s.fetchUserNames(db, collectCreatedBy(customers))

	list := make([]dto.ShopCustomerResp, 0, len(customers))
	for i := range customers {
		resp := shopCustomerToResp(&customers[i])
		resp.CreatedByName = nameMap[customers[i].CreatedBy]
		list = append(list, resp)
	}
	return list, total, nil
}

func (s *ShopCustomerService) Get(c *gin.Context, db *gorm.DB, id, tenantID uint64) (*dto.ShopCustomerResp, error) {
	q := db.Model(&entity.ShopCustomer{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var customer entity.ShopCustomer
	if err := q.First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrShopCustomerNotFound
		}
		return nil, err
	}
	resp := shopCustomerToResp(&customer)
	if customer.CreatedBy != 0 {
		nameMap := s.fetchUserNames(db, []uint64{customer.CreatedBy})
		resp.CreatedByName = nameMap[customer.CreatedBy]
	}
	return &resp, nil
}

func (s *ShopCustomerService) Create(c *gin.Context, db *gorm.DB, tenantID, createdBy uint64, req *dto.ShopCustomerCreateReq) (*dto.ShopCustomerResp, error) {
	_ = c

	count, err := s.repo.CountByNameAndType(db, tenantID, req.CustomerName, req.CustomerType, 0)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, shared.ErrShopCustomerNameExists
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	customer := &entity.ShopCustomer{
		TenantID:      tenantID,
		CustomerName:  req.CustomerName,
		CustomerType:  req.CustomerType,
		ContactPerson: req.ContactPerson,
		ContactPhone:  req.ContactPhone,
		ContactEmail:  req.ContactEmail,
		Address:       req.Address,
		Remark:        req.Remark,
		Status:        status,
		CreatedBy:     createdBy,
		UpdatedBy:     createdBy,
	}

	if err := s.repo.Create(db, customer); err != nil {
		return nil, err
	}

	resp := shopCustomerToResp(customer)
	if createdBy != 0 {
		nameMap := s.fetchUserNames(db, []uint64{createdBy})
		resp.CreatedByName = nameMap[createdBy]
	}
	return &resp, nil
}

func (s *ShopCustomerService) Update(c *gin.Context, db *gorm.DB, id, tenantID, updatedBy uint64, req *dto.ShopCustomerUpdateReq) error {
	q := db.Model(&entity.ShopCustomer{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var customer entity.ShopCustomer
	if err := q.First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopCustomerNotFound
		}
		return err
	}

	if customer.CustomerName != req.CustomerName || customer.CustomerType != req.CustomerType {
		count, err := s.repo.CountByNameAndType(db, tenantID, req.CustomerName, req.CustomerType, id)
		if err != nil {
			return err
		}
		if count > 0 {
			return shared.ErrShopCustomerNameExists
		}
	}

	customer.CustomerName = req.CustomerName
	customer.CustomerType = req.CustomerType
	customer.ContactPerson = req.ContactPerson
	customer.ContactPhone = req.ContactPhone
	customer.ContactEmail = req.ContactEmail
	customer.Address = req.Address
	customer.Remark = req.Remark
	if req.Status != 0 {
		customer.Status = req.Status
	}
	customer.UpdatedBy = updatedBy

	return s.repo.Update(db, &customer)
}

func (s *ShopCustomerService) Delete(c *gin.Context, db *gorm.DB, id, tenantID uint64) error {
	q := db.Model(&entity.ShopCustomer{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var customer entity.ShopCustomer
	if err := q.First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopCustomerNotFound
		}
		return err
	}

	hasOrders, err := s.repo.HasOrders(db, id)
	if err != nil {
		return err
	}
	if hasOrders {
		return shared.ErrShopCustomerHasOrders
	}

	return s.repo.Delete(db, id)
}

func (s *ShopCustomerService) ListOrders(c *gin.Context, db *gorm.DB, tenantID, customerID uint64) ([]dto.ShopCustomerOrderResp, error) {
	q := db.Model(&entity.ShopCustomer{}).Where("id = ? AND tenant_id = ?", customerID, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var customer entity.ShopCustomer
	if err := q.First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrShopCustomerNotFound
		}
		return nil, err
	}

	return s.repo.ListOrders(db, tenantID, customerID)
}

func (s *ShopCustomerService) Export(c *gin.Context, db *gorm.DB, tenantID uint64, req *dto.ShopCustomerListReq) ([]dto.ShopCustomerResp, error) {
	q := db.Model(&entity.ShopCustomer{}).Where("tenant_id = ?", tenantID)
	q = middleware.ApplyUserScope(c, q)
	if req.CustomerName != "" {
		q = q.Where("customer_name LIKE ?", "%"+req.CustomerName+"%")
	}
	if req.ContactPerson != "" {
		q = q.Where("contact_person LIKE ?", "%"+req.ContactPerson+"%")
	}
	if req.CustomerType != nil {
		q = q.Where("customer_type = ?", *req.CustomerType)
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}

	var customers []entity.ShopCustomer
	if err := q.Order("id DESC").Find(&customers).Error; err != nil {
		return nil, err
	}

	nameMap := s.fetchUserNames(db, collectCreatedBy(customers))

	list := make([]dto.ShopCustomerResp, 0, len(customers))
	for i := range customers {
		resp := shopCustomerToResp(&customers[i])
		resp.CreatedByName = nameMap[customers[i].CreatedBy]
		list = append(list, resp)
	}
	return list, nil
}

func (s *ShopCustomerService) fetchUserNames(db *gorm.DB, ids []uint64) map[uint64]string {
	result := make(map[uint64]string, len(ids))
	if len(ids) == 0 {
		return result
	}
	type row struct {
		ID       uint64
		RealName string
	}
	var rows []row
	if err := db.Table("sys_user").
		Select("id, real_name").
		Where("id IN ?", ids).
		Scan(&rows).Error; err != nil {
		return result
	}
	for _, r := range rows {
		result[r.ID] = r.RealName
	}
	return result
}

func collectCreatedBy(customers []entity.ShopCustomer) []uint64 {
	idSet := make(map[uint64]struct{}, len(customers))
	ids := make([]uint64, 0, len(customers))
	for _, c := range customers {
		if c.CreatedBy == 0 {
			continue
		}
		if _, ok := idSet[c.CreatedBy]; ok {
			continue
		}
		idSet[c.CreatedBy] = struct{}{}
		ids = append(ids, c.CreatedBy)
	}
	return ids
}

func shopCustomerToResp(c *entity.ShopCustomer) dto.ShopCustomerResp {
	return dto.ShopCustomerResp{
		ID:            c.ID,
		CustomerName:  c.CustomerName,
		CustomerType:  c.CustomerType,
		ContactPerson: c.ContactPerson,
		ContactPhone:  c.ContactPhone,
		ContactEmail:  c.ContactEmail,
		Address:       c.Address,
		Remark:        c.Remark,
		Status:        c.Status,
		CreatedAt:     c.CreatedAt,
		CreatedBy:     c.CreatedBy,
		UpdatedAt:     c.UpdatedAt,
		UpdatedBy:     c.UpdatedBy,
	}
}
