package dto

import "time"

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
	Address       string `json:"address"`
	Remark        string `json:"remark"`
	AdminUsername string `json:"admin_username" binding:"required,min=3"`
	AdminPassword string `json:"admin_password" binding:"required,min=6"`
	AdminRealName string `json:"admin_real_name"`
}

// ShopUpdateReq update shop request.
type ShopUpdateReq struct {
	ShopName string `json:"shop_name" binding:"required"`
	Contact  string `json:"contact"`
	Phone    string `json:"phone"`
	Email    string `json:"email" binding:"omitempty,email"`
	Address  string `json:"address"`
	Remark   string `json:"remark"`
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
	Status   *int16 `form:"status"`
}

// ShopResp shop response.
type ShopResp struct {
	ID           uint64     `json:"id"`
	ShopCode     string     `json:"shop_code"`
	ShopName     string     `json:"shop_name"`
	Contact      string     `json:"contact"`
	Phone        string     `json:"phone"`
	Email        string     `json:"email"`
	Address      string     `json:"address"`
	Remark       string     `json:"remark"`
	Status       int16      `json:"status"`
	AdminUserID  *uint64    `json:"admin_user_id"`
	AdminUsername string    `json:"admin_username"`
	ExpiresAt    *time.Time `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	CreatedBy    uint64     `json:"created_by"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    uint64     `json:"updated_by"`
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
	ProductCode     string  `json:"product_code" binding:"required"`
	ProductName     string  `json:"product_name" binding:"required"`
	CategoryID      *uint64 `json:"category_id" binding:"required"`
	Price           float64 `json:"price" binding:"required"`
	Sort            int16   `json:"sort"`
	Status          int16   `json:"status"`
	MallProductCode string  `json:"mall_product_code"`
	Description     string  `json:"description"`
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
	Price           float64           `json:"price"`
	Sort            int16             `json:"sort"`
	Status          int16             `json:"status"`
	MallProductCode string            `json:"mall_product_code"`
	Description     string            `json:"description"`
	CreatedAt       time.Time         `json:"created_at"`
	CreatedBy       uint64            `json:"created_by"`
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
	Page          int    `form:"page"`
	PageSize      int    `form:"page_size"`
	CustomerName  string `form:"customer_name"`
	ContactPerson string `form:"contact_person"`
	Status        *int16 `form:"status"`
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
