package shop

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"platform/internal/middleware"
	"platform/internal/model/dto"
	"platform/internal/model/entity"
	"platform/internal/pkg/storage"
	"platform/internal/repository/shop"
	"platform/internal/service/shared"
)

const (
	financeRecordReviewStatusPending  int16 = 1
	financeRecordReviewStatusApproved int16 = 2
	financeRecordReviewStatusRejected int16 = 3
)

type RecordService struct {
	repo         shop.RecordRepository
	accountRepo  shop.ShopFinAccountRepository
	categoryRepo shop.ShopFinCategoryRepository
	orderRepo    shop.OrderRepository
	userRepo     shop.UserRepository
	storage      *storage.MinIOStorage
}

func NewRecordService(
	repo shop.RecordRepository,
	accountRepo shop.ShopFinAccountRepository,
	categoryRepo shop.ShopFinCategoryRepository,
	orderRepo shop.OrderRepository,
	userRepo shop.UserRepository,
	minioStorage *storage.MinIOStorage,
) *RecordService {
	return &RecordService{
		repo:         repo,
		accountRepo:  accountRepo,
		categoryRepo: categoryRepo,
		orderRepo:    orderRepo,
		userRepo:     userRepo,
		storage:      minioStorage,
	}
}

func (s *RecordService) List(c *gin.Context, db *gorm.DB, tenantID uint64, req *dto.FinanceRecordListReq) ([]dto.FinanceRecordResp, int64, error) {
	q := db.Model(&entity.FinanceRecord{}).Where("tenant_id = ?", tenantID)
	q = middleware.ApplyUserScope(c, q)
	if req.RecordNo != "" {
		q = q.Where("record_no LIKE ?", "%"+req.RecordNo+"%")
	}
	if req.AccountID != nil {
		q = q.Where("account_id = ?", *req.AccountID)
	}
	if req.CategoryID != nil {
		q = q.Where("category_id = ?", *req.CategoryID)
	}
	if req.CategoryL1 != nil {
		q = q.Where("category_l1 = ?", *req.CategoryL1)
	}
	if req.CategoryL2 != nil {
		q = q.Where("category_l2 = ?", *req.CategoryL2)
	}
	if req.CategoryL3 != nil {
		q = q.Where("category_l3 = ?", *req.CategoryL3)
	}
	if req.RecordType != nil {
		q = q.Where("record_type = ?", *req.RecordType)
	}
	if req.ReviewStatus != nil {
		q = q.Where("review_status = ?", *req.ReviewStatus)
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

	var records []entity.FinanceRecord
	if err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

	nameMap := s.fetchUserNames(db, collectRecordUserIDs(records))

	list := make([]dto.FinanceRecordResp, 0, len(records))
	for i := range records {
		resp := recordToResp(&records[i])
		resp.CreatedByName = nameMap[records[i].CreatedBy]
		resp.ReviewByName = nameMap[records[i].ReviewBy]
		list = append(list, resp)
	}
	return list, total, nil
}

func (s *RecordService) Get(c *gin.Context, db *gorm.DB, id, tenantID uint64) (*dto.FinanceRecordResp, error) {
	q := db.Model(&entity.FinanceRecord{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var rec entity.FinanceRecord
	if err := q.First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrFinRecordNotFound
		}
		return nil, err
	}
	resp := recordToResp(&rec)
	ids := make([]uint64, 0, 2)
	if rec.CreatedBy != 0 {
		ids = append(ids, rec.CreatedBy)
	}
	if rec.ReviewBy != 0 {
		ids = append(ids, rec.ReviewBy)
	}
	if len(ids) > 0 {
		nameMap := s.fetchUserNames(db, ids)
		resp.CreatedByName = nameMap[rec.CreatedBy]
		resp.ReviewByName = nameMap[rec.ReviewBy]
	}
	return &resp, nil
}

func (s *RecordService) Create(c *gin.Context, db *gorm.DB, tenantID, createdBy uint64, req *dto.FinanceRecordCreateReq) (*dto.FinanceRecordResp, error) {
	_ = c
	account, err := s.accountRepo.GetByID(db, req.AccountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrFinRecordAccountInvalid
		}
		return nil, err
	}
	if account.TenantID != tenantID {
		return nil, shared.ErrFinRecordAccountInvalid
	}

	category, err := s.categoryRepo.GetByID(db, req.CategoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrFinRecordCategoryInvalid
		}
		return nil, err
	}
	if category.TenantID != tenantID {
		return nil, shared.ErrFinRecordCategoryInvalid
	}
	if category.Level != 3 {
		return nil, shared.ErrFinRecordCategoryNotLeaf
	}

	categoryPath, levels, err := s.buildCategoryPath(db, category)
	if err != nil {
		return nil, err
	}

	if req.OrderGroupID != nil {
		group, err := s.orderRepo.GetGroup(db, *req.OrderGroupID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, shared.ErrFinRecordOrderInvalid
			}
			return nil, err
		}
		if group.TenantID != tenantID {
			return nil, shared.ErrFinRecordOrderInvalid
		}
	}

	recordDate, err := parseDate(req.RecordDate)
	if err != nil {
		return nil, err
	}

	recordNo := s.generateRecordNo(db, tenantID)

	rec := &entity.FinanceRecord{
		TenantID:              tenantID,
		RecordNo:              recordNo,
		AccountID:             account.ID,
		AccountName:           account.AccountName,
		AccountType:           account.AccountType,
		AccountInitialBalance: account.InitialBalance,
		CategoryID:            category.ID,
		CategoryName:          category.CategoryName,
		CategoryPath:          categoryPath,
		CategoryL1:            levels[0],
		CategoryL2:            levels[1],
		CategoryL3:            levels[2],
		RecordType:            req.RecordType,
		Amount:                req.Amount,
		ActualAmount:          0,
		OrderGroupID:          req.OrderGroupID,
		ReviewStatus:          financeRecordReviewStatusPending,
		RecordDate:            recordDate,
		Remark:                req.Remark,
		CreatedBy:             createdBy,
		UpdatedBy:             createdBy,
	}

	if err := s.repo.Create(db, rec); err != nil {
		return nil, err
	}

	resp := recordToResp(rec)
	if createdBy != 0 {
		nameMap := s.fetchUserNames(db, []uint64{createdBy})
		resp.CreatedByName = nameMap[createdBy]
	}
	return &resp, nil
}

