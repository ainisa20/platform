import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { UserInfo } from '@/types/api'
import { getToken, setToken, clearAuth, setRefreshToken, getSystemType } from '@/utils/auth'
import * as platformAuth from '@/api/platform/auth'
import * as shopAuth from '@/api/shop/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref<string | null>(getToken())
  const refreshTokenVal = ref<string | null>(null)
  const userInfo = ref<UserInfo | null>(null)

  function isPlatform() {
    return getSystemType() === 'platform'
  }

  async function login(loginData: { username: string; password: string; shop_code?: string }) {
    const authApi = isPlatform() ? platformAuth : shopAuth
    const res = await authApi.login(loginData as any)
    const data = res.data.data
    token.value = data.access_token
    refreshTokenVal.value = data.refresh_token
    setToken(data.access_token)
    setRefreshToken(data.refresh_token)
  }

  async function getUserInfoAction() {
    const authApi = isPlatform() ? platformAuth : shopAuth
    const res = await authApi.getUserInfo()
    userInfo.value = res.data.data
    return userInfo.value
  }

  async function logout() {
    try {
      const authApi = isPlatform() ? platformAuth : shopAuth
      await authApi.logout()
    } finally {
      resetState()
    }
  }

  function resetState() {
    token.value = null
    refreshTokenVal.value = null
    userInfo.value = null
    clearAuth()
  }

  return {
    token,
    refreshToken: refreshTokenVal,
    userInfo,
    login,
    getUserInfo: getUserInfoAction,
    logout,
    resetState,
  }
})
