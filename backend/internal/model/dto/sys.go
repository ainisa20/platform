package dto

import (
	"time"

	"gorm.io/datatypes"
)

// ========== Common ==========

// PageReq generic pagination request.
type PageReq struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// ========== User ==========

// UserCreateReq create user request.
type UserCreateReq struct {
	Username string   `json:"username" binding:"required"`
	Password string   `json:"password" binding:"required,min=6"`
	RealName string   `json:"real_name" binding:"required"`
	Phone    string   `json:"phone"`
	Email    string   `json:"email"`
	DeptID   *uint64  `json:"dept_id" binding:"required"`
	RoleIDs  []uint64 `json:"role_ids" binding:"required,min=1"`
	Status   int16    `json:"status"`
}

// UserUpdateReq update user request (username/password not changeable here).
type UserUpdateReq struct {
	RealName string   `json:"real_name"`
	Phone    string   `json:"phone"`
	Email    string   `json:"email"`
	DeptID   *uint64  `json:"dept_id"`
	RoleIDs  []uint64 `json:"role_ids"`
	Status   int16    `json:"status"`
}

// UserListReq user list with pagination and search.
type UserListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Username string `form:"username"`
	RealName string `form:"real_name"`
	Phone    string `form:"phone"`
	Status   *int16 `form:"status"`
}

// UserResp user response with nested roles.
type UserResp struct {
	ID          uint64     `json:"id"`
	TenantID    uint64     `json:"tenant_id"`
	DeptID      *uint64    `json:"dept_id"`
	Username    string     `json:"username"`
	RealName    string     `json:"real_name"`
	Phone       string     `json:"phone"`
	Email       string     `json:"email"`
	Avatar      string     `json:"avatar"`
	Status      int16      `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LastLoginIP string     `json:"last_login_ip"`
	CreatedAt   time.Time  `json:"created_at"`
	CreatedBy   uint64     `json:"created_by"`
	UpdatedAt   time.Time  `json:"updated_at"`
	UpdatedBy   uint64     `json:"updated_by"`
	Roles       []RoleResp `json:"roles"`
}

// PasswordResetReq reset password request.
type PasswordResetReq struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ========== Shop ==========

// ShopCreateReq create shop request. AdminUsername/AdminPassword initialize
// the shop's admin user; AdminRealName defaults to "店长" when omitted.
type ShopCreateReq struct {
	ShopCode      string `json:"shop_code" binding:"required"`
	ShopName      string `json:"shop_name" binding:"required"`
	Contact       string `json:"contact"`
	Phone         string `json:"phone"`
	Email         string `json:"email" binding:"omitempty,email"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	DetailAddress string `json:"detail_address"`
	Address       string `json:"address"`
	Remark        string `json:"remark"`
	AdminUsername string `json:"admin_username" binding:"required,min=3"`
	AdminPassword string `json:"admin_password" binding:"required,min=6"`
	AdminRealName string `json:"admin_real_name"`
}

// ShopUpdateReq update shop request.
type ShopUpdateReq struct {
	ShopName      string `json:"shop_name" binding:"required"`
	Contact       string `json:"contact"`
	Phone         string `json:"phone"`
	Email         string `json:"email" binding:"omitempty,email"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	DetailAddress string `json:"detail_address"`
	Address       string `json:"address"`
	Remark        string `json:"remark"`
}

// ShopStatusReq enable/disable shop.
type ShopStatusReq struct {
	Status int16 `json:"status" binding:"required,oneof=1 2"`
}

// ShopListReq paginated list with search.
type ShopListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	ShopCode string `form:"shop_code"`
	ShopName string `form:"shop_name"`
	Province string `form:"province"`
	City     string `form:"city"`
	District string `form:"district"`
	Status   *int16 `form:"status"`
}

// ShopResp shop response.
type ShopResp struct {
	ID            uint64     `json:"id"`
	ShopCode      string     `json:"shop_code"`
	ShopName      string     `json:"shop_name"`
	Contact       string     `json:"contact"`
	Phone         string     `json:"phone"`
	Email         string     `json:"email"`
	Province      string     `json:"province"`
	City          string     `json:"city"`
	District      string     `json:"district"`
	DetailAddress string     `json:"detail_address"`
	Address       string     `json:"address"`
	Remark        string     `json:"remark"`
	Status        int16      `json:"status"`
	AdminUserID   *uint64    `json:"admin_user_id"`
	AdminUsername  string    `json:"admin_username"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	CreatedBy     uint64     `json:"created_by"`
	UpdatedAt     time.Time  `json:"updated_at"`
	UpdatedBy     uint64     `json:"updated_by"`
}

