export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface LoginParams {
  username: string
  password: string
}

export interface ShopLoginParams {
  username: string
  password: string
  shop_code: string
}

export interface LoginResult {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface UserInfo {
  id: number
  tenant_id: number
  username: string
  real_name: string
  roles: string[]
  permissions: string[]
}

export interface RefreshParams {
  refresh_token: string
}
