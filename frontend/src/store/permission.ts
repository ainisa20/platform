import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getSystemType } from '@/utils/auth'
import * as platformAuth from '@/api/platform/auth'
import * as shopAuth from '@/api/shop/auth'

export const usePermissionStore = defineStore('permission', () => {
  const permissionCodes = ref<string[]>([])
  const menuRoutes = ref<any[]>([])

  async function fetchPermissions() {
    const authApi = getSystemType() === 'platform' ? platformAuth : shopAuth
    const res = await authApi.getPermissions()
    permissionCodes.value = res.data.data || []
    return permissionCodes.value
  }

  function setPermissionCodes(codes: string[]) {
    permissionCodes.value = codes
  }

  function generateRoutes(routes: any[]) {
    const filtered = filterRoutes(routes)
    menuRoutes.value = filtered
    return filtered
  }

  function filterRoutes(routes: any[]): any[] {
    return routes.filter((route) => {
      if (route.children) {
        route.children = filterRoutes(route.children)
      }
      if (route.meta?.permissionCode) {
        return permissionCodes.value.includes(route.meta.permissionCode)
      }
      return true
    })
  }

  function resetState() {
    permissionCodes.value = []
    menuRoutes.value = []
  }

  return {
    permissionCodes,
    menuRoutes,
    fetchPermissions,
    setPermissionCodes,
    generateRoutes,
    resetState,
  }
})