// ========== Role ==========

// RoleCreateReq create role request.
type RoleCreateReq struct {
	RoleName  string `json:"role_name" binding:"required"`
	RoleCode  string `json:"role_code" binding:"required"`
	Remark    string `json:"remark"`
	DataScope int16  `json:"data_scope"`
	Sort      int16  `json:"sort"`
	Status    int16  `json:"status"`
}

// RoleUpdateReq update role request (same fields as create).
type RoleUpdateReq struct {
	RoleName  string `json:"role_name" binding:"required"`
	RoleCode  string `json:"role_code" binding:"required"`
	Remark    string `json:"remark"`
	DataScope int16  `json:"data_scope"`
	Sort      int16  `json:"sort"`
	Status    int16  `json:"status"`
}

// RoleListReq role list with pagination.
type RoleListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	RoleName string `form:"role_name"`
	Status   *int16 `form:"status"`
}

// RoleResp role response with nested permissions.
type RoleResp struct {
	ID          uint64           `json:"id"`
	TenantID    uint64           `json:"tenant_id"`
	RoleName    string           `json:"role_name"`
	RoleCode    string           `json:"role_code"`
	DataScope   int16            `json:"data_scope"`
	Sort        int16            `json:"sort"`
	Status      int16            `json:"status"`
	Remark      string           `json:"remark"`
	CreatedAt   time.Time        `json:"created_at"`
	CreatedBy   uint64           `json:"created_by"`
	UpdatedAt   time.Time        `json:"updated_at"`
	UpdatedBy   uint64           `json:"updated_by"`
	Permissions []PermissionResp `json:"permissions"`
}

// RoleAssignPermsReq assign permissions to role.
type RoleAssignPermsReq struct {
	PermissionIDs []uint64 `json:"permission_ids"`
}

// ========== Dept ==========

// DeptCreateReq create department request.
type DeptCreateReq struct {
	ParentID uint64 `json:"parent_id"`
	DeptName string `json:"dept_name" binding:"required"`
	Leader   string `json:"leader"`
	Phone    string `json:"phone"`
	Sort     int16  `json:"sort"`
	Status   int16  `json:"status"`
}

// DeptUpdateReq update department request.
type DeptUpdateReq struct {
	ParentID uint64 `json:"parent_id"`
	DeptName string `json:"dept_name"`
	Leader   string `json:"leader"`
	Phone    string `json:"phone"`
	Sort     int16  `json:"sort"`
	Status   int16  `json:"status"`
}

// DeptResp department tree response.
type DeptResp struct {
	ID        uint64     `json:"id"`
	TenantID  uint64     `json:"tenant_id"`
	ParentID  uint64     `json:"parent_id"`
	Ancestors string     `json:"ancestors"`
	DeptName  string     `json:"dept_name"`
	Sort      int16      `json:"sort"`
	Leader    string     `json:"leader"`
	Phone     string     `json:"phone"`
	Status    int16      `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	CreatedBy uint64     `json:"created_by"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdatedBy uint64     `json:"updated_by"`
	Children  []DeptResp `json:"children"`
}

// ========== Permission ==========