func (s *RecordService) Update(c *gin.Context, db *gorm.DB, tenantID, id, updatedBy uint64, req *dto.FinanceRecordUpdateReq) error {
	q := db.Model(&entity.FinanceRecord{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var rec entity.FinanceRecord
	if err := q.First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrFinRecordNotFound
		}
		return err
	}

	if rec.ReviewStatus == financeRecordReviewStatusApproved {
		return shared.ErrFinRecordApproved
	}

	account, err := s.accountRepo.GetByID(db, req.AccountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrFinRecordAccountInvalid
		}
		return err
	}
	if account.TenantID != tenantID {
		return shared.ErrFinRecordAccountInvalid
	}

	category, err := s.categoryRepo.GetByID(db, req.CategoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrFinRecordCategoryInvalid
		}
		return err
	}
	if category.TenantID != tenantID {
		return shared.ErrFinRecordCategoryInvalid
	}
	if category.Level != 3 {
		return shared.ErrFinRecordCategoryNotLeaf
	}

	categoryPath, levels, err := s.buildCategoryPath(db, category)
	if err != nil {
		return err
	}

	if req.OrderGroupID != nil {
		group, err := s.orderRepo.GetGroup(db, *req.OrderGroupID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return shared.ErrFinRecordOrderInvalid
			}
			return err
		}
		if group.TenantID != tenantID {
			return shared.ErrFinRecordOrderInvalid
		}
	}

	recordDate, err := parseDate(req.RecordDate)
	if err != nil {
		return err
	}

	rec.AccountID = account.ID
	rec.CategoryID = category.ID
	rec.CategoryName = category.CategoryName
	rec.CategoryPath = categoryPath
	rec.CategoryL1 = levels[0]
	rec.CategoryL2 = levels[1]
	rec.CategoryL3 = levels[2]
	rec.RecordType = req.RecordType
	rec.Amount = req.Amount
	rec.OrderGroupID = req.OrderGroupID
	rec.RecordDate = recordDate
	rec.Remark = req.Remark
	rec.UpdatedBy = updatedBy

	if rec.ReviewStatus == financeRecordReviewStatusRejected {
		rec.ReviewStatus = financeRecordReviewStatusPending
		rec.ReviewBy = 0
		rec.ReviewAt = nil
		rec.ReviewNotes = ""
	}

	return s.repo.Update(db, &rec)
}

