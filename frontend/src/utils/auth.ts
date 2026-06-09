const PLATFORM_TOKEN_KEY = 'platform_token'
const SHOP_TOKEN_KEY = 'shop_token'
const PLATFORM_REFRESH_KEY = 'platform_refresh_token'
const SHOP_REFRESH_KEY = 'shop_refresh_token'

export type SystemType = 'platform' | 'shop'

export function getSystemType(): SystemType {
  if (typeof window === 'undefined') return 'platform'
  const path = window.location.pathname
  if (path.startsWith('/shop')) return 'shop'
  return 'platform'
}

export function getToken(): string | null {
  const system = getSystemType()
  return localStorage.getItem(system === 'platform' ? PLATFORM_TOKEN_KEY : SHOP_TOKEN_KEY)
}

export function setToken(token: string): void {
  const system = getSystemType()
  localStorage.setItem(system === 'platform' ? PLATFORM_TOKEN_KEY : SHOP_TOKEN_KEY, token)
}

export function removeToken(): void {
  const system = getSystemType()
  localStorage.removeItem(system === 'platform' ? PLATFORM_TOKEN_KEY : SHOP_TOKEN_KEY)
}

export function getRefreshToken(): string | null {
  const system = getSystemType()
  return localStorage.getItem(system === 'platform' ? PLATFORM_REFRESH_KEY : SHOP_REFRESH_KEY)
}

export function setRefreshToken(token: string): void {
  const system = getSystemType()
  localStorage.setItem(system === 'platform' ? PLATFORM_REFRESH_KEY : SHOP_REFRESH_KEY, token)
}

export function removeRefreshToken(): void {
  const system = getSystemType()
  localStorage.removeItem(system === 'platform' ? PLATFORM_REFRESH_KEY : SHOP_REFRESH_KEY)
}

export function clearAuth(): void {
  removeToken()
  removeRefreshToken()
}
