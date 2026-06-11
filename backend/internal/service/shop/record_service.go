package shop

import (
	"archive/zip"
	"bytes"
	"context"
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
	"github.com/xuri/excelize/v2"
	"gorm.io/driver/postgres"
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
	dsn          string
}

func NewRecordService(
	repo shop.RecordRepository,
	accountRepo shop.ShopFinAccountRepository,
	categoryRepo shop.ShopFinCategoryRepository,
	orderRepo shop.OrderRepository,
	userRepo shop.UserRepository,
	minioStorage *storage.MinIOStorage,
	dsn string,
) *RecordService {
	return &RecordService{
		repo:         repo,
		accountRepo:  accountRepo,
		categoryRepo: categoryRepo,
		orderRepo:    orderRepo,
		userRepo:     userRepo,
		storage:      minioStorage,
		dsn:          dsn,
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

const exportMaxRecords = 1000

func (s *RecordService) CreateExportTask(db *gorm.DB, tenantID, userID uint64, req *dto.FinanceRecordListReq) (*entity.ExportTask, error) {
	q := db.Model(&entity.FinanceRecord{}).Where("tenant_id = ?", tenantID)
	s.applyExportFilters(q, req)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, errors.New("没有可导出的数据")
	}
	if total > exportMaxRecords {
		return nil, fmt.Errorf("导出数据量 %d 超过上限 %d，请缩小筛选范围", total, exportMaxRecords)
	}

	task := &entity.ExportTask{
		TenantID:   tenantID,
		TaskType:   "finance_record_zip",
		Status:     0,
		TotalCount: int(total),
		CreatedBy:  userID,
		FileName:   fmt.Sprintf("收支记录_%s.zip", time.Now().Format("20060102_150405")),
	}
	if err := db.Create(task).Error; err != nil {
		return nil, err
	}

	go s.runExportTask(task.ID, tenantID, req)

	return task, nil
}

func (s *RecordService) GetExportTask(db *gorm.DB, taskID, tenantID uint64) (*entity.ExportTask, error) {
	var task entity.ExportTask
	if err := db.Where("id = ? AND tenant_id = ?", taskID, tenantID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *RecordService) GetExportDownloadURL(db *gorm.DB, taskID, tenantID uint64) (string, error) {
	var task entity.ExportTask
	if err := db.Where("id = ? AND tenant_id = ? AND status = 1", taskID, tenantID).First(&task).Error; err != nil {
		return "", err
	}
	parts := strings.SplitN(task.FileKey, "/", 2)
	if len(parts) != 2 {
		return "", errors.New("invalid file key")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.storage.GetDownloadURL(ctx, parts[0], parts[1], time.Hour)
}

type ExportFileInfo struct {
	Reader   io.ReadCloser
	FileName string
	Size     int64
}

func (s *RecordService) DownloadExportFile(db *gorm.DB, taskID, tenantID uint64) (*ExportFileInfo, error) {
	var task entity.ExportTask
	if err := db.Where("id = ? AND tenant_id = ? AND status = 1", taskID, tenantID).First(&task).Error; err != nil {
		return nil, err
	}
	parts := strings.SplitN(task.FileKey, "/", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid file key")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	_ = cancel
	reader, err := s.storage.GetObject(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}
	return &ExportFileInfo{
		Reader:   reader,
		FileName: task.FileName,
	}, nil
}

func (s *RecordService) runExportTask(taskID, tenantID uint64, req *dto.FinanceRecordListReq) {
	db, err := gorm.Open(postgres.Open(s.dsn), &gorm.Config{})
	if err != nil {
		return
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	updateStatus := func(status int16, done int, fileKey, errMsg string) {
		now := time.Now()
		updates := map[string]interface{}{
			"status":       status,
			"done_count":   done,
			"error_msg":    errMsg,
			"completed_at": &now,
		}
		if fileKey != "" {
			updates["file_key"] = fileKey
		}
		db.Model(&entity.ExportTask{}).Where("id = ?", taskID).Updates(updates)
	}

	q := db.Model(&entity.FinanceRecord{}).Where("tenant_id = ?", tenantID)
	s.applyExportFilters(q, req)
	var records []entity.FinanceRecord
	if err := q.Order("id DESC").Find(&records).Error; err != nil {
		updateStatus(2, 0, "", "查询记录失败: "+err.Error())
		return
	}

	recordIDs := make([]uint64, len(records))
	for i, r := range records {
		recordIDs[i] = r.ID
	}
	attachments, _ := s.repo.ListAttachmentsByRecordIDs(db, recordIDs)
	attMap := make(map[uint64][]entity.FinanceAttachment)
	for _, a := range attachments {
		attMap[a.FinanceRecordID] = append(attMap[a.FinanceRecordID], a)
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	excelData, err := s.generateExportExcel(records)
	if err != nil {
		updateStatus(2, 0, "", "生成Excel失败: "+err.Error())
		return
	}
	w, _ := zipWriter.Create("收支记录.xlsx")
	w.Write(excelData)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	doneCount := 0
	for _, rec := range records {
		atts := attMap[rec.ID]
		if len(atts) > 0 {
			dirName := filepath.Join("附件", rec.RecordNo)
			for _, att := range atts {
				fileParts := strings.SplitN(att.FilePath, "/", 2)
				if len(fileParts) != 2 {
					continue
				}
				obj, err := s.storage.GetObject(ctx, fileParts[0], fileParts[1])
				if err != nil {
					continue
				}
				aw, err := zipWriter.Create(filepath.Join(dirName, att.FileName))
				if err != nil {
					obj.Close()
					continue
				}
				io.Copy(aw, obj)
				obj.Close()
			}
		}
		doneCount++
		if doneCount%10 == 0 {
			db.Model(&entity.ExportTask{}).Where("id = ?", taskID).Update("done_count", doneCount)
		}
	}
	zipWriter.Close()

	bucketName := fmt.Sprintf("tenant-%d", tenantID)
	zipKey := fmt.Sprintf("exports/%d/%d.zip", tenantID, taskID)
	reader := bytes.NewReader(buf.Bytes())
	if err := s.storage.Upload(ctx, bucketName, zipKey, reader, int64(buf.Len()), "application/zip"); err != nil {
		updateStatus(2, doneCount, "", "上传失败: "+err.Error())
		return
	}

	fileKey := fmt.Sprintf("%s/%s", bucketName, zipKey)
	now := time.Now()
	db.Model(&entity.ExportTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":       1,
		"done_count":   doneCount,
		"file_key":     fileKey,
		"completed_at": &now,
	})
}

func (s *RecordService) generateExportExcel(records []entity.FinanceRecord) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "收支记录"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"记录号", "记账日期", "账户", "账户类型", "一级分类", "二级分类", "三级分类", "收支", "金额", "实际金额", "审核状态", "备注"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	statusMap := map[int16]string{1: "待审核", 2: "已通过", 3: "已驳回"}
	typeMap := map[int16]string{1: "收入", 2: "支出"}
	acctTypeMap := map[int16]string{1: "对公", 2: "对私"}

	for i, r := range records {
		row := i + 2
		dateStr := ""
		if !r.RecordDate.IsZero() {
			dateStr = r.RecordDate.Format("2006-01-02")
		}
		vals := []interface{}{
			r.RecordNo, dateStr, r.AccountName, acctTypeMap[r.AccountType],
			r.CategoryL1, r.CategoryL2, r.CategoryL3,
			typeMap[r.RecordType], r.Amount, r.ActualAmount,
			statusMap[r.ReviewStatus], r.Remark,
		}
		for j, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

	buf, err := f.WriteToBuffer()
	return buf.Bytes(), err
}

func (s *RecordService) applyExportFilters(q *gorm.DB, req *dto.FinanceRecordListReq) {
	if req.RecordNo != "" {
		q = q.Where("record_no LIKE ?", "%"+req.RecordNo+"%")
	}
	if req.AccountID != nil {
		q = q.Where("account_id = ?", *req.AccountID)
	}
	if req.AccountType != nil {
		q = q.Where("account_type = ?", *req.AccountType)
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
	if req.RecordDateStart != "" {
		q = q.Where("record_date >= ?", req.RecordDateStart)
	}
	if req.RecordDateEnd != "" {
		q = q.Where("record_date <= ?", req.RecordDateEnd)
	}
	if req.CreatedBy != nil {
		q = q.Where("created_by = ?", *req.CreatedBy)
	}
}