func (s *RecordService) Delete(c *gin.Context, db *gorm.DB, tenantID, id uint64) error {
	q := db.Model(&entity.FinanceRecord{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var rec entity.FinanceRecord
	if err := q.First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrFinRecordNotFound
		}
		return err
	}

	if rec.ReviewStatus == financeRecordReviewStatusApproved {
		return shared.ErrFinRecordApproved
	}

	return s.repo.Delete(db, id)
}

func (s *RecordService) Review(c *gin.Context, db *gorm.DB, tenantID, id, userID uint64, req *dto.FinanceReviewReq) error {
	q := db.Model(&entity.FinanceRecord{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var rec entity.FinanceRecord
	if err := q.First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrFinRecordNotFound
		}
		return err
	}

	if rec.CreatedBy == userID {
		return shared.ErrFinRecordReviewerIsCreator
	}

	if rec.ReviewStatus == financeRecordReviewStatusApproved {
		return shared.ErrFinRecordInvalidStatus
	}

	now := time.Now()
	switch req.Action {
	case "approve":
		if req.ActualAmount == nil || *req.ActualAmount <= 0 {
			return shared.ErrFinRecordActualAmountRequired
		}
		rec.ActualAmount = *req.ActualAmount
		rec.ReviewStatus = financeRecordReviewStatusApproved
	case "reject":
		rec.ActualAmount = 0
		rec.ReviewStatus = financeRecordReviewStatusRejected
	default:
		return shared.ErrFinRecordInvalidAction
	}

	rec.ReviewBy = userID
	rec.ReviewAt = &now
	rec.ReviewNotes = req.Notes
	rec.UpdatedBy = userID

	return s.repo.Update(db, &rec)
}

func (s *RecordService) ListAttachments(c *gin.Context, db *gorm.DB, tenantID, id uint64) ([]dto.FinanceAttachmentResp, error) {
	q := db.Model(&entity.FinanceRecord{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var rec entity.FinanceRecord
	if err := q.First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrFinRecordNotFound
		}
		return nil, err
	}

	atts, err := s.repo.ListAttachments(db, id)
	if err != nil {
		return nil, err
	}
	resps := make([]dto.FinanceAttachmentResp, 0, len(atts))
	for _, a := range atts {
		resps = append(resps, dto.FinanceAttachmentResp{
			ID:        a.ID,
			FileName:  a.FileName,
			FileSize:  a.FileSize,
			FileType:  a.FileType,
			CreatedAt: a.CreatedAt,
		})
	}
	return resps, nil
}

