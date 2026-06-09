package entity

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy uint64         `gorm:"index" json:"created_by"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy uint64         `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type SysUser struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     uint64         `gorm:"not null;default:0;index" json:"tenant_id"`
	DeptID       *uint64        `gorm:"index" json:"dept_id"`
	Username     string         `gorm:"type:varchar(64);not null" json:"username"`
	Password     string         `gorm:"type:varchar(255);not null" json:"-"`
	RealName     string         `gorm:"type:varchar(64);not null" json:"real_name"`
	Phone        string         `gorm:"type:varchar(20)" json:"phone"`
	Email        string         `gorm:"type:varchar(128)" json:"email"`
	Avatar       string         `gorm:"type:varchar(500)" json:"avatar"`
	Status       int16          `gorm:"default:1" json:"status"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	LastLoginIP  string         `gorm:"type:varchar(45)" json:"last_login_ip"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy    uint64         `json:"created_by"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy    uint64         `json:"updated_by"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (SysUser) TableName() string { return "sys_user" }

type SysRole struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  uint64         `gorm:"not null;default:0;index" json:"tenant_id"`
	RoleName  string         `gorm:"type:varchar(64);not null" json:"role_name"`
	RoleCode  string         `gorm:"type:varchar(64);not null" json:"role_code"`
	DataScope int16          `gorm:"default:1" json:"data_scope"`
	Sort      int16          `gorm:"default:0" json:"sort"`
	Status    int16          `gorm:"default:1" json:"status"`
	Remark    string         `gorm:"type:varchar(255)" json:"remark"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy uint64         `json:"created_by"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy uint64         `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (SysRole) TableName() string { return "sys_role" }

type SysPermission struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentID   uint64     `gorm:"default:0" json:"parent_id"`
	SystemType string     `gorm:"type:varchar(16);not null" json:"system_type"`
	Name       string     `gorm:"type:varchar(64);not null" json:"name"`
	Type       int16      `gorm:"not null" json:"type"`
	Path       string     `gorm:"type:varchar(255)" json:"path"`
	Component  string     `gorm:"type:varchar(255)" json:"component"`
	PermsCode  string     `gorm:"type:varchar(100)" json:"perms_code"`
	Icon       string     `gorm:"type:varchar(64)" json:"icon"`
	Sort       int16      `gorm:"default:0" json:"sort"`
	Visible    bool       `gorm:"default:true" json:"visible"`
	Status     int16      `gorm:"default:1" json:"status"`
	AutoSynced bool       `gorm:"default:true" json:"auto_synced"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SysPermission) TableName() string { return "sys_permission" }

type SysRolePermission struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     uint64 `gorm:"not null;index" json:"tenant_id"`
	RoleID       uint64 `gorm:"not null" json:"role_id"`
	PermissionID uint64 `gorm:"not null" json:"permission_id"`
}

func (SysRolePermission) TableName() string { return "sys_role_permission" }

type SysUserRole struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID uint64 `gorm:"not null" json:"user_id"`
	RoleID uint64 `gorm:"not null" json:"role_id"`
}

func (SysUserRole) TableName() string { return "sys_user_role" }

type SysDept struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  uint64         `gorm:"not null;default:0;index" json:"tenant_id"`
	ParentID  uint64         `gorm:"default:0" json:"parent_id"`
	Ancestors string         `gorm:"type:varchar(500)" json:"ancestors"`
	DeptName  string         `gorm:"type:varchar(64);not null" json:"dept_name"`
	Sort      int16          `gorm:"default:0" json:"sort"`
	Leader    string         `gorm:"type:varchar(64)" json:"leader"`
	Phone     string         `gorm:"type:varchar(20)" json:"phone"`
	Status    int16          `gorm:"default:1" json:"status"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy uint64         `json:"created_by"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy uint64         `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (SysDept) TableName() string { return "sys_dept" }

type SysDeptClosure struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     uint64 `gorm:"not null;index" json:"tenant_id"`
	AncestorID   uint64 `gorm:"not null;index" json:"ancestor_id"`
	DescendantID uint64 `gorm:"not null;index" json:"descendant_id"`
	Depth        int16  `gorm:"not null;default:0" json:"depth"`
}

func (SysDeptClosure) TableName() string { return "sys_dept_closure" }

type SysOperationLog struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID   uint64    `gorm:"index" json:"tenant_id"`
	UserID     uint64    `json:"user_id"`
	Username   string    `gorm:"type:varchar(64)" json:"username"`
	Action     string    `gorm:"type:varchar(64)" json:"action"`
	Method     string    `gorm:"type:varchar(10)" json:"method"`
	URL        string    `gorm:"type:varchar(500)" json:"url"`
	Params     string    `gorm:"type:text" json:"params"`
	IP         string    `gorm:"type:varchar(45)" json:"ip"`
	DurationMs int       `json:"duration_ms"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (SysOperationLog) TableName() string { return "sys_operation_log" }

type SysShop struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopCode   string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"shop_code"`
	ShopName   string         `gorm:"type:varchar(64);not null" json:"shop_name"`
	Contact    string         `gorm:"type:varchar(64)" json:"contact"`
	Phone      string         `gorm:"type:varchar(20)" json:"phone"`
	Email      string         `gorm:"type:varchar(128)" json:"email"`
	Address    string         `gorm:"type:varchar(255)" json:"address"`
	Remark     string         `gorm:"type:varchar(500)" json:"remark"`
	Status     int16          `gorm:"default:1" json:"status"`
	AdminUserID *uint64       `json:"admin_user_id"`
	ExpiresAt  *time.Time     `json:"expires_at"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy  uint64         `json:"created_by"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy  uint64         `json:"updated_by"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (SysShop) TableName() string { return "sys_shop" }

