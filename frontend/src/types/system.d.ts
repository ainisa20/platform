/** 通用分页请求 */
export interface PageReq {
  page?: number
  page_size?: number
}

/** 通用分页响应 */
export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

// ==================== 角色 ====================

export interface RoleResp {
  id: number
  tenant_id: number
  role_name: string
  role_code: string
  data_scope: number
  sort: number
  status: number
  remark: string
  created_at: string
  created_by: number
  updated_at: string
  updated_by: number
  permissions?: PermissionResp[]
}

export interface RoleListReq extends PageReq {
  role_name?: string
  status?: number | null
}

export interface RoleCreateReq {
  role_name: string
  role_code: string
  remark?: string
  data_scope?: number
  sort?: number
  status?: number
}

export interface RoleUpdateReq {
  role_name?: string
  role_code?: string
  remark?: string
  data_scope?: number
  sort?: number
  status?: number
}

export interface RoleAssignPermsReq {
  permission_ids: number[]
}

// ==================== 部门 ====================

export interface DeptResp {
  id: number
  tenant_id: number
  parent_id: number
  ancestors: string
  dept_name: string
  sort: number
  leader: string
  phone: string
  status: number
  created_at: string
  created_by: number
  updated_at: string
  updated_by: number
  children?: DeptResp[]
}

export interface DeptListReq {
  status?: number | null
}

export interface DeptCreateReq {
  parent_id?: number
  dept_name: string
  leader?: string
  phone?: string
  sort?: number
  status?: number
}

export interface DeptUpdateReq {
  parent_id?: number
  dept_name?: string
  leader?: string
  phone?: string
  sort?: number
  status?: number
}

// ==================== 权限 ====================

export interface PermissionResp {
  id: number
  parent_id: number
  system_type: string
  name: string
  type: number
  path: string
  component: string
  perms_code: string
  icon: string
  sort: number
  visible: boolean
  status: number
  children?: PermissionResp[]
}

// ==================== 用户 ====================

export interface UserResp {
  id: number
  tenant_id: number
  dept_id: number | null
  username: string
  real_name: string
  phone: string
  email: string
  avatar: string
  status: number
  last_login_at: string | null
  last_login_ip: string
  created_at: string
  created_by: number
  updated_at: string
  updated_by: number
  roles: RoleResp[]
}

export interface UserCreateReq {
  username: string
  password: string
  real_name: string
  phone?: string
  email?: string
  dept_id?: number | null
  role_ids?: number[]
  status?: number
}

export interface UserUpdateReq {
  real_name?: string
  phone?: string
  email?: string
  dept_id?: number | null
  role_ids?: number[]
  status?: number
}

export interface UserListReq extends PageReq {
  username?: string
  real_name?: string
  phone?: string
  status?: number | null
}

export interface PasswordResetReq {
  new_password: string
}

export interface AssignRolesReq {
  role_ids: number[]
}

// ==================== 店铺 ====================

export interface ShopResp {
  id: number
  shop_code: string
  shop_name: string
  contact: string
  phone: string
  email: string
  address: string
  remark: string
  status: number
  admin_user_id: number | null
  admin_username?: string
  expires_at: string | null
  created_at: string
  created_by: number
  updated_at: string
  updated_by: number
}

export interface ShopListReq extends PageReq {
  shop_code?: string
  shop_name?: string
  status?: number | null
}

export interface ShopCreateReq {
  shop_code: string
  shop_name: string
  contact?: string
  phone?: string
  email?: string
  address?: string
  remark?: string
  admin_username: string
  admin_password: string
  admin_real_name?: string
}

export interface ShopUpdateReq {
  shop_name: string
  contact?: string
  phone?: string
  email?: string
  address?: string
  remark?: string
}

export interface ShopStatusReq {
  status: number
}

// ==================== 商品分类 ====================

export interface ProductCategoryResp {
  id: number
  category_code: string
  category_name: string
  sort: number
  status: number
  created_at: string
  created_by: number
  updated_at: string
  updated_by: number
}

export interface ProductCategoryListReq extends PageReq {
  category_name?: string
}

export interface ProductCategoryCreateReq {
  category_code?: string
  category_name: string
  sort?: number
  status?: number
}

export interface ProductCategoryUpdateReq {
  category_code?: string
  category_name?: string
  sort?: number
  status?: number
}