func (s *RecordService) CreateAttachment(c *gin.Context, db *gorm.DB, tenantID, id, userID uint64, fileHeader *multipart.FileHeader) (*dto.FinanceAttachmentResp, error) {
	q := db.Model(&entity.FinanceRecord{}).Where("id = ? AND tenant_id = ?", id, tenantID)
	q = middleware.ApplyUserScope(c, q)
	var rec entity.FinanceRecord
	if err := q.First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrFinRecordNotFound
		}
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Detect content type from first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file for content detection: %w", err)
	}
	contentType := http.DetectContentType(buf[:n])
	if ext := filepath.Ext(fileHeader.Filename); ext != "" {
		if mt := mime.TypeByExtension(ext); mt != "" {
			contentType = mt
		}
	}
	// Seek back to start for full upload
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	token := newAttachmentToken()
	ext := filepath.Ext(fileHeader.Filename)
	objectPath := fmt.Sprintf("finance-attachments/%d/%d/%s%s", tenantID, id, token, ext)

	bucketName := fmt.Sprintf("tenant-%d", tenantID)
	if err := s.storage.Upload(c.Request.Context(), bucketName, objectPath, file, fileHeader.Size, contentType); err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	a := &entity.FinanceAttachment{
		TenantID:        tenantID,
		FinanceRecordID: id,
		FileName:        fileHeader.Filename,
		FilePath:        fmt.Sprintf("%s/%s", bucketName, objectPath),
		FileSize:        fileHeader.Size,
		FileType:        contentType,
		CreatedBy:       userID,
	}
	if err := s.repo.CreateAttachment(db, a); err != nil {
		return nil, err
	}
	return &dto.FinanceAttachmentResp{
		ID:        a.ID,
		FileName:  a.FileName,
		FileSize:  a.FileSize,
		FileType:  a.FileType,
		CreatedAt: a.CreatedAt,
	}, nil
}

func (s *RecordService) GetAttachmentDownloadURL(c *gin.Context, db *gorm.DB, tenantID, recordID, attID uint64) (string, error) {
	q := db.Model(&entity.FinanceRecord{}).Where("id = ? AND tenant_id = ?", recordID, tenantID)
	var rec entity.FinanceRecord
	if err := q.First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", shared.ErrFinRecordNotFound
		}
		return "", err
	}

	att, err := s.repo.GetAttachment(db, attID)
	if err != nil {
		return "", err
	}
	if att.FinanceRecordID != recordID {
		return "", shared.ErrFinRecordNotFound
	}

	parts := strings.SplitN(att.FilePath, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid file path format")
	}
	bucket, objectName := parts[0], parts[1]

	return s.storage.GetDownloadURL(c.Request.Context(), bucket, objectName, time.Hour)
}

func (s *RecordService) Export(c *gin.Context, db *gorm.DB, tenantID uint64, req *dto.FinanceRecordListReq) ([]dto.FinanceRecordResp, error) {
	q := db.Model(&entity.FinanceRecord{}).Where("tenant_id = ?", tenantID)
	q = middleware.ApplyUserScope(c, q)
	if req.RecordNo != "" {
		q = q.Where("record_no LIKE ?", "%"+req.RecordNo+"%")
	}
	if req.AccountID != nil {
		q = q.Where("account_id = ?", *req.AccountID)
	}
	if req.CategoryID != nil {
		q = q.Where("category_id = ?", *req.CategoryID)
	}
	if req.CategoryL1 != nil {
		q = q.Where("category_l1 = ?", *req.CategoryL1)
	}
	if req.CategoryL2 != nil {
		q = q.Where("category_l2 = ?", *req.CategoryL2)
	}
	if req.CategoryL3 != nil {
		q = q.Where("category_l3 = ?", *req.CategoryL3)
	}
	if req.RecordType != nil {
		q = q.Where("record_type = ?", *req.RecordType)
	}
	if req.ReviewStatus != nil {
		q = q.Where("review_status = ?", *req.ReviewStatus)
	}

	var records []entity.FinanceRecord
	if err := q.Order("id DESC").Find(&records).Error; err != nil {
		return nil, err
	}

	nameMap := s.fetchUserNames(db, collectRecordUserIDs(records))

	list := make([]dto.FinanceRecordResp, 0, len(records))
	for i := range records {
		resp := recordToResp(&records[i])
		resp.CreatedByName = nameMap[records[i].CreatedBy]
		resp.ReviewByName = nameMap[records[i].ReviewBy]
		list = append(list, resp)
	}
	return list, nil
}