// ========== Product (platform-level, no tenant_id) ==========

type ProductCategory struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryCode string         `gorm:"type:varchar(64)" json:"category_code"`
	CategoryName string         `gorm:"type:varchar(128);not null" json:"category_name"`
	Sort         int16          `gorm:"default:0" json:"sort"`
	Status       int16          `gorm:"default:1" json:"status"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy    uint64         `json:"created_by"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy    uint64         `json:"updated_by"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (ProductCategory) TableName() string { return "product_category" }

type Product struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductCode    string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"product_code"`
	ProductName    string         `gorm:"type:varchar(128);not null" json:"product_name"`
	CategoryID     *uint64        `gorm:"index" json:"category_id"`
	Price          float64        `gorm:"type:numeric(12,2);not null" json:"price"`
	Sort           int16          `gorm:"default:0" json:"sort"`
	Status         int16          `gorm:"default:1" json:"status"`
	MallProductCode string        `gorm:"type:varchar(64)" json:"mall_product_code"`
	Description    string         `gorm:"type:text" json:"description"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy      uint64         `json:"created_by"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy      uint64         `json:"updated_by"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Product) TableName() string { return "product" }

type ProductWorkflowNode struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID uint64    `gorm:"not null;index;uniqueIndex:idx_product_node" json:"product_id"`
	NodeIndex int16     `gorm:"not null;uniqueIndex:idx_product_node" json:"node_index"`
	NodeCode  string    `gorm:"type:varchar(32);not null" json:"node_code"`
	NodeName  string    `gorm:"type:varchar(64);not null" json:"node_name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy uint64    `json:"created_by"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy uint64    `json:"updated_by"`
}

func (ProductWorkflowNode) TableName() string { return "product_workflow_node" }

type FinanceCategory struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentID     uint64         `gorm:"default:0;index" json:"parent_id"`
	Level        int16          `gorm:"not null" json:"level"`
	CategoryType int16          `gorm:"not null" json:"category_type"`
	CategoryCode string         `gorm:"type:varchar(64)" json:"category_code"`
	CategoryName string         `gorm:"type:varchar(128);not null" json:"category_name"`
	FinanceCode  string         `gorm:"type:varchar(64)" json:"finance_code"`
	Sort         int16          `gorm:"default:0" json:"sort"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy    uint64         `json:"created_by"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy    uint64         `json:"updated_by"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (FinanceCategory) TableName() string { return "finance_category" }

type ShopFinanceCategory struct {
	ID                 uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID           uint64         `gorm:"not null;index" json:"tenant_id"`
	PlatformCategoryID uint64         `gorm:"not null" json:"platform_category_id"`
	ParentID           uint64         `gorm:"default:0" json:"parent_id"`
	Level              int16          `gorm:"not null" json:"level"`
	CategoryType       int16          `gorm:"not null" json:"category_type"`
	CategoryCode       string         `gorm:"type:varchar(64)" json:"category_code"`
	CategoryName       string         `gorm:"type:varchar(128);not null" json:"category_name"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy          uint64         `json:"created_by"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy          uint64         `json:"updated_by"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (ShopFinanceCategory) TableName() string { return "shop_finance_category" }

type ShopProduct struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          uint64         `gorm:"not null;index" json:"tenant_id"`
	PlatformProductID uint64         `gorm:"not null;uniqueIndex:idx_tenant_product" json:"platform_product_id"`
	ProductCode       string         `gorm:"type:varchar(64)" json:"product_code"`
	ProductName       string         `gorm:"type:varchar(128);not null" json:"product_name"`
	PlatformPrice     float64        `gorm:"type:numeric(12,2)" json:"platform_price"`
	ShopPrice         float64        `gorm:"type:numeric(12,2);not null" json:"shop_price"`
	Status            int16          `gorm:"default:1" json:"status"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy         uint64         `json:"created_by"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy         uint64         `json:"updated_by"`
	DeletedAt         gorm.DeletedAt `gorm:"index;uniqueIndex:idx_tenant_product" json:"deleted_at"`
}

func (ShopProduct) TableName() string { return "shop_product" }

type ShopCustomer struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      uint64         `gorm:"not null;index" json:"tenant_id"`
	CustomerName  string         `gorm:"type:varchar(128);not null" json:"customer_name"`
	CustomerType  int16          `gorm:"not null" json:"customer_type"`
	ContactPerson string         `gorm:"type:varchar(64)" json:"contact_person"`
	ContactPhone  string         `gorm:"type:varchar(20)" json:"contact_phone"`
	ContactEmail  string         `gorm:"type:varchar(128)" json:"contact_email"`
	Address       string         `gorm:"type:varchar(255)" json:"address"`
	Remark        string         `gorm:"type:varchar(500)" json:"remark"`
	Status        int16          `gorm:"default:1" json:"status"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy     uint64         `gorm:"index" json:"created_by"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy     uint64         `json:"updated_by"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (ShopCustomer) TableName() string { return "shop_customer" }