// PermissionResp permission tree response (read-only).
type PermissionResp struct {
	ID         uint64           `json:"id"`
	ParentID   uint64           `json:"parent_id"`
	SystemType string           `json:"system_type"`
	Name       string           `json:"name"`
	Type       int16            `json:"type"`
	Path       string           `json:"path"`
	Component  string           `json:"component"`
	PermsCode  string           `json:"perms_code"`
	Icon       string           `json:"icon"`
	Sort       int16            `json:"sort"`
	Visible    bool             `json:"visible"`
	Status     int16            `json:"status"`
	Children   []PermissionResp `json:"children"`
}

// ========== Product Category ==========

type CategoryCreateReq struct {
	CategoryCode string `json:"category_code"`
	CategoryName string `json:"category_name" binding:"required"`
	Sort         int16  `json:"sort"`
	Status       int16  `json:"status"`
}

type CategoryUpdateReq struct {
	CategoryCode string `json:"category_code"`
	CategoryName string `json:"category_name" binding:"required"`
	Sort         int16  `json:"sort"`
	Status       int16  `json:"status"`
}

type CategoryListReq struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	CategoryName string `form:"category_name"`
	Status       *int16 `form:"status"`
}

type CategoryResp struct {
	ID           uint64    `json:"id"`
	CategoryCode string    `json:"category_code"`
	CategoryName string    `json:"category_name"`
	Sort         int16     `json:"sort"`
	Status       int16     `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    uint64    `json:"created_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    uint64    `json:"updated_by"`
}

// ========== Product ==========

type ProductCreateReq struct {
	ProductCode     string           `json:"product_code" binding:"required"`
	ProductName     string           `json:"product_name" binding:"required"`
	CategoryID      *uint64          `json:"category_id" binding:"required"`
	Price           float64          `json:"price" binding:"required"`
	Sort            int16            `json:"sort"`
	Status          int16            `json:"status"`
	MallProductCode string           `json:"mall_product_code"`
	Description     string           `json:"description"`
	WorkflowNodes   []WorkflowNodeReq `json:"workflow_nodes" binding:"required,dive"`
}

type ProductUpdateReq struct {
	ProductCode     string  `json:"product_code"`
	ProductName     string  `json:"product_name" binding:"required"`
	CategoryID      *uint64 `json:"category_id"`
	Price           float64 `json:"price"`
	Sort            int16   `json:"sort"`
	Status          int16   `json:"status"`
	MallProductCode string  `json:"mall_product_code"`
	Description     string  `json:"description"`
}

type ProductListReq struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	ProductName string `form:"product_name"`
	CategoryID *uint64 `form:"category_id"`
	Status     *int16 `form:"status"`
}

type ProductStatusReq struct {
	Status int16 `json:"status" binding:"required,oneof=1 2"`
}

type ProductResp struct {
	ID              uint64            `json:"id"`
	ProductCode     string            `json:"product_code"`
	ProductName     string            `json:"product_name"`
	CategoryID      *uint64           `json:"category_id"`
	CategoryName    string            `json:"category_name"`
	Price           float64           `json:"price"`
	Sort            int16             `json:"sort"`
	Status          int16             `json:"status"`
	MallProductCode string            `json:"mall_product_code"`
	Description     string            `json:"description"`
	CreatedAt       time.Time         `json:"created_at"`
	CreatedBy       uint64            `json:"created_by"`
	CreatedByName   string            `json:"created_by_name"`
	UpdatedAt       time.Time         `json:"updated_at"`
	UpdatedBy       uint64            `json:"updated_by"`
	WorkflowNodes   []WorkflowNodeResp `json:"workflow_nodes,omitempty"`
}

// ========== Workflow ==========

type WorkflowNodeReq struct {
	NodeIndex int16  `json:"node_index" binding:"required"`
	NodeCode  string `json:"node_code" binding:"required"`
	NodeName  string `json:"node_name" binding:"required"`
}