func (s *RecordService) generateRecordNo(db *gorm.DB, tenantID uint64) string {
	timestamp := time.Now().Format("20060102150405")
	prefix := "FIN" + timestamp
	count, _ := s.repo.CountByRecordNoPrefix(db, tenantID, prefix)
	seq := count + 1
	return fmt.Sprintf("%s%03d", prefix, seq)
}

func parseDate(dateStr string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid record_date: %w", err)
	}
	return t, nil
}

func (s *RecordService) fetchUserNames(db *gorm.DB, ids []uint64) map[uint64]string {
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

func collectRecordUserIDs(records []entity.FinanceRecord) []uint64 {
	idSet := make(map[uint64]struct{}, len(records)*2)
	ids := make([]uint64, 0, len(records)*2)
	for _, r := range records {
		if r.CreatedBy != 0 {
			if _, ok := idSet[r.CreatedBy]; !ok {
				idSet[r.CreatedBy] = struct{}{}
				ids = append(ids, r.CreatedBy)
			}
		}
		if r.ReviewBy != 0 {
			if _, ok := idSet[r.ReviewBy]; !ok {
				idSet[r.ReviewBy] = struct{}{}
				ids = append(ids, r.ReviewBy)
			}
		}
	}
	return ids
}

func recordToResp(r *entity.FinanceRecord) dto.FinanceRecordResp {
	dateStr := ""
	if !r.RecordDate.IsZero() {
		dateStr = r.RecordDate.Format("2006-01-02")
	}
	return dto.FinanceRecordResp{
		ID:                    r.ID,
		RecordNo:              r.RecordNo,
		AccountID:             r.AccountID,
		AccountName:           r.AccountName,
		AccountType:           r.AccountType,
		AccountInitialBalance: r.AccountInitialBalance,
		CategoryID:            r.CategoryID,
		CategoryName:          r.CategoryName,
		CategoryPath:          r.CategoryPath,
		CategoryL1:            r.CategoryL1,
		CategoryL2:            r.CategoryL2,
		CategoryL3:            r.CategoryL3,
		RecordType:            r.RecordType,
		Amount:                r.Amount,
		ActualAmount:          r.ActualAmount,
		OrderGroupID:          r.OrderGroupID,
		ReviewStatus:          r.ReviewStatus,
		ReviewBy:              r.ReviewBy,
		ReviewAt:              r.ReviewAt,
		ReviewNotes:           r.ReviewNotes,
		RecordDate:            dateStr,
		Remark:                r.Remark,
		CreatedAt:             r.CreatedAt,
		CreatedBy:             r.CreatedBy,
		UpdatedAt:             r.UpdatedAt,
	}
}

func newAttachmentToken() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func (s *RecordService) buildCategoryPath(db *gorm.DB, leaf *entity.ShopFinanceCategory) (string, [3]string, error) {
	if leaf.Level == 0 {
		return leaf.CategoryName, [3]string{leaf.CategoryName, "", ""}, nil
	}
	all, err := s.categoryRepo.ListSynced(db, leaf.TenantID, &dto.ShopFinCategoryListReq{})
	if err != nil {
		return "", [3]string{}, err
	}
	idMap := make(map[uint64]*entity.ShopFinanceCategory, len(all))
	for i := range all {
		idMap[all[i].ID] = &all[i]
	}
	chain := []string{leaf.CategoryName}
	current := leaf
	for current.ParentID != 0 {
		parent, ok := idMap[current.ParentID]
		if !ok {
			break
		}
		chain = append([]string{parent.CategoryName}, chain...)
		current = parent
	}

	var levels [3]string
	for i, name := range chain {
		if i < 3 {
			levels[i] = name
		}
	}
	path := strings.Join(chain, " / ")
	return path, levels, nil
}