// ==================== 商品 ====================

export interface ProductResp {
  id: number
  product_code: string
  product_name: string
  category_id: number
  category_name?: string
  price: number
  sort: number
  status: number
  description: string
  created_at: string
  created_by: number
  updated_at: string
  updated_by: number
  workflow_nodes?: WorkflowNodeResp[]
}

export interface ProductListReq extends PageReq {
  product_name?: string
  category_id?: number
  status?: number | null
}

export interface ProductCreateReq {
  product_code: string
  product_name: string
  category_id?: number
  price: number
  sort?: number
  status?: number
  description?: string
}

export interface ProductUpdateReq {
  product_name?: string
  category_id?: number
  price?: number
  sort?: number
  status?: number
  description?: string
}

// ==================== 工作流 ====================

export interface WorkflowNodeResp {
  id: number
  product_id: number
  node_index: number
  node_code: string
  node_name: string
}

export interface WorkflowNodeReq {
  node_index: number
  node_code: string
  node_name: string
}

export interface WorkflowSaveReq {
  nodes: WorkflowNodeReq[]
}

// ==================== 收支分类 ====================

export interface FinanceCategoryResp {
  id: number
  parent_id: number
  level: number
  category_type: number
  category_code: string
  category_name: string
  finance_code: string
  sort: number
  created_at: string
  created_by: number
  updated_at: string
  updated_by: number
  children?: FinanceCategoryResp[]
}

export interface FinanceCategoryListReq {
  category_type?: number | null
  category_name?: string
}

export interface FinanceCategoryCreateReq {
  parent_id?: number
  category_type: number
  category_code?: string
  category_name: string
  finance_code?: string
  sort?: number
}

export interface FinanceCategoryUpdateReq {
  parent_id?: number
  category_type?: number
  category_code?: string
  category_name: string
  finance_code?: string
  sort?: number
}

// ==================== 店铺收支分类 ====================

export interface ShopFinCategoryResp {
  id: number
  platform_category_id: number
  parent_id: number
  level: number
  category_type: number
  category_code: string
  category_name: string
  created_at: string
  created_by: number
  children?: ShopFinCategoryResp[]
}

export interface ShopFinCategoryAvailableResp {
  id: number
  parent_id: number
  level: number
  category_type: number
  category_code: string
  category_name: string
  children?: ShopFinCategoryAvailableResp[]
}

export interface ShopFinCategorySyncReq {
  platform_category_ids: number[]
}

// ==================== 店铺选品 ====================

export interface ShopProductResp {
  id: number
  platform_product_id: number
  product_code: string
  product_name: string
  platform_price: number
  shop_price: number
  status: number
  created_at: string
  created_by: number
  updated_at: string
  updated_by: number
}

export interface ShopProductListReq extends PageReq {
  product_name?: string
  status?: number | null
}

export interface ShopProductSelectReq {
  platform_product_ids: number[]
}

export interface ShopProductPriceReq {
  shop_price: number
}

export interface ShopProductStatusReq {
  status: number
}

export interface ShopPlatformProductResp {
  id: number
  product_code: string
  product_name: string
  price: number
  category_id: number | null
  description: string
  category_name?: string
}

// ==================== 店铺客户 ====================

export interface ShopCustomerResp {
  id: number
  customer_name: string
  customer_type: number
  contact_person: string
  contact_phone: string
  contact_email: string
  address: string
  remark: string
  status: number
  created_at: string
  created_by: number
  created_by_name?: string
  updated_at: string
  updated_by: number
}

export interface ShopCustomerListReq extends PageReq {
  customer_name?: string
  contact_person?: string
  customer_type?: number | null
  status?: number | null
}

export interface ShopCustomerCreateReq {
  customer_name: string
  customer_type: number
  contact_person?: string
  contact_phone?: string
  contact_email?: string
  address?: string
  remark?: string
  status?: number
}

export interface ShopCustomerUpdateReq {
  customer_name: string
  customer_type: number
  contact_person?: string
  contact_phone?: string
  contact_email?: string
  address?: string
  remark?: string
  status?: number
}

export interface ShopCustomerOrderResp {
  id: number
  order_no: string
  total_amount: number
  status: number
  created_at: string
  created_by: number
}