type WorkflowSaveReq struct {
	Nodes []WorkflowNodeReq `json:"nodes" binding:"required,dive"`
}

type WorkflowNodeResp struct {
	ID        uint64    `json:"id"`
	ProductID uint64    `json:"product_id"`
	NodeIndex int16     `json:"node_index"`
	NodeCode  string    `json:"node_code"`
	NodeName  string    `json:"node_name"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy uint64    `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy uint64    `json:"updated_by"`
}

// ========== Finance Category ==========

type FinanceCategoryCreateReq struct {
	ParentID     uint64 `json:"parent_id"`
	CategoryType int16  `json:"category_type" binding:"required,oneof=1 2"`
	CategoryCode string `json:"category_code"`
	CategoryName string `json:"category_name" binding:"required"`
	FinanceCode  string `json:"finance_code"`
	Sort         int16  `json:"sort"`
}

type FinanceCategoryUpdateReq struct {
	ParentID     uint64 `json:"parent_id"`
	CategoryType int16  `json:"category_type"`
	CategoryCode string `json:"category_code"`
	CategoryName string `json:"category_name" binding:"required"`
	FinanceCode  string `json:"finance_code"`
	Sort         int16  `json:"sort"`
}

type FinanceCategoryListReq struct {
	CategoryType *int16 `form:"category_type"`
	CategoryName string `form:"category_name"`
}

type FinanceCategoryResp struct {
	ID           uint64                `json:"id"`
	ParentID     uint64                `json:"parent_id"`
	Level        int16                 `json:"level"`
	CategoryType int16                 `json:"category_type"`
	CategoryCode string                `json:"category_code"`
	CategoryName string                `json:"category_name"`
	FinanceCode  string                `json:"finance_code"`
	Sort         int16                 `json:"sort"`
	CreatedAt    time.Time             `json:"created_at"`
	CreatedBy    uint64                `json:"created_by"`
	UpdatedAt    time.Time             `json:"updated_at"`
	UpdatedBy    uint64                `json:"updated_by"`
	Children     []FinanceCategoryResp `json:"children"`
}

// ========== Shop Finance Category ==========

type ShopFinCategorySyncReq struct {
	PlatformCategoryIDs []uint64 `json:"platform_category_ids" binding:"required,min=1"`
}

type ShopFinCategoryListReq struct {
	CategoryType *int16 `form:"category_type"`
}

type ShopFinCategoryResp struct {
	ID                 uint64                `json:"id"`
	PlatformCategoryID uint64                `json:"platform_category_id"`
	ParentID           uint64                `json:"parent_id"`
	Level              int16                 `json:"level"`
	CategoryType       int16                 `json:"category_type"`
	CategoryCode       string                `json:"category_code"`
	CategoryName       string                `json:"category_name"`
	CreatedAt          time.Time             `json:"created_at"`
	CreatedBy          uint64                `json:"created_by"`
	Children           []ShopFinCategoryResp `json:"children"`
}

type ShopFinCategoryAvailableResp struct {
	ID           uint64                          `json:"id"`
	ParentID     uint64                          `json:"parent_id"`
	Level        int16                           `json:"level"`
	CategoryType int16                           `json:"category_type"`
	CategoryCode string                          `json:"category_code"`
	CategoryName string                          `json:"category_name"`
	Children     []ShopFinCategoryAvailableResp  `json:"children"`
}

// ========== Shop Product ==========

type ShopProductSelectReq struct {
	PlatformProductIDs []uint64 `json:"platform_product_ids" binding:"required,min=1"`
}

type ShopProductPriceReq struct {
	ShopPrice float64 `json:"shop_price" binding:"required"`
}

type ShopProductStatusReq struct {
	Status int16 `json:"status" binding:"required,oneof=1 2"`
}

type ShopProductListReq struct {
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
	ProductName string `form:"product_name"`
	Status      *int16 `form:"status"`
}

