package shop

import (
	"platform/internal/model/dto"
	"platform/internal/model/entity"

	"gorm.io/gorm"
)

type RecordRepository interface {
	List(db *gorm.DB, tenantID uint64, req *dto.FinanceRecordListReq) ([]entity.FinanceRecord, int64, error)
	GetByID(db *gorm.DB, id uint64) (*entity.FinanceRecord, error)
	CountByRecordNoPrefix(db *gorm.DB, tenantID uint64, prefix string) (int64, error)
	Create(db *gorm.DB, r *entity.FinanceRecord) error
	Update(db *gorm.DB, r *entity.FinanceRecord) error
	Delete(db *gorm.DB, id uint64) error
	ListAttachments(db *gorm.DB, recordID uint64) ([]entity.FinanceAttachment, error)
	GetAttachment(db *gorm.DB, id uint64) (*entity.FinanceAttachment, error)
	CreateAttachment(db *gorm.DB, a *entity.FinanceAttachment) error
}

type recordRepository struct{}

func NewRecordRepository() RecordRepository {
	return &recordRepository{}
}

func (r *recordRepository) List(db *gorm.DB, tenantID uint64, req *dto.FinanceRecordListReq) ([]entity.FinanceRecord, int64, error) {
	var records []entity.FinanceRecord
	var total int64

	q := db.Model(&entity.FinanceRecord{}).Where("tenant_id = ?", tenantID)
	if req.RecordNo != "" {
		q = q.Where("record_no LIKE ?", "%"+req.RecordNo+"%")
	}
	if req.AccountID != nil {
		q = q.Where("account_id = ?", *req.AccountID)
	}
	if req.CategoryID != nil {
		q = q.Where("category_id = ?", *req.CategoryID)
	}
	if req.RecordType != nil {
		q = q.Where("record_type = ?", *req.RecordType)
	}
	if req.ReviewStatus != nil {
		q = q.Where("review_status = ?", *req.ReviewStatus)
	}

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

	if err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (r *recordRepository) GetByID(db *gorm.DB, id uint64) (*entity.FinanceRecord, error) {
	var rec entity.FinanceRecord
	if err := db.First(&rec, id).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *recordRepository) CountByRecordNoPrefix(db *gorm.DB, tenantID uint64, prefix string) (int64, error) {
	var count int64
	if err := db.Model(&entity.FinanceRecord{}).
		Where("tenant_id = ? AND record_no LIKE ?", tenantID, prefix+"%").
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *recordRepository) Create(db *gorm.DB, rec *entity.FinanceRecord) error {
	return db.Create(rec).Error
}

func (r *recordRepository) Update(db *gorm.DB, rec *entity.FinanceRecord) error {
	return db.Save(rec).Error
}

func (r *recordRepository) Delete(db *gorm.DB, id uint64) error {
	return db.Delete(&entity.FinanceRecord{}, id).Error
}

func (r *recordRepository) ListAttachments(db *gorm.DB, recordID uint64) ([]entity.FinanceAttachment, error) {
	var atts []entity.FinanceAttachment
	if err := db.Where("finance_record_id = ? AND deleted_at IS NULL", recordID).
		Order("id DESC").Find(&atts).Error; err != nil {
		return nil, err
	}
	return atts, nil
}

func (r *recordRepository) GetAttachment(db *gorm.DB, id uint64) (*entity.FinanceAttachment, error) {
	var att entity.FinanceAttachment
	if err := db.Where("id = ?", id).First(&att).Error; err != nil {
		return nil, err
	}
	return &att, nil
}

func (r *recordRepository) CreateAttachment(db *gorm.DB, a *entity.FinanceAttachment) error {
	return db.Create(a).Error
}