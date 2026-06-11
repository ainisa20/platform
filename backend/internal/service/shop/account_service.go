package shop

import (
	"errors"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"platform/internal/model/dto"
	"platform/internal/model/entity"
	shoprepo "platform/internal/repository/shop"
	"platform/internal/service/shared"
)

type ShopFinAccountService struct {
	repo     shoprepo.ShopFinAccountRepository
	userRepo shoprepo.UserRepository
}

func NewShopFinAccountService(repo shoprepo.ShopFinAccountRepository, userRepo shoprepo.UserRepository) *ShopFinAccountService {
	return &ShopFinAccountService{repo: repo, userRepo: userRepo}
}

func (s *ShopFinAccountService) List(db *gorm.DB, tenantID uint64, req *dto.ShopFinAccountListReq) ([]dto.ShopFinAccountResp, int64, error) {
	q := db.Model(&entity.ShopFinanceAccount{}).Where("tenant_id = ?", tenantID)
	if req.AccountName != "" {
		q = q.Where("account_name LIKE ?", "%"+req.AccountName+"%")
	}
	if req.AccountType != nil {
		q = q.Where("account_type = ?", *req.AccountType)
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

	var accounts []entity.ShopFinanceAccount
	if err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&accounts).Error; err != nil {
		return nil, 0, err
	}

	nameMap := s.fetchUserNames(db, collectFinAccountCreatedBy(accounts))

	balanceMap := s.calculateBalances(db, collectAccountIDs(accounts))

	list := make([]dto.ShopFinAccountResp, 0, len(accounts))
	for i := range accounts {
		resp := accountToResp(&accounts[i])
		resp.CreatedByName = nameMap[accounts[i].CreatedBy]
		resp.Balance = balanceMap[accounts[i].ID]
		list = append(list, resp)
	}
	return list, total, nil
}

func (s *ShopFinAccountService) Get(db *gorm.DB, id, tenantID uint64) (*dto.ShopFinAccountResp, error) {
	var a entity.ShopFinanceAccount
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrShopFinAccountNotFound
		}
		return nil, err
	}
	resp := accountToResp(&a)
	balanceMap := s.calculateBalances(db, []uint64{a.ID})
	resp.Balance = balanceMap[a.ID]
	if a.CreatedBy != 0 {
		nameMap := s.fetchUserNames(db, []uint64{a.CreatedBy})
		resp.CreatedByName = nameMap[a.CreatedBy]
	}
	return &resp, nil
}

func (s *ShopFinAccountService) Create(db *gorm.DB, tenantID, createdBy uint64, req *dto.ShopFinAccountCreateReq) (*dto.ShopFinAccountResp, error) {
	count, err := s.repo.CountByName(db, tenantID, req.AccountName, 0)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, shared.ErrShopFinAccountNameExists
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	config := req.Config
	if len(config) == 0 {
		config = datatypes.JSON("{}")
	}

	account := &entity.ShopFinanceAccount{
		TenantID:       tenantID,
		AccountName:    req.AccountName,
		AccountType:    req.AccountType,
		AccountNo:      req.AccountNo,
		InitialBalance: req.InitialBalance,
		Config:         config,
		Status:         status,
		CreatedBy:      createdBy,
		UpdatedBy:      createdBy,
	}

	if err := s.repo.Create(db, account); err != nil {
		return nil, err
	}

	resp := accountToResp(account)
	if createdBy != 0 {
		nameMap := s.fetchUserNames(db, []uint64{createdBy})
		resp.CreatedByName = nameMap[createdBy]
	}
	return &resp, nil
}

func (s *ShopFinAccountService) Update(db *gorm.DB, tenantID, id, updatedBy uint64, req *dto.ShopFinAccountUpdateReq) error {
	var account entity.ShopFinanceAccount
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopFinAccountNotFound
		}
		return err
	}

	if account.AccountName != req.AccountName {
		count, err := s.repo.CountByName(db, tenantID, req.AccountName, id)
		if err != nil {
			return err
		}
		if count > 0 {
			return shared.ErrShopFinAccountNameExists
		}
	}

	account.AccountName = req.AccountName
	account.AccountNo = req.AccountNo
	account.Config = req.Config
	if req.Status != 0 {
		account.Status = req.Status
	}
	account.UpdatedBy = updatedBy

	return s.repo.Update(db, &account)
}

func (s *ShopFinAccountService) Delete(db *gorm.DB, tenantID, id uint64) error {
	var account entity.ShopFinanceAccount
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrShopFinAccountNotFound
		}
		return err
	}

	return s.repo.Delete(db, id)
}

func (s *ShopFinAccountService) fetchUserNames(db *gorm.DB, ids []uint64) map[uint64]string {
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

func collectFinAccountCreatedBy(accounts []entity.ShopFinanceAccount) []uint64 {
	idSet := make(map[uint64]struct{}, len(accounts))
	ids := make([]uint64, 0, len(accounts))
	for _, a := range accounts {
		if a.CreatedBy == 0 {
			continue
		}
		if _, ok := idSet[a.CreatedBy]; ok {
			continue
		}
		idSet[a.CreatedBy] = struct{}{}
		ids = append(ids, a.CreatedBy)
	}
	return ids
}

func collectAccountIDs(accounts []entity.ShopFinanceAccount) []uint64 {
	ids := make([]uint64, len(accounts))
	for i, a := range accounts {
		ids[i] = a.ID
	}
	return ids
}

func (s *ShopFinAccountService) calculateBalances(db *gorm.DB, accountIDs []uint64) map[uint64]float64 {
	result := make(map[uint64]float64, len(accountIDs))
	if len(accountIDs) == 0 {
		return result
	}

	type balanceRow struct {
		AccountID uint64
		Income    float64
		Expense   float64
	}
	var rows []balanceRow
	db.Table("finance_record").
		Select("account_id, COALESCE(SUM(CASE WHEN record_type = 1 THEN actual_amount ELSE 0 END), 0) as income, COALESCE(SUM(CASE WHEN record_type = 2 THEN actual_amount ELSE 0 END), 0) as expense").
		Where("account_id IN ? AND review_status = 2 AND deleted_at IS NULL", accountIDs).
		Group("account_id").
		Scan(&rows)

	incomeExpense := make(map[uint64]struct{ income, expense float64 }, len(rows))
	for _, r := range rows {
		incomeExpense[r.AccountID] = struct{ income, expense float64 }{r.Income, r.Expense}
	}

	var accounts []entity.ShopFinanceAccount
	db.Where("id IN ?", accountIDs).Select("id, initial_balance").Find(&accounts)

	for _, a := range accounts {
		ie := incomeExpense[a.ID]
		result[a.ID] = a.InitialBalance + ie.income - ie.expense
	}
	return result
}

func accountToResp(a *entity.ShopFinanceAccount) dto.ShopFinAccountResp {
	return dto.ShopFinAccountResp{
		ID:             a.ID,
		AccountName:    a.AccountName,
		AccountType:    a.AccountType,
		AccountNo:      a.AccountNo,
		InitialBalance: a.InitialBalance,
		Config:         a.Config,
		Status:         a.Status,
		CreatedAt:      a.CreatedAt,
		CreatedBy:      a.CreatedBy,
		UpdatedAt:      a.UpdatedAt,
	}
}