type ShopProductResp struct {
	ID                uint64    `json:"id"`
	PlatformProductID uint64    `json:"platform_product_id"`
	ProductCode       string    `json:"product_code"`
	ProductName       string    `json:"product_name"`
	PlatformPrice     float64   `json:"platform_price"`
	ShopPrice         float64   `json:"shop_price"`
	Status            int16     `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         uint64    `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         uint64    `json:"updated_by"`
}

type ShopPlatformProductResp struct {
	ID           uint64  `json:"id"`
	ProductCode  string  `json:"product_code"`
	ProductName  string  `json:"product_name"`
	Price        float64 `json:"price"`
	CategoryID   *uint64 `json:"category_id"`
	Description  string  `json:"description"`
	CategoryName string  `json:"category_name,omitempty"`
}

// ========== Shop Customer ==========

type ShopCustomerCreateReq struct {
	CustomerName  string `json:"customer_name" binding:"required"`
	CustomerType  int16  `json:"customer_type" binding:"required,oneof=1 2"`
	ContactPerson string `json:"contact_person"`
	ContactPhone  string `json:"contact_phone"`
	ContactEmail  string `json:"contact_email" binding:"omitempty,email"`
	Address       string `json:"address"`
	Remark        string `json:"remark"`
	Status        int16  `json:"status"`
}

type ShopCustomerUpdateReq struct {
	CustomerName  string `json:"customer_name" binding:"required"`
	CustomerType  int16  `json:"customer_type" binding:"required,oneof=1 2"`
	ContactPerson string `json:"contact_person"`
	ContactPhone  string `json:"contact_phone"`
	ContactEmail  string `json:"contact_email" binding:"omitempty,email"`
	Address       string `json:"address"`
	Remark        string `json:"remark"`
	Status        int16  `json:"status"`
}

type ShopCustomerListReq struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	CustomerName string `form:"customer_name"`
	ContactPerson string `form:"contact_person"`
	CustomerType *int16 `form:"customer_type"`
	Status       *int16 `form:"status"`
}

type ShopCustomerResp struct {
	ID            uint64    `json:"id"`
	CustomerName  string    `json:"customer_name"`
	CustomerType  int16     `json:"customer_type"`
	ContactPerson string    `json:"contact_person"`
	ContactPhone  string    `json:"contact_phone"`
	ContactEmail  string    `json:"contact_email"`
	Address       string    `json:"address"`
	Remark        string    `json:"remark"`
	Status        int16     `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     uint64    `json:"created_by"`
	CreatedByName string    `json:"created_by_name"`
	UpdatedAt     time.Time `json:"updated_at"`
	UpdatedBy     uint64    `json:"updated_by"`
}

type ShopCustomerOrderResp struct {
	ID          uint64    `json:"id"`
	OrderNo     string    `json:"order_no"`
	TotalAmount float64   `json:"total_amount"`
	Status      int16     `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   uint64    `json:"created_by"`
}

// ========== Shop Order ==========

type OrderItemReq struct {
	ShopProductID uint64 `json:"shop_product_id" binding:"required"`
	Quantity      int16  `json:"quantity" binding:"required,min=1"`
}

type OrderCreateReq struct {
	CustomerID uint64         `json:"customer_id" binding:"required"`
	Remark     string         `json:"remark"`
	Items      []OrderItemReq `json:"items" binding:"required,min=1,dive"`
}

type OrderListReq struct {
	Page             int     `form:"page"`
	PageSize         int     `form:"page_size"`
	OrderNo          string  `form:"order_no"`
	CustomerID       *uint64 `form:"customer_id"`
	OrderStatus      *int16  `form:"order_status"`
	ExcludeCancelled bool    `form:"exclude_cancelled"`
}

type OrderItemResp struct {
	ID               uint64    `json:"id"`
	OrderGroupID     uint64    `json:"order_group_id"`
	ShopProductID    uint64    `json:"shop_product_id"`
	ProductName      string    `json:"product_name"`
	Quantity         int16     `json:"quantity"`
	UnitPrice        float64   `json:"unit_price"`
	TotalPrice       float64   `json:"total_price"`
	CurrentNodeIndex int16     `json:"current_node_index"`
	CurrentNodeName  string    `json:"current_node_name"`
	NextNodeName     string    `json:"next_node_name"`
	ItemStatus       int16     `json:"item_status"`
	CreatedAt        time.Time `json:"created_at"`
}

type OrderResp struct {
	ID            uint64          `json:"id"`
	OrderNo       string          `json:"order_no"`
	CustomerID    uint64          `json:"customer_id"`
	CustomerName  string          `json:"customer_name"`
	TotalAmount   float64         `json:"total_amount"`
	OrderStatus   int16           `json:"order_status"`
	Remark        string          `json:"remark"`
	ItemCount     int             `json:"item_count"`
	Items         []OrderItemResp `json:"items,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	CreatedBy     uint64          `json:"created_by"`
	CreatedByName string          `json:"created_by_name"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type OrderWorkflowAdvanceReq struct {
	Notes string `json:"notes"`
}

type OrderWorkflowLogResp struct {
	ID           uint64    `json:"id"`
	OrderItemID  uint64    `json:"order_item_id"`
	NodeIndex    int16     `json:"node_index"`
	NodeCode     string    `json:"node_code"`
	NodeName     string    `json:"node_name"`
	Notes        string    `json:"notes"`
	OperatorID   uint64    `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	OperatedAt   time.Time `json:"operated_at"`
}

type OrderAttachmentResp struct {
	ID            uint64    `json:"id"`
	FileName      string    `json:"file_name"`
	FileSize      int64     `json:"file_size"`
	FileType      string    `json:"file_type"`
	WorkflowLogID *uint64   `json:"workflow_log_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// ========== Shop Finance Account ==========

type ShopFinAccountCreateReq struct {
	AccountName    string         `json:"account_name" binding:"required"`
	AccountType    int16          `json:"account_type" binding:"required,oneof=1 2"`
	AccountNo      string         `json:"account_no"`
	InitialBalance float64        `json:"initial_balance"`
	Config         datatypes.JSON `json:"config"`
	Status         int16          `json:"status"`
}

type ShopFinAccountUpdateReq struct {
	AccountName string         `json:"account_name" binding:"required"`
	AccountNo   string         `json:"account_no"`
	Config      datatypes.JSON `json:"config"`
	Status      int16          `json:"status"`
}

type ShopFinAccountListReq struct {
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
	AccountName string `form:"account_name"`
	AccountType *int16 `form:"account_type"`
	Status      *int16 `form:"status"`
}

type ShopFinAccountResp struct {
	ID             uint64         `json:"id"`
	AccountName    string         `json:"account_name"`
	AccountType    int16          `json:"account_type"`
	AccountNo      string         `json:"account_no"`
	InitialBalance float64        `json:"initial_balance"`
	Balance        float64        `json:"balance"`
	Config         datatypes.JSON `json:"config"`
	Status         int16          `json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	CreatedBy      uint64         `json:"created_by"`
	CreatedByName  string         `json:"created_by_name"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type FinanceRecordCreateReq struct {
	AccountID    uint64  `json:"account_id" binding:"required"`
	CategoryID   uint64  `json:"category_id" binding:"required"`
	RecordType   int16   `json:"record_type" binding:"required,oneof=1 2"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	OrderGroupID *uint64 `json:"order_group_id"`
	RecordDate   string  `json:"record_date" binding:"required"`
	Remark       string  `json:"remark"`
}

type FinanceRecordUpdateReq struct {
	AccountID    uint64  `json:"account_id" binding:"required"`
	CategoryID   uint64  `json:"category_id" binding:"required"`
	RecordType   int16   `json:"record_type" binding:"required,oneof=1 2"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	OrderGroupID *uint64 `json:"order_group_id"`
	RecordDate   string  `json:"record_date" binding:"required"`
	Remark       string  `json:"remark"`
}

type FinanceRecordListReq struct {
	Page         int     `form:"page"`
	PageSize     int     `form:"page_size"`
	RecordNo     string  `form:"record_no"`
	AccountID    *uint64 `form:"account_id"`
	AccountType  *int16  `form:"account_type"`
	CategoryID   *uint64 `form:"category_id"`
	CategoryL1   *string `form:"category_l1"`
	CategoryL2   *string `form:"category_l2"`
	CategoryL3   *string `form:"category_l3"`
	RecordType   *int16  `form:"record_type"`
	ReviewStatus *int16  `form:"review_status"`
	RecordDateStart string `form:"record_date_start"`
	RecordDateEnd   string `form:"record_date_end"`
	CreatedBy    *uint64 `form:"created_by"`
}

type FinanceReviewReq struct {
	Action       string   `json:"action" binding:"required,oneof=approve reject"`
	ActualAmount *float64 `json:"actual_amount"`
	Notes        string   `json:"notes"`
}

type FinanceRecordResp struct {
	ID                    uint64     `json:"id"`
	RecordNo              string     `json:"record_no"`
	AccountID             uint64     `json:"account_id"`
	AccountName           string     `json:"account_name"`
	AccountType           int16      `json:"account_type"`
	AccountInitialBalance float64    `json:"account_initial_balance"`
	CategoryID            uint64     `json:"category_id"`
	CategoryName          string     `json:"category_name"`
	CategoryPath          string     `json:"category_path"`
	CategoryL1            string     `json:"category_l1"`
	CategoryL2            string     `json:"category_l2"`
	CategoryL3            string     `json:"category_l3"`
	RecordType            int16      `json:"record_type"`
	Amount                float64    `json:"amount"`
	ActualAmount          float64    `json:"actual_amount"`
	OrderGroupID          *uint64    `json:"order_group_id"`
	ReviewStatus          int16      `json:"review_status"`
	ReviewBy              uint64     `json:"review_by"`
	ReviewByName          string     `json:"review_by_name"`
	ReviewAt              *time.Time `json:"review_at"`
	ReviewNotes           string     `json:"review_notes"`
	RecordDate            string     `json:"record_date"`
	Remark                string     `json:"remark"`
	CreatedAt             time.Time  `json:"created_at"`
	CreatedBy             uint64     `json:"created_by"`
	CreatedByName         string     `json:"created_by_name"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type FinanceAttachmentReq struct {
	FileName string `json:"file_name" binding:"required"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
}

type FinanceAttachmentResp struct {
	ID        uint64    `json:"id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	FileType  string    `json:"file_type"`
	CreatedAt time.Time `json:"created_at"`
}

// ==================== Finance Report ====================

type FinanceReportReq struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

type FinanceSummaryResp struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	NetProfit    float64 `json:"net_profit"`
}

type FinanceTrendItem struct {
	Month   string  `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

type ProfitLossCategory struct {
	Name     string               `json:"name"`
	Type     string               `json:"type"`
	Subtotal float64              `json:"subtotal"`
	Children []*ProfitLossCategory `json:"children,omitempty" gorm:"-"`
}

type FinanceProfitLossResp struct {
	Categories []*ProfitLossCategory `json:"categories"`
}

// ==================== Platform Finance Report ====================

type PlatformFinanceReportReq struct {
	ShopID    *uint64 `form:"shop_id"`
	StartDate string  `form:"start_date"`
	EndDate   string  `form:"end_date"`
}

type FinanceReportShopSummary struct {
	ShopID     uint64  `json:"shop_id"`
	ShopName   string  `json:"shop_name"`
	Income     float64 `json:"income"`
	Expense    float64 `json:"expense"`
	NetProfit  float64 `json:"net_profit"`
